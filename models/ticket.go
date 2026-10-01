package models

type Ticket struct {
	ID 			string
	BookingID	string
	ShowID		string
	SeatIDs 	[]string
	QRCode		string
}