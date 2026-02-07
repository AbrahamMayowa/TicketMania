package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) router() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/v1/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/login", app.LoginUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/create-event", app.requireAuthentication(app.createEventHandler))
	router.HandlerFunc(http.MethodGet, "/v1/events", app.listEventsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/events/:id", app.getEventHandler)
	router.HandlerFunc(http.MethodPost, "/v1/buy-ticket", app.createTicket )

	return app.recoverPanic(app.authenticate(router))
}
