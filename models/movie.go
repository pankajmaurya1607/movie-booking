package models

import "time"


type Movie struct {
	ID			string
	Title 		string
	Description string
	Language 	string
	Genre		string
	Duration	time.Duration
}