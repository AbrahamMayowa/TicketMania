package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("Record not found")
	ErrUserAlreadyExists  = errors.New("User with given email already exists")
)

type Models struct {
	Users      UserModel
	Events     EventModel
	Tickets    TicketModel
	TicketTypes TicketTypeModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Events: EventModel{DB: db},
		Tickets: TicketModel{DB: db},
		TicketTypes: TicketTypeModel{DB: db},
	}
}