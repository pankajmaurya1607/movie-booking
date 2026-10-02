package repository

import (
	"errors"
	"movie-booking/models"
	"sync"
)


type BookingRepository interface {
	Create(booking *models.Booking) error

	GetByID(id string) (*models.Booking, error) 

	Update(booking *models.Booking) error
}	

type InMemoryBookingRepository struct {
	mu 		sync.RWMutex
	bookings map[string]*models.Booking
}

func NewInMemoryBookingRepository() *InMemoryBookingRepository {
	return &InMemoryBookingRepository{
		bookings: make(map[string]*models.Booking),
	}
}

func (r *InMemoryBookingRepository) Create(
	booking *models.Booking,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookings[booking.ID]; exists {
		return errors.New("booking already exist")
	}

	copyBooking := *booking
	copyBooking.SeatIDs = append([]string(nil), booking.SeatIDs...)

	r.bookings[booking.ID] = &copyBooking

	return nil
}

func (r *InMemoryBookingRepository) GetByID (
	id string, 
) (*models.Booking, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	booking, exists := r.bookings[id]; 

	if !exists {
		return nil, errors.New("booking not found")
	}

	copyBooking := *booking
	copyBooking.SeatIDs = append([]string(nil), booking.SeatIDs...)

	return &copyBooking, nil
}

func (r *InMemoryBookingRepository) Update (
	booking *models.Booking,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookings[booking.ID]; !exists {
		return errors.New("booking not found")
	}

	copyBooking := *booking
	copyBooking.SeatIDs = append([]string(nil), booking.SeatIDs...)

	r.bookings[booking.ID] = &copyBooking
	return nil
}