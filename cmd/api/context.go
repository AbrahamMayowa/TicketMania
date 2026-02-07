package main

import (
	"context"
	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"net/http"
)

type contextKey string

const contextUserKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), contextUserKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(contextUserKey).(*data.User)
	if !ok {
		panic("missing user value in context")
	}
	return user
}
