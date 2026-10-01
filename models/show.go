package models

import "time"

type ShowStatus string

const (
	ShowScheduled 	ShowStatus = "SCHEDULED"
	ShowCancelled 	ShowStatus = "CANCELLED"
	ShowCompleted 	ShowStatus = "COMPLETED"
)

type Show struct {
	ID 			string
	MovieID		string
	ScreenID	string
	StartTime 	time.Time
	EndTime		time.Time
	Status		ShowStatus
}