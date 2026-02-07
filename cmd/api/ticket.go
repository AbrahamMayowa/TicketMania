package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AbrahamMayowa/ticketmania/internal/data"
	"github.com/AbrahamMayowa/ticketmania/internal/validator"
)

func (app *application) createTicket(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		EventID *int64 `json:"eventId"`
		TicketTypes []struct {
			TicketTypeID *int64 `json:"ticketTypeId"`
			Quantity     int `json:"quantity"`
			BuyerEmail   string `json:"buyerEmail"`
			BuyerPhone   string `json:"buyerPhone"`
		} `json:"ticketTypes"`
	}

	err := app.readJSON(w,r, &input)

	if err != nil {
		app.badRequestResponse(w,r,err)
		return
	}

	if len(input.TicketTypes) == 0 {
		app.badRequestResponse(w, r, errors.New("At least one ticket type is required"))
		return
	}

	v := validator.New()

	if input.EventID == nil {
		app.badRequestResponse(w, r, errors.New("EventId is required"))
		return
	}

	//anonymous user can still create ticket
	ticketType := &data.TicketPurchaseRequest{
		EventID: input.EventID,
		
	}

	fmt.Printf("user.Id: %+v\n", user.Id)

	if user.Id != nil {
		ticketType.UserID = user.Id
	}

	fmt.Printf("ticketType: %+v\n", ticketType)

	for _,item := range input.TicketTypes {
		ticketItem := &data.TicketPurchaseItem{
			TicketTypeID: item.TicketTypeID,
			Quantity: item.Quantity,
			BuyerEmail: item.BuyerEmail,
			BuyerPhone: item.BuyerPhone,
		}
		if data.ValidateTicket(v, ticketItem); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return;
		}
		ticketType.Items = append(ticketType.Items, ticketItem)
	}

	newTickets, err := app.models.Tickets.InsertTickets(ticketType)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrTicketNotFound):
			app.badRequestResponse(w, r, err)
			return
		case errors.Is(err, data.ErrTicketNotAvailable):
			app.badRequestResponse(w, r, err)
			return
		default:
			app.serverErrorResponse(w, r, err, input)
			return
		}
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": newTickets}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}