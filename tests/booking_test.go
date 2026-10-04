package tests

import (
	"context"
	"fmt"
	"movie-booking/models"
	"movie-booking/repository"
	"movie-booking/service"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCreateBooking(t *testing.T) {

	seatRepo := repository.NewInMemorySeatRepository()

	bookinRepo := repository.NewInMemoryBookingRepository()

	payment := &service.UPIProvider{}

	ticketService := service.NewTicketService()

	bookinService := service.NewBookingService(
		seatRepo,
		bookinRepo,
		payment,
		ticketService,
	)

	err := seatRepo.AddSeat(
		&models.ShowSeat{
			ShowID: "SHOW1",
			SeatID: "A1",
			Status: models.Available,
			Price:  25000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	booking, err := bookinService.CreateBooking(
		context.Background(),
		"USER1",
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	if booking.Status != models.Pending {
		t.Fatalf(
			"expected PENDING got %s",
			booking.Status,
		)
	}

}

func TestConfirmBooking(t *testing.T) {

	seatRepo := repository.NewInMemorySeatRepository()

	bookingRepo := repository.NewInMemoryBookingRepository()

	payment := &service.UPIProvider{
		ShouldFail: false,
	}

	ticketService := service.NewTicketService()

	bookingService := service.NewBookingService(
		seatRepo,
		bookingRepo,
		payment,
		ticketService,
	)

	err := seatRepo.AddSeat(
		&models.ShowSeat{
			ShowID: "SHOW1",
			SeatID: "A1",
			Status: models.Available,
			Price:  25000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	booking, err := bookingService.CreateBooking(
		context.Background(),
		"USER1",
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	ticket, err := bookingService.ConfirmBooking(
		context.Background(),
		booking.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	if ticket == nil {
		t.Fatal("ticket should not be nil")
	}

	finalBooking, err := bookingService.GetBooking(
		booking.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	if finalBooking.Status != models.Confirmed {
		t.Fatalf(
			"expected CONFIRMED, got %s",
			finalBooking.Status,
		)
	}
}

func TestPaymentFailure(t *testing.T) {

	seatRepo := repository.NewInMemorySeatRepository()

	bookingRepo := repository.NewInMemoryBookingRepository()

	payment := &service.UPIProvider{
		ShouldFail: true,
	}

	ticketService := service.NewTicketService()

	bookingService := service.NewBookingService(
		seatRepo,
		bookingRepo,
		payment,
		ticketService,
	)

	err := seatRepo.AddSeat(
		&models.ShowSeat{
			ShowID: "SHOW1",
			SeatID: "A1",
			Status: models.Available,
			Price:  25000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	booking, err := bookingService.CreateBooking(
		context.Background(),
		"USER1",
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	_, err = bookingService.ConfirmBooking(
		context.Background(),
		booking.ID,
	)

	if err == nil {
		t.Fatal("expected payment failure")
	}

	seats, err := seatRepo.GetSeats(
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	if seats[0].Status != models.Available {
		t.Fatalf(
			"expected seat AVAILABLE, got %s",
			seats[0].Status,
		)
	}
}

func TestConcurrentBooking(t *testing.T) {

	seatRepo := repository.NewInMemorySeatRepository()

	bookingRepo := repository.NewInMemoryBookingRepository()

	payment := &service.UPIProvider{}

	ticketService := service.NewTicketService()

	bookingService := service.NewBookingService(
		seatRepo,
		bookingRepo,
		payment,
		ticketService,
	)

	err := seatRepo.AddSeat(
		&models.ShowSeat{
			ShowID: "SHOW1",
			SeatID: "A1",
			Status: models.Available,
			Price:  25000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	var successCount atomic.Int32

	const users = 100

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(userID int) {
			defer wg.Done()

			_, err := bookingService.CreateBooking(
				context.Background(),
				fmt.Sprintf("USER-%d", userID),
				"SHOW1",
				[]string{"A1"},
			)
			if err == nil {
				successCount.Add(1)
			}

		}(i)

	}

	wg.Wait()

	if successCount.Load() != 1 {
		t.Fatalf(
			"expected exactly one success booking got %d",
			successCount.Load(),
		)
	}

}


func TestCancelBooking(t *testing.T) {

	seatRepo := repository.NewInMemorySeatRepository()

	bookingRepo := repository.NewInMemoryBookingRepository()

	payment := &service.UPIProvider{}

	ticketService := service.NewTicketService()

	bookingService := service.NewBookingService(
		seatRepo,
		bookingRepo,
		payment,
		ticketService,
	)

	err := seatRepo.AddSeat(
		&models.ShowSeat{
			ShowID: "SHOW1",
			SeatID: "A1",
			Status: models.Available,
			Price: 25000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	booking, err := bookingService.CreateBooking(
		context.Background(),
		"USER1",
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	_, err = bookingService.ConfirmBooking(
		context.Background(),
		booking.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	err = bookingService.CancelBooking(
		context.Background(),
		booking.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	finalBooking, err := bookingService.GetBooking(
		booking.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	if finalBooking.Status != models.Cancelled {
		t.Fatalf(
			"expected CANCELLED, got %s",
			finalBooking.Status,
		)
	}

	seats, err := seatRepo.GetSeats(
		"SHOW1",
		[]string{"A1"},
	)

	if err != nil {
		t.Fatal(err)
	}

	if seats[0].Status != models.Available {
		t.Fatalf(
			"expected seat AVAILABLE, got %s",
			seats[0].Status,
		)
	}
}

