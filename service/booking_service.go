package service

import (
	"context"
	"errors"
	"fmt"
	"movie-booking/models"
	"movie-booking/repository"
	"time"
)

const bookingHoldDuration = 5 * time.Minute

type BookingService struct {
	seatRepo    repository.SeatRepository
	bookingRepo repository.BookingRepository
	payment     PaymentGateway
	ticket      *TicketService
}

func NewBookingService(
	seatRepo repository.SeatRepository,
	bookingRepo repository.BookingRepository,
	payment PaymentGateway,
	ticket *TicketService,
) *BookingService {

	return &BookingService{
		seatRepo:    seatRepo,
		bookingRepo: bookingRepo,
		payment:     payment,
		ticket:      ticket,
	}
}

func (s *BookingService) CreateBooking(
	ctx context.Context,
	userID string,
	showID string,
	seatIDs []string,
) (*models.Booking, error) {

	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if showID == "" {
		return nil, errors.New("show ID is required")
	}

	if len(seatIDs) == 0 {
		return nil, errors.New("at least one seat is required")
	}

	// Prevent duplicate seat in same request

	var seen = make(map[string]bool)

	for _, seatID := range seatIDs {

		if seen[seatID] {
			return nil, fmt.Errorf(
				"Duplicate seat selected: %s",
				seatID,
			)
		}

		seen[seatID] = true
	}

	// Get seat price

	seats, err := s.seatRepo.GetSeats(
		showID,
		seatIDs,
	)

	if err != nil {
		return nil, err
	}

	var total int64

	for _, seat := range seats {
		total += seat.Price
	}

	bookingID := fmt.Sprintf(
		"BOOK-%d",
		time.Now().UnixNano(),
	)

	now := time.Now()

	expiresAt := now.Add(
		bookingHoldDuration,
	)

	booking := &models.Booking{
		ID:        bookingID,
		UserID:    userID,
		ShowID:    showID,
		SeatIDs:   append([]string(nil), seatIDs...),
		Status:    models.Pending,
		Total:     total,
		CreatedAt: now,
		ExpiredAt: expiresAt,
	}

	/*
		Most important operation.

		HoldSeats internally performs:
		check availability
		+
		update state

		atomically.
	*/

	err = s.seatRepo.HoldSeats(
		showID,
		seatIDs,
		bookingID,
		expiresAt,
	)

	if err != nil {
		return nil, err
	}

	/*
		Create booking after seats are successfully held.
	*/

	err = s.bookingRepo.Create(booking)

	if err != nil {

		// Compensating action.

		_ = s.seatRepo.ReleaseHeldSeats(
			showID,
			seatIDs,
			bookingID,
		)

		return nil, err
	}

	return booking, nil
}

func (s *BookingService) ConfirmBooking(
	ctx context.Context,
	bookingID string,
) (*models.Ticket, error) {

	booking, err := s.bookingRepo.GetByID(
		bookingID,
	)

	if err != nil {
		return nil, err
	}

	/*
		Idempotency:

		If already confirmed,
		don't charge user again.
	*/

	if booking.Status == models.Confirmed {
		return s.ticket.GenerateTicket(
			booking,
		)
	}

	if booking.Status != models.Pending {
		return nil, errors.New(
			"booking is not in pending state",
		)
	}

	if !booking.ExpiredAt.After(time.Now()) {
		booking.Status = models.Expired

		_ = s.bookingRepo.Update(booking)

		_ = s.seatRepo.ReleaseHeldSeats(
			booking.ShowID,
			booking.SeatIDs,
			booking.ID,
		)

		return nil, errors.New(
			"booking has expired",
		)
	}

	/*
			Payment happens OUTSIDE any seat lock.

		We don't want to hold inventory lock
		while calling an external payment gateway.
	*/

	payment, err := s.payment.Pay(
		ctx,
		booking.ID,
		booking.Total,
	)

	if err != nil || payment.Status != models.PaymentSuccess {

		booking.Status = models.Failed

		_ = s.bookingRepo.Update(booking)

		_ = s.seatRepo.ReleaseHeldSeats(
			booking.ShowID,
			booking.SeatIDs,
			booking.ID,
		)

		if err != nil {
			return nil, err
		}

		return nil, errors.New(
			"payment failed.",
		)

	}

	/*
			Payment successful.

		Now convert HELD -> BOOKED.
	*/

	err = s.seatRepo.ConfirmSeats(
		booking.ShowID,
		booking.SeatIDs,
		booking.ID,
	)

	if err != nil {

		/*
			Important distributed-systems problem:

			Payment succeeded,
			but seat confirmation failed.

			Production:
			- retry confirmation
			- reconcile payment
			- refund if booking cannot be confirmed
		*/

		return nil, fmt.Errorf(
			"payment succeeded but seat confirmation failed: %w",
			err,
		)
	}

	booking.Status = models.Confirmed

	err = s.bookingRepo.Update(booking)

	if err != nil {

		/*
			Another distributed consistency issue.

			Seats are confirmed but booking update failed.

			Production system should recover this
			using durable state + retry/reconciliation.
		*/

		return nil, fmt.Errorf(
			"seat confirmed but booking update failed: %w",
			err,
		)

	}

	return s.ticket.GenerateTicket(
		booking,
	)

}

func (s *BookingService) CancelBooking(
	ctx context.Context,
	bookingID string,
) error {

	booking, err := s.bookingRepo.GetByID(
		bookingID,
	)

	if err != nil {
		return err
	}

	/*
		Idempotent cancellation.
	*/

	if booking.Status == models.Cancelled {
		return nil
	}

	if booking.Status != models.Confirmed {
		return errors.New(
			"only confirmed booking can be cancelled",
		)
	}

	/*
		Production system:

		1. Check cancellation policy.
		2. Initiate refund.
		3. Update refund state.
		4. Release seats.
	*/

	err = s.seatRepo.CancelBookedSeats(
		booking.ShowID,
		booking.SeatIDs,
		booking.ID,
	)

	if err != nil {
		return err
	}

	booking.Status = models.Cancelled

	return s.bookingRepo.Update(booking)
}

func (s *BookingService) GetBooking(
	bookingID string,
) (*models.Booking, error) {

	return s.bookingRepo.GetByID(
		bookingID,
	)
}

func (s *BookingService) ExpireBooking(
	bookingID string,
) error {

	booking, err := s.bookingRepo.GetByID(
		bookingID,
	)

	if err != nil {
		return err
	}

	if booking.Status != models.Pending {
		return nil
	}

	if booking.ExpiredAt.After(time.Now()) {
		return nil
	}

	err = s.seatRepo.ReleaseHeldSeats(
		booking.ShowID,
		booking.SeatIDs,
		booking.ID,
	)

	if err != nil {
		return err
	}

	booking.Status = models.Expired

	return s.bookingRepo.Update(booking)
}
