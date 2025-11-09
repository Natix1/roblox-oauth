package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	appTypes "github.com/natix1/roblox-oauth/src/app_types"
	"github.com/natix1/roblox-oauth/src/server"
	"github.com/natix1/roblox-oauth/src/session"
)

func AuthUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	server.Logger.Debug("requested auth url")

	redirectUrl, err := url.Parse("https://apis.roblox.com/oauth/v1/authorize")
	if err != nil {
		panic("Failed parsing literal string URL")
	}

	values := redirectUrl.Query()
	values.Add("client_id", server.Environment.ClientId)
	values.Add("redirect_uri", server.Environment.ClientURI)
	values.Add("scope", server.Environment.OAuthScope)
	values.Add("response_type", "code")
	redirectUrl.RawQuery = values.Encode()

	serialized, err := json.Marshal(appTypes.AuthUriReply{
		Url: redirectUrl.String(),
	})
	if err != nil {
		server.Logger.Warn(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(serialized)
}

func AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	server.Logger.Debug("callback handler called")
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	code := query.Get("code")
	if code == "" {
		errorName := query.Get("error")
		errorDescription := query.Get("error_description")

		if errorName != "" && errorDescription == "" {
			http.Error(w, fmt.Sprintf("%s: %s", errorName, errorDescription), http.StatusInternalServerError)
		} else {
			http.Error(w, "Invalid URL query parameters", http.StatusBadRequest)
		}

		return
	}
	server.Logger.Debug("our access token is", code, "exchanging now")

	tokenReply, err := session.FetchToken(session.RefreshWithAccessToken, code)
	if err != nil {
		server.Logger.Warn("error in callback auth while fetching code data", "error", err.Error(), "code", code)
		return
	}
	_, err = session.CreateSsidStore(w, tokenReply)
	if err != nil {
		server.Logger.Warn("error in callback auth while creating ssid store", "error", err.Error(), "code", code)
		return
	}

	http.Redirect(w, r, server.Environment.CORSDomain+"/dash", http.StatusTemporaryRedirect)
}

func AuthLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	store, err := session.GetStore(w, r)
	if err != nil {
		if err == session.ErrNeedsReAuth {
			http.Error(w, "Already logged out", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	session.DropSession(w, store)
	http.Redirect(w, r, server.Environment.CORSDomain+"/", http.StatusTemporaryRedirect)
}
