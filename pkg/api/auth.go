package api

import (
	"net/http"
)

type AuthMiddleware struct {
	Token string
	Next  http.Handler
}

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if no token provided, skip auth check
	if a.Token == "" {
		a.Next.ServeHTTP(w, r)
		return
	}

	// Assume we're inside the Router and can access its routes.
	route, ok := a.Next.(*Router).Routes[r.URL.Path]
	if !ok || !route.AuthRequired {
		a.Next.ServeHTTP(w, r)
		return
	}

	headerToken := r.Header.Get("Authorization")
	if headerToken != a.Token {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	a.Next.ServeHTTP(w, r)
}
