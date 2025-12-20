package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) router() *httprouter.Router {
	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/v1/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/login", app.LoginUserHandler)
	return router
}