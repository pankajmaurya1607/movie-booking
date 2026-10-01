package models

import "time"

type BookingStatus string

const (
	Pending 	BookingStatus = "PENDING"
	Confirmed	BookingStatus = "CONFIRMED"
	Cancelled   BookingStatus = "CANCELLED"
	Expired 	BookingStatus = "EXPIRED"
	Failed		BookingStatus = "FAILED"
)



type Booking struct {
	ID 			string
	UserID		string
	ShowID		string
	SeatIDs		[]string
	Status  	BookingStatus
	Total  		int64
	CreatedAt	time.Time
	ExpiredAt 	time.Time
}