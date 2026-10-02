package service

import (
	"errors"
	"fmt"
	"movie-booking/models"
)

type TicketService struct {}

func NewTicketService() *TicketService{
	return &TicketService{}
}

func (s *TicketService) GenerateTicket(
	booking *models.Booking,
) (*models.Ticket, error) {

	if booking.Status != models.Confirmed {
		return nil, errors.New(
			"ticket can only be generated for confirmed bookings",
		)
	}

	ticket := &models.Ticket{
		ID:	"Ticket-" + booking.ID,
		BookingID: booking.ID,
		ShowID: booking.ShowID,
		SeatIDs: append([]string(nil), booking.SeatIDs...),
		QRCode: fmt.Sprintf("QR-%v", booking.ID),
	}

	return ticket, nil
}