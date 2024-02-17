package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var users = map[string]string{
	"test": "secret",
}

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func is_authorization(r *http.Request) bool {
	is_auth := false
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if ok {
		is_auth = auth
	} else {
		username, password, ok := r.BasicAuth()
		if ok {
			pass, ok := users[username]
			if ok {
				is_auth = password == pass
			}
		}
	}

	return is_auth
}

func view_basic_auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if !is_authorization(r) {
		w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid username or password"}`))
	} else {
		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = true
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "welcome to golang world!"}`))
	}
}

func view_sign_out(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}
