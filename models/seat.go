package models

import "time"

type SeatCategory string

const (
	Regular 	SeatCategory = "REGULAR"
	Premium 	SeatCategory = "PREMIUM"
	Recliner 	SeatCategory = "RECLINER"
)

type SeatStatus string

const (
	Available 	SeatStatus = "AVAILABLE"
	Held		SeatStatus = "HELD"
	Booked		SeatStatus = "BOOKED"
)


type Seat struct{
	ID 			string
	Row			string
	Number 		string
	category 	SeatCategory
}

type ShowSeat struct {
	ShowID 		string
	SeatID		string
	Status      SeatStatus
	BookingID	string
	ExpiresAt	time.Time
	Price 		int64	
}