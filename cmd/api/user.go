package main

import (
	"errors"
	"net/http"

	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"github.com/AbrahamMayowa/ticketmania/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email: input.Email,
	}

	v := validator.New()

	v.Check(input.Password != "", "password", "must be provided")

	if input.Password != "" {
		v.Check(validator.ValidatePassword(input.Password), "password", "must be at least 6 characters long, contain at least one number and one special character, and be no more than 12 characters long")
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data.ValidateUser(v, user)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	_, err = app.models.Users.GetByEmail(user.Email)

	switch {
	case err == nil:
		app.conflictResponse(w, r, nil, "A user with this email address already exists")
		return
	case errors.Is(err, data.ErrRecordNotFound):
		// user does NOT exist continue
	default:
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserAlreadyExists):
			app.conflictResponse(w, r, err, "A user with this email address already exists")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	env := envelope{"data": user}
	err = app.writeJSON(w, http.StatusOK, env, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string  `json:"email"`
		Password *string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Password != nil, "password", "must be provided")
	v.Check(validator.Matches(input.Email, validator.EmailRegex), "email", "Valid email must be provided")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponsee(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	isValidPassword, err := user.Password.Matches(*input.Password)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !isValidPassword {
		app.invalidCredentialResponsee(w, r)
		return
	}

	token, err := app.GenerateToken(ScopeAuthentication, user)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"token": token, "user": user}
	err = app.writeJSON(w, http.StatusOK, env, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
