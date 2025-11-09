package main

import (
	"net/http"

	"github.com/natix1/roblox-oauth/src/handlers"
	"github.com/natix1/roblox-oauth/src/middleware"
)

func authMux() *http.ServeMux {
	loginMux := http.NewServeMux()
	loginMux.HandleFunc("/auth_url", handlers.AuthUrlHandler)
	loginMux.HandleFunc("/callback", handlers.AuthCallbackHandler)

	return loginMux
}

func sessionMux() *http.ServeMux {
	sessionMux := http.NewServeMux()
	sessionMux.HandleFunc("/session", handlers.SessionHandler)

	return sessionMux
}

func main() {
	root := http.NewServeMux()

	root.Handle("/auth/", http.StripPrefix("/auth", authMux()))
	root.Handle("/sessions/", http.StripPrefix("/sessions", middleware.AccessToken(sessionMux())))

	http.ListenAndServe(":6969", middleware.Cors(root))
}
