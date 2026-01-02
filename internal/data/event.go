package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"time"
	"github.com/AbrahamMayowa/ticketmania/internal/validator"
)

type EventStatus string
type TicketStatus string

const (
	EventDraft     EventStatus = "draft"
	EventPublished EventStatus = "published"
	EventCancelled EventStatus = "cancelled"
	EventCompleted EventStatus = "completed"

	TicketAvailable TicketStatus = "available"
	TicketReserved  TicketStatus = "reserved"
	TicketPaid      TicketStatus = "paid"
	TicketCancelled TicketStatus = "cancelled"
	TicketUsed      TicketStatus = "used"
)

type Event struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`

	Location  string    `json:"location"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`

	UserID int64 `json:"user_id"`

	Status EventStatus `json:"status"`

	Date time.Time `json:"date"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TicketType struct {
	ID      int64  `json:"id"`
	EventID int64  `json:"event_id"`
	Event   *Event `json:"event,omitempty"`

	Name     string `json:"name"` // Regular, VIP, Early Bird
	Price    int64  `json:"price"`
	Currency string `json:"currency"`

	TotalQty int `json:"total_qty"` // total available
	SoldQty  int `json:"sold_qty"`  // cached counter (important)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Ticket struct {
	ID           int64 `json:"id"`
	EventID      int64 `json:"event_id"`
	TicketTypeID int64 `json:"ticket_type_id"`

	UserID *int64 `json:"user_id,omitempty"`

	Status TicketStatus `json:"status"`

	PaidAt *time.Time `json:"paid_at,omitempty"`
	UsedAt *time.Time `json:"used_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	BuyerEmail string `json:"buyer_email"`
	BuyerPhone string `json:"buyer_phone"`
}

type EventModel struct {
	DB *sql.DB
}

type TicketTypeModel struct {
	DB *sql.DB
}

type TicketModel struct {
	DB *sql.DB
}

type EventWithTicketTypes struct {
	Event       Event         `json:"event"`
	TicketTypes []*TicketType `json:"ticket_types"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	Total       int `json:"total"`
}

type EventListResponse struct {
	Data []*EventWithTicketTypes `json:"data"`
	Meta PaginationMeta         `json:"meta"`
}

func validateTimeRange(start, end string) bool {
	startT, err := time.Parse("15:04", start)
	if err != nil {
		return false
	}

	endT, err := time.Parse("15:04", end)
	if err != nil {
		return false
	}

	if !startT.Before(endT) {
		return false
	}

	return true
}

func ValidateEvent(v *validator.Validator, e *Event) {
	v.Check(e.Title != "", "title", "must be provided")
	v.Check(len(e.Title) <= 500, "title", "must not be more than 500 bytes")
	v.Check(e.StartTime != "", "start_time", "must be provided")
	v.Check(e.EndTime != "", "end_time", "must be provided")
	v.Check(!e.Date.IsZero(), "date", "must be provided")
	v.Check(validateTimeRange(e.StartTime, e.EndTime), "start_time", "must be valid time, and start_time before end_time")
	v.Check(e.UserID > 0, "user_id", "must be provided")
}

// ValidateTicketType runs basic checks on a TicketType.
func ValidateTicketType(v *validator.Validator, tt *TicketType) {
	v.Check(tt.Name != "", "name", "must be provided")
	v.Check(len(tt.Name) <= 255, "name", "must not be more than 255 bytes")
	v.Check(tt.Price >= 0, "price", "must be >= 0")
	v.Check(tt.Currency != "", "currency", "must be provided")
	v.Check(tt.TotalQty >= 0, "total_qty", "must be >= 0")
}

func (m EventModel) InsertEvent(e *Event, ticketTypes []*TicketType) error {
	ctx := context.Background()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Insert event
	eventQuery := `
        INSERT INTO events (title, description, location, start_time, end_time, user_id, status, date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at
    `
	err = tx.QueryRowContext(ctx, eventQuery,
		e.Title,
		e.Description,
		e.Location,
		e.StartTime,
		e.EndTime,
		e.UserID,
		e.Status,
		e.Date,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert ticket types
	for _, tt := range ticketTypes {
		tt.EventID = e.ID

		ttQuery := `
            INSERT INTO ticket_types
                (event_id, name, price, currency, total_qty, sold_qty)
            VALUES ($1,$2,$3,$4,$5,$6)
            RETURNING id, created_at, updated_at
        `
		err = tx.QueryRowContext(ctx, ttQuery,
			tt.EventID,
			tt.Name,
			tt.Price,
			tt.Currency,
			tt.TotalQty,
			tt.SoldQty,
		).Scan(&tt.ID, &tt.CreatedAt, &tt.UpdatedAt)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (m EventModel) GetWithTicketTypes(ctx context.Context, eventID int64) (*EventWithTicketTypes, error) {
	query := `
	SELECT
	  e.id,
	  e.title,
	  e.description,
	  e.location,
	  e.start_time,
	  e.end_time,
	  e.user_id,
	  e.status,
	  e.created_at,
	  e.updated_at,
	  e.date,

	  tt.id,
	  tt.event_id,
	  tt.name,
	  tt.price,
	  tt.currency,
	  tt.total_qty,
	  tt.sold_qty,
	  tt.created_at,
	  tt.updated_at
	FROM events e
	LEFT JOIN ticket_types tt ON tt.event_id = e.id
	WHERE e.id = $1
	ORDER BY tt.created_at ASC;
	`

	rows, err := m.DB.QueryContext(ctx, query, eventID)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var event *Event
	ticketTypes := make([]*TicketType, 0)

	for rows.Next() {
		var e Event
		tt  := &TicketType{}

		err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Location,
			&e.StartTime,
			&e.EndTime,
			&e.UserID,
			&e.Status,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.Date,

			&tt.ID,
			&tt.EventID,
			&tt.Name,
			&tt.Price,
			&tt.Currency,
			&tt.TotalQty,
			&tt.SoldQty,
			&tt.CreatedAt,
			&tt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Initialize event once
		if event == nil {
			event = &e
		}
		if tt.ID != 0 {
			ticketTypes = append(ticketTypes, tt)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if event == nil {
		return nil, sql.ErrNoRows
	}

	return &EventWithTicketTypes{
		Event:       *event,
		TicketTypes: ticketTypes,
	}, nil
}

func (m EventModel) GetEventList(ctx context.Context, perPage int,
	offset int) (*EventListResponse, error) {
	query := `
	SELECT
		e.id,
		e.title,
		e.description,
		e.location,
		e.start_time,
		e.end_time,
		e.user_id,
		e.status,
		e.created_at,
		e.updated_at,
		e.date,
		COALESCE(
			json_agg(
				json_build_object(
					'id', tt.id,
					'name', tt.name,
					'price', tt.price,
					'currency', tt.currency,
					'total_qty', tt.total_qty,
					'sold_qty', tt.sold_qty,
					'created_at', tt.created_at,
					'updated_at', tt.updated_at
				) ORDER BY tt.id
			) FILTER (WHERE
				tt.id IS NOT NULL
				AND now() BETWEEN tt.sales_start AND tt.sales_end
				AND tt.total_qty > tt.sold_qty
			),
			'[]'
		) AS ticket_types
	FROM events e
	LEFT JOIN ticket_types tt ON tt.event_id = e.id
	WHERE e.status = 'published'
	  AND EXISTS (
	      SELECT 1 FROM ticket_types tt2
	      WHERE tt2.event_id = e.id
	        AND now() BETWEEN tt2.sales_start AND tt2.sales_end
	        AND tt2.total_qty > tt2.sold_qty
	  )
	GROUP BY e.id
	ORDER BY e.start_time ASC
	LIMIT $1 OFFSET $2;
	`

	countQuery := `
		SELECT COUNT(*)
		FROM events e
		WHERE e.status = 'published'
		  AND EXISTS (
		      SELECT 1 FROM ticket_types tt
		      WHERE tt.event_id = e.id
		        AND now() BETWEEN tt.sales_start AND tt.sales_end
		        AND tt.total_qty > tt.sold_qty
		  )
	`

	// Execute count query first
	var totalEvents int
	err := m.DB.QueryRowContext(ctx, countQuery).Scan(&totalEvents)
	if err != nil {
		return nil, err
	}

	// Execute main query
	rows, err := m.DB.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*EventWithTicketTypes{}

	for rows.Next() {
		var e Event
		var ticketTypesJSON []byte

		err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Location,
			&e.StartTime,
			&e.EndTime,
			&e.UserID,
			&e.Status,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.Date,
			&ticketTypesJSON,
		)
		if err != nil {
			return nil, err
		}

		var ticketTypes []*TicketType
		err = json.Unmarshal(ticketTypesJSON, &ticketTypes)
		if err != nil {
			return nil, err
		}

		eventWithTypes := &EventWithTicketTypes{
			Event:       e,
			TicketTypes: ticketTypes,
		}
		events = append(events, eventWithTypes)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, sql.ErrNoRows
	}

	return &EventListResponse{
		Data: events,
		Meta: PaginationMeta{
			CurrentPage: (offset / perPage) + 1,
			PerPage:     perPage,
			Total:       totalEvents,
			TotalPages:  int(math.Ceil(float64(totalEvents) / float64(perPage))),
		},
	}, nil
}
