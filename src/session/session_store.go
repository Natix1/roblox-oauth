package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	appTypes "github.com/natix1/roblox-oauth/src/app_types"
	"github.com/natix1/roblox-oauth/src/server"
)

const (
	SessionTokenCookieName = "session_cookie"
	RedisNamespace         = "robloxoauth:"
	StoreExpiry            = time.Hour * 24 * 90
	AccessTokenExpiry      = time.Second * 15
)

type RefreshWithToken int

const (
	RefreshWithRefreshToken RefreshWithToken = iota
	RefreshWithAccessToken
)

var (
	ErrNeedsReAuth = errors.New("needs re-authentication")
	redisContext   = context.Background()
)

type SessionStore struct {
	Ssid         string    `redis:"ssid"`
	AccessToken  string    `redis:"access_token"`
	RefreshToken string    `redis:"refresh_token"`
	CreatedEpoch time.Time `redis:"created_epoch"`
}

type OAuthError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

func (store *SessionStore) Expired() bool {
	expiryTime := time.Until(store.CreatedEpoch.Add(AccessTokenExpiry))
	if expiryTime.Seconds() <= 0 {
		return true
	} else {
		return false
	}
}

func expandKey(key string) string {
	return strings.ToLower(RedisNamespace) + strings.ToLower(key)
}

func GenerateSessionId() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func StoreSession(w http.ResponseWriter, store *SessionStore) error {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionTokenCookieName,
		Value:    store.Ssid,
		Path:     "/",
		Secure:   server.Environment.RunContext == server.RunContextProduction,
		HttpOnly: true,
		MaxAge:   int(StoreExpiry),
	})
	key := expandKey(store.Ssid)
	cmd := server.RedisClient.HSet(redisContext, key, store)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	err := server.RedisClient.Expire(redisContext, key, StoreExpiry).Err()
	if err != nil {
		server.Logger.Warn(err.Error())
	}

	return nil
}

func DropSession(w http.ResponseWriter, store *SessionStore) error {
	http.SetCookie(w, &http.Cookie{
		Name:   SessionTokenCookieName,
		Value:  "",
		MaxAge: -1,
	})
	err := server.RedisClient.Del(redisContext, expandKey(store.Ssid)).Err()
	if err != nil {
		return err
	}

	return nil
}

func RetrieveSession(r *http.Request) (*SessionStore, error) {
	cookie, err := r.Cookie(SessionTokenCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrNeedsReAuth
		}
		return nil, err
	}

	key := expandKey(cookie.Value)

	var ssidStore SessionStore
	err = server.RedisClient.HGetAll(redisContext, key).Scan(&ssidStore)
	if err != nil {
		return nil, err
	}

	if ssidStore.Ssid == "" {
		return nil, ErrNeedsReAuth
	}

	return &ssidStore, nil
}

func FetchToken(with RefreshWithToken, token string) (*appTypes.TokenReply, error) {
	values := url.Values{
		"client_id":     {server.Environment.ClientId},
		"client_secret": {server.Environment.ClientSecret},
	}

	switch with {
	case RefreshWithAccessToken:
		values.Set("grant_type", "authorization_code")
		values.Set("code", token)
	case RefreshWithRefreshToken:
		values.Set("grant_type", "refresh_token")
		values.Set("refresh_token", token)
	}

	resp, err := server.HttpClient.PostForm("https://apis.roblox.com/oauth/v1/token", values)
	if err != nil {
		server.Logger.Warn(err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		server.Logger.Warn(err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized && with == RefreshWithRefreshToken {
			return nil, ErrNeedsReAuth
		}

		server.Logger.Warn("non-200 status code", "url", resp.Request.URL, "code", resp.Status, "body", string(body))
		return nil, ErrNeedsReAuth
	}

	var data appTypes.TokenReply
	err = json.Unmarshal(body, &data)
	if err != nil {
		server.Logger.Warn(err.Error())
		return nil, err
	}

	return &data, nil
}

func CreateSsidStore(w http.ResponseWriter, tokenReply *appTypes.TokenReply) (*SessionStore, error) {
	ssid, err := GenerateSessionId()
	if err != nil {
		return nil, err
	}
	store := &SessionStore{
		Ssid:         ssid,
		AccessToken:  tokenReply.AccessToken,
		RefreshToken: tokenReply.RefreshToken,
		CreatedEpoch: time.Now(),
	}
	err = StoreSession(w, store)

	if err != nil {
		return nil, err
	}

	return store, nil
}

func GetStore(w http.ResponseWriter, r *http.Request) (*SessionStore, error) {
	var store *SessionStore
	var err error

	store, err = RetrieveSession(r)
	if err != nil {
		return nil, err
	}

	if store.Expired() {
		server.Logger.Debug("code expired, refetching", "refresh token", store.RefreshToken[1:8]+"...", "access token (expired)", store.AccessToken[1:8]+"...")
		reply, err := FetchToken(RefreshWithRefreshToken, store.RefreshToken)
		if err != nil {
			return nil, err
		}

		store.AccessToken = reply.AccessToken
		store.RefreshToken = reply.RefreshToken
		store.CreatedEpoch = time.Now()

		StoreSession(w, store)

	} else {
		server.Logger.Debug("code valid, reusing", "access token (valid)", store.AccessToken[1:8]+"...")
	}

	return store, nil
}

func GetAccessToken(w http.ResponseWriter, r *http.Request) (string, error) {
	store, err := GetStore(w, r)
	if err != nil {
		return "", err
	}

	return store.AccessToken, nil
}
