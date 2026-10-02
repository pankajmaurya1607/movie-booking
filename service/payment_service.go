package service

import (
	"context"
	"errors"
	"movie-booking/models"
)


type PaymentGateway interface {
	Pay(
		ctx context.Context,
		bookingID string,
		amount int64,
	) (*models.Payment, error)
} 

// UPI Payment

type UPIProvider struct {
	ShouldFail bool
}

func (p *UPIProvider) Pay(
	ctx context.Context,
	bookingID string,
	amount int64,
) (*models.Payment, error) {

	if p.ShouldFail {
		return &models.Payment{
			ID:	"UPI-" + bookingID,
			BookingID: bookingID,
			Amount: amount,
			Method: "UPI",
			Status: models.PaymentFailed,
		}, errors.New("payment failed")
	}

	return &models.Payment{
		ID:	"UPI-" + bookingID,
		BookingID: bookingID,
		Amount: amount,
		Method: "UPI",
		Status: models.PaymentSuccess,
	}, nil
}


// CARD Payment


type CardProvider struct {
	ShouldFail bool
}


func (p *CardProvider) Pay(
	ctx context.Context,
	bookingID string,
	amount int64,
) (*models.Payment, error) {

	if p.ShouldFail {
		return &models.Payment{
			ID:	"CARD-" + bookingID,
			BookingID: bookingID,
			Amount: amount,
			Method: "CARD",
			Status: models.PaymentFailed,
		}, errors.New("payment failed")
	}

	return &models.Payment{
		ID:	"CARD-" + bookingID,
		BookingID: bookingID,
		Amount: amount,
		Method: "CARD",
		Status: models.PaymentSuccess,
	}, nil
}