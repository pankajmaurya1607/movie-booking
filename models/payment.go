package models

type PaymentStatus string 

const (
	PaymentPending	PaymentStatus = "PENDING"
	PaymentSuccess	PaymentStatus = "SUCCESS"
	PaymentFailed	PaymentStatus = "FAILED"
)


type Payment struct {
	ID			string
	BookingID	string
	Amount 		int64
	Method 		string
	Status 		PaymentStatus
}