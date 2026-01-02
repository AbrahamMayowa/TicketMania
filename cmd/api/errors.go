package main

import (
	"net/http"
	"fmt"
)

func (app *application) logError(err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message, "status": status, "success": false}

	err :=  app.writeJSON(w, status, env, nil)
	
	if err != nil {
		app.logError(err)
		w.WriteHeader(500)
	}

}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(err)
	message := "Internal server occured"
	app.errorResponse(w, r, http.StatusInternalServerError, message);
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error, message string) {
	app.logError(err)
	app.errorResponse(w, r, http.StatusConflict, message);
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(err)
	app.errorResponse(w, r, http.StatusBadRequest, err.Error());
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message) 
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not allowed for this resources", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) invalidCredentialResponsee(w http.ResponseWriter, r *http.Request) {
	message := "Password or email is invalid";
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	message := "you do not have permission to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
