package middleware

import (
	"net/http"

	"github.com/natix1/roblox-oauth/src/server"
	"github.com/natix1/roblox-oauth/src/session"
)

func AccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := session.GetAccessToken(w, r)
		if err != nil {
			if err == session.ErrNeedsReAuth {
				http.Redirect(w, r, server.Environment.CORSDomain+"/login", http.StatusTemporaryRedirect)
			}

			server.Logger.Warn("error while doing access token middleware", "error", err.Error())
		}

		next.ServeHTTP(w, r)
	})
}
