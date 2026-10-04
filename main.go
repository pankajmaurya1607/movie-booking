package main

import (
	"context"
	"fmt"
	"movie-booking/models"
	"movie-booking/repository"
	"movie-booking/service"
	"time"
)

func main() {

	/*
		-----------------------------------
		Repositories
		-----------------------------------
	*/

	seatRepo := repository.NewInMemorySeatRepository()

	bookingRepo := repository.NewInMemoryBookingRepository()

	showRepo := repository.NewInMemoryShowRepository()

	/*
		-----------------------------------
		Create Movie
		-----------------------------------
	*/

	movie := &models.Movie{
		ID:          "M101",
		Title:       "Avengers",
		Description: "Superhero Movie",
		Language:    "English",
		Genre:       "Action",
		Duration:    2*time.Hour + 30*time.Minute,
	}

	fmt.Println("Movie: ", movie.Title)

	/*
		-----------------------------------
		Create Show
		-----------------------------------
	*/

	show := &models.Show{
		ID:        "SHOW101",
		MovieID:   movie.ID,
		ScreenID:  "SCREEN101",
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(3 * time.Hour),
		Status:    models.ShowScheduled,
	}

	err := showRepo.Create(show)

	if err != nil {
		panic(err)
	}

	/*
		-----------------------------------
		Create Show Seats
		-----------------------------------
	*/

	seats := []struct {
		id    string
		price int64
	}{
		{"A1", 25000},
		{"A2", 25000},
		{"A3", 25000},
		{"B1", 20000},
		{"B2", 20000},
	}

	for _, s := range seats {

		err := seatRepo.AddSeat(
			&models.ShowSeat{
				ShowID: show.ID,
				SeatID: s.id,
				Status: models.Available,
				Price:  s.price,
			},
		)

		if err != nil {
			panic(err)
		}
	}

	/*
		-----------------------------------
		Payment
		-----------------------------------
	*/

	paymentGateway := &service.UPIProvider{
		ShouldFail: false,
	}

	/*
		-----------------------------------
		Ticket Service
		-----------------------------------
	*/

	ticketService := service.NewTicketService()

	/*
		-----------------------------------
		Booking Service
		-----------------------------------
	*/

	bookingService := service.NewBookingService(
		seatRepo,
		bookingRepo,
		paymentGateway,
		ticketService,
	)

	ctx := context.Background()

	/*
		-----------------------------------
		Create Booking
		-----------------------------------
	*/

	booking, err := bookingService.CreateBooking(
		ctx,
		"USER101",
		show.ID,
		[]string{"A1", "A2"},
	)

	if err != nil {
		fmt.Println("Booking failed:", err)
		return
	}

	fmt.Println(
		"Booking created:",
		booking.ID,
	)

	fmt.Println(
		"Booking status:",
		booking.Status,
	)

	fmt.Println(
		"Total:",
		booking.Total,
		"paise",
	)

	/*
		-----------------------------------
		Confirm Booking
		-----------------------------------
	*/

	ticket, err := bookingService.ConfirmBooking(
		ctx,
		booking.ID,
	)

	if err != nil {
		fmt.Println(
			"Booking confirmation failed:",
			err,
		)
		return
	}

	fmt.Println(
		"Booking confirmed!",
	)

	fmt.Println(
		"Ticket:",
		ticket.ID,
	)

	fmt.Println(
		"QR:",
		ticket.QRCode,
	)

	/*
		-----------------------------------
		Get Booking
		-----------------------------------
	*/

	finalBooking, err := bookingService.GetBooking(
		booking.ID,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(
		"Final status:",
		finalBooking.Status,
	)

}
