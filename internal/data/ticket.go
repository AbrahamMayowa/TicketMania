package data

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/AbrahamMayowa/ticketmania/internal/validator"
)

type TicketStatus string

const (
	TicketAvailable TicketStatus = "available"
	TicketReserved  TicketStatus = "reserved"
	TicketPaid      TicketStatus = "paid"
	TicketCancelled TicketStatus = "cancelled"
	TicketUsed      TicketStatus = "used"
)

type Ticket struct {
	ID           int64  `json:"id"`
	EventID      *int64  `json:"event_id"`
	TicketTypeID *int64 `json:"ticket_type_id"`

	UserID *int64 `json:"user_id,omitempty"`

	Status TicketStatus `json:"status"`

	PaidAt *time.Time `json:"paid_at,omitempty"`
	UsedAt *time.Time `json:"used_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	BuyerEmail *string `json:"buyer_email"`
	BuyerPhone *string `json:"buyer_phone"`
}

type TicketModel struct {
	DB *sql.DB
}

func ValidateTicket(v *validator.Validator, t *TicketPurchaseItem) {
	v.Check(t.TicketTypeID != nil, "ticket type", "must be provided")
	v.Check(t.BuyerEmail != "", "buyer_email", "must be provided")
	v.Check(t.BuyerPhone != "", "buyer_phone", "must be provided")
	v.Check(t.Quantity != 0, "quantity", "must be provided")
}


// TicketPurchaseResult contains the created tickets
type TicketPurchaseResult struct {
	Tickets []*Ticket
}


type TicketPurchaseItem struct {
	TicketTypeID *int64
	Quantity     int
	BuyerEmail   string
	BuyerPhone   string
}

// TicketPurchaseRequest represents the entire purchase request
type TicketPurchaseRequest struct {
	EventID *int64
	UserID  *int64
	Items   []*TicketPurchaseItem
}

func (m TicketModel) InsertTickets(tickets *TicketPurchaseRequest) (*TicketPurchaseResult, error) {
	for i, item := range tickets.Items {
    if item != nil {
        fmt.Printf("Item %d: %+v\n", i, *item)
    }
}
	// Start a transaction
	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}
	
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	result := &TicketPurchaseResult{
		Tickets: make([]*Ticket, 0),
	}

	// Group items by ticket type to get total quantities needed
	ticketTypeQuantities := make(map[*int64]int)
	for _, item := range tickets.Items {
		ticketTypeQuantities[item.TicketTypeID] += item.Quantity
	}

	fmt.Println(ticketTypeQuantities)

	// Lock all ticket types and verify availability
	ticketTypes := make(map[int64]*TicketType)
	for ticketTypeID, totalQty := range ticketTypeQuantities {
		var tt TicketType
		query := `
			SELECT id, event_id, name, price, currency, total_qty, sold_qty, created_at, updated_at
			FROM ticket_types
			WHERE id = $1 AND event_id = $2
			FOR UPDATE
		`
	
		err = tx.QueryRow(query, ticketTypeID, tickets.EventID).Scan(
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
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("ticket type %d not found: %w", *ticketTypeID, ErrTicketNotFound)
			}
			return nil, err
		}

		// Check availability
		availableQty := tt.TotalQty - tt.SoldQty
		if availableQty < totalQty {
			return nil, fmt.Errorf("insufficient tickets for type %s: requested %d, available %d: %w", 
				tt.Name, totalQty, availableQty, ErrTicketNotAvailable)
		}

		ticketTypes[*ticketTypeID] = &tt
	}

	// Insert tickets for each item
	insertQuery := `
		INSERT INTO tickets (event_id, ticket_type_id, user_id, status, paid_at, used_at, buyer_email, buyer_phone, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	for _, item := range tickets.Items {
		// Create tickets based on quantity
		for i := 0; i < item.Quantity; i++ {
			ticket := &Ticket{
				EventID:      tickets.EventID,
				TicketTypeID: item.TicketTypeID,
				Status:       TicketPaid, 
				BuyerEmail:   &item.BuyerEmail,
				BuyerPhone:   &item.BuyerPhone,
			}

			if tickets.UserID != nil {
				ticket.UserID = tickets.UserID
			}

			err = tx.QueryRow(
				insertQuery,
				ticket.EventID,
				ticket.TicketTypeID,
				ticket.UserID,
				ticket.Status,
				ticket.PaidAt,
				ticket.UsedAt,
				ticket.BuyerEmail,
				ticket.BuyerPhone,
				time.Now(),
			).Scan(&ticket.ID, &ticket.CreatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to insert ticket: %w", err)
			}

			result.Tickets = append(result.Tickets, ticket)
		}
	}

	// Update sold_qty for each ticket type
	updateQuery := `
		UPDATE ticket_types
		SET sold_qty = sold_qty + $1,
		    updated_at = $2
		WHERE id = $3
	`

	for ticketTypeID, qty := range ticketTypeQuantities {
		_, err = tx.Exec(updateQuery, qty, time.Now(), ticketTypeID)
		if err != nil {
			return nil, fmt.Errorf("failed to update sold quantity: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	return result, nil
}

