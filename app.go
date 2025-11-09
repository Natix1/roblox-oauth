package main

import (
	"net/http"

	handlers_login "github.com/natix1/roblox-oauth/src/handlers/auth"
	handlers_session "github.com/natix1/roblox-oauth/src/handlers/session"
	"github.com/natix1/roblox-oauth/src/middleware"
)

func authMux() *http.ServeMux {
	loginMux := http.NewServeMux()
	loginMux.HandleFunc("/auth_url", handlers_login.AuthUrlHandler)
	loginMux.HandleFunc("/callback", handlers_login.CallbackHandler)

	return loginMux
}

func sessionMux() *http.ServeMux {
	sessionMux := http.NewServeMux()
	sessionMux.HandleFunc("/session", handlers_session.SessionHandler)

	return sessionMux
}

func main() {
	root := http.NewServeMux()

	root.Handle("/auth/", http.StripPrefix("/auth", authMux()))
	root.Handle("/sessions/", http.StripPrefix("/sessions", middleware.AccessToken(sessionMux())))

	http.ListenAndServe(":6969", middleware.Cors(root))
}
