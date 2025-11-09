package handlers_session

import (
	"io"
	"net/http"

	"github.com/natix1/roblox-oauth/src/server"
	"github.com/natix1/roblox-oauth/src/session"
)

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	token, err := session.GetAccessToken(w, r)
	if err != nil {
		server.Logger.Warn(err.Error())
		http.Error(w, "Failed getting access token", http.StatusInternalServerError)
		return
	}

	request, err := http.NewRequest(http.MethodGet, "https://apis.roblox.com/oauth/v1/userinfo", nil)
	if err != nil {
		server.Logger.Warn(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := server.HttpClient.Do(request)
	if err != nil {
		server.Logger.Warn(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if response.StatusCode != http.StatusOK {
		server.Logger.Warn("non 200 status code", "status", response.Status)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		server.Logger.Warn(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()
	w.Write(body)
}
