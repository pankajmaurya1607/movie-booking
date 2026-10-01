package repository

import (
	"errors"
	"movie-booking/models"
	"sync"
	"time"
)

type SeatRepository interface {
	AddSeat(seat *models.ShowSeat) error

	GetSeats(
		showID string,
		seatIDs []string,
	) ([]*models.ShowSeat, error)

	GetAllSeats(
		showID string,
	) ([]*models.ShowSeat, error)

	HoldSeats(
		showID string,
		seatIDs []string,
		bookingID string,
		expiresAt time.Time,
	) error

	ConfirmSeats(
		showID string,
		seatIDs []string,
		bookingID string,
	) error

	ReleaseHeldSeats(
		showID string,
		seatIDs []string,
		bookingID string,
	) error

	CancelBookedSeats(
		showID string,
		seatIDs []string,
		bookingID string,
	) error
}

type InMemorySeatRepository struct {
	mu    sync.RWMutex
	seats map[string]*models.ShowSeat
}

func NewInMemorySeatRepository() *InMemorySeatRepository {
	return &InMemorySeatRepository{
		seats: make(map[string]*models.ShowSeat),
	}
}

func seatKey(showID, seatID string) string {
	return showID + ":" + seatID
}

func (r *InMemorySeatRepository) AddSeat(
	seat *models.ShowSeat,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := seatKey(seat.ShowID, seat.SeatID)

	if _, exists := r.seats[key]; exists {
		return errors.New("seat already exists")
	}

	copySeat := *seat
	r.seats[key] = &copySeat

	return nil
}

func (r *InMemorySeatRepository) GetSeats(
	showID string,
	seatIDs []string,
) ([]*models.ShowSeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	results := make([]*models.ShowSeat, 0, len(seatIDs))

	for _, seatId := range seatIDs {

		seat, exists := r.seats[seatKey(showID, seatId)]

		if !exists {
			return nil, errors.New("seat not found: " + seatId)
		}

		seatCopy := *seat
		results = append(results, &seatCopy)
	}
	return results, nil
}

func (r *InMemorySeatRepository) GetAllSeats(
	showID string,
) ([]*models.ShowSeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*models.ShowSeat

	for _, seat := range r.seats {

		if seat.ShowID != showID {
			continue
		}

		seatCopy := *seat
		result = append(result, &seatCopy)
	}
	return result, nil
}

func (r *InMemorySeatRepository) HoldSeats(
	showID string,
	seatIDs []string,
	bookingID string,
	expiresAt time.Time,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	/*
		Step 1:
		Validate all seats first.

		We don't modify anything yet
	*/

	for _, seatID := range seatIDs {

		seat, exists := r.seats[seatKey(showID, seatID)]

		if !exists {
			return errors.New("seat not found: " + seatID)
		}

		switch seat.Status {

		case models.Booked:
			return errors.New("seat already booked: " + seatID)

		case models.Held:
			if seat.ExpiresAt.After(now) {
				return errors.New("seat already held: " + seatID)
			}

			// Existing hold has expired
			// We can reclaim it

		case models.Available:
			// Fine.

		default:
			return errors.New("invalid seat status")

		}
	}

	/*
		Step 2:
		Only after every seat is validated,
		hold all seats.
	*/

	for _, seatID := range seatIDs {

		seat := r.seats[seatKey(showID, seatID)]

		seat.Status = models.Held
		seat.BookingID = bookingID
		seat.ExpiresAt = expiresAt
	}
	return nil
}

func (r *InMemorySeatRepository) ConfirmSeats(
	showID string,
	seatIDs []string,
	bookingID string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Validate first.

	for _, seatID := range seatIDs {

		seat, exists := r.seats[seatKey(showID, seatID)]

		if !exists {
			return errors.New("seat not found")
		}

		if seat.Status != models.Held {
			return errors.New("seat is not held: " + seatID)
		}

		if seat.BookingID != bookingID {
			return errors.New("seat belong to another booking")
		}

		if !seat.ExpiresAt.After(now) {
			return errors.New("seat hold expired")
		}
	}

	// confirm

	for _, seatID := range seatIDs {

		seat := r.seats[seatKey(showID, seatID)]

		seat.Status = models.Booked
		seat.ExpiresAt = time.Time{}
	}
	return nil
}

func (r *InMemorySeatRepository) ReleaseHeldSeats(
	showID string,
	seatIDs []string,
	bookingID string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, seatID := range seatIDs {

		seat, exists := r.seats[seatKey(showID, seatID)]

		if !exists {
			return errors.New("seat not found")
		}

		if seat.Status == models.Held &&
			seat.BookingID == bookingID {

			seat.Status = models.Available
			seat.BookingID = ""
			seat.ExpiresAt = time.Time{}
		}
	}
	return nil
}

func (r *InMemorySeatRepository) CancelBookedSeats(
	showID string,
	seatIDs []string,
	bookingID string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, seatID := range seatIDs {

		seat, exists := r.seats[seatKey(showID, seatID)]

		if !exists {
			return errors.New("seat not found")
		}

		if seat.Status != models.Booked {
			return errors.New("seat is not booked: " + seatID)
		}

		if seat.BookingID != bookingID {
			return errors.New("this seat belong to different booking")
		}
	}

	// Release

	for _, seatID := range seatIDs {

		seat := r.seats[seatKey(showID, seatID)]

		seat.Status = models.Available
		seat.BookingID = ""
		seat.ExpiresAt = time.Time{}
	}
	return nil
}
