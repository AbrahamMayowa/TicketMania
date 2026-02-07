package main

import (
	"fmt"
	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"net/http"
	"strings"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// no token present, set user to anonymous and proceed
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		authicateValue := strings.Split(authHeader, " ")
		if len(authicateValue) != 2 || authicateValue[0] != "Bearer" {
			app.unauthorizedResponse(w, r, "invalid or expired token")
			return
		}

		claims, err := app.ValidateToken(authicateValue[1])
		if err != nil {
			app.unauthorizedResponse(w, r, "invalid or expired token")
			return
		}

		user := &data.User{
			Id:    &claims.Id,
			Email: claims.Email,
			Scope: string(claims.Scope),
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.requireAuthenticationResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
