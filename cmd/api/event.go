package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"github.com/AbrahamMayowa/ticketmania/internal/validator"
)

func (app *application) createEventHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Date        string `json:"date"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`	
		TicketTypes []struct {
			Name     string `json:"name"`
			Price    int64  `json:"price"`
			Currency string `json:"currency"`
			TotalQty int    `json:"total_qty"`
		} `json:"ticket_types"`
	}

	err := app.readJSON(w, r, &input)
	fmt.Println(input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	parsedDate, err := validator.ParseDate(input.Date)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if len(input.TicketTypes) == 0 {
		app.badRequestResponse(w, r, errors.New("At least one ticket type is required"))
		return
	}

	event := &data.Event{
		Title:       input.Title,
		Description: input.Description,
		Location:    input.Location,
		UserID:      *user.Id,
		Status:      data.EventPublished	,
		Date:        parsedDate,
		StartTime: input.StartTime,
		EndTime: input.EndTime,
	}

	v := validator.New()

	ticketTypes := []*data.TicketType{}
	for _, tt := range input.TicketTypes {
		ticket := &data.TicketType{
			Name:     tt.Name,
			Price:    tt.Price,
			Currency: tt.Currency,
			TotalQty: tt.TotalQty,
		}

		if data.ValidateTicketType(v, ticket); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		ticketTypes = append(ticketTypes, ticket)
	}

	if data.ValidateEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	

	app.logger.PrintInfo("create ticket event", map[string]string{"event": event.Title, "user": strconv.FormatInt(*user.Id, 10)})
	err = app.models.Events.InsertEvent(event, ticketTypes)
	if err != nil {
		app.serverErrorResponse(w, r, err, input)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) getEventHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	event, err := app.models.Events.GetWithTicketTypes(r.Context(), id)
	if err != nil {
	switch {
	case errors.Is(err, data.ErrRecordNotFound):
		app.notFoundResponse(w, r)
		return
	default:
		app.serverErrorResponse(w, r, err)
		return
	}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"data": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	perPage := app.readInt(qs, "limit", 20)
	page := app.readInt(qs, "page", 1)
	offset := (page - 1) * perPage

	events, err := app.models.Events.GetEventList(r.Context(), perPage, offset)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w,r)
			return; 
		default:
			app.serverErrorResponse(w, r, err, map[string]string{"page": strconv.Itoa(perPage), "offset": strconv.Itoa(offset)})
			return
		}
		
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"event": events}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
