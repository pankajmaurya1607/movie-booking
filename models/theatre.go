package models

type Theatre struct {
	ID 		string
	Name 	string
	City	string
	Address string
	Screens []*Screen
}

type Screen struct {
	ID 			string
	TheatreID	string
	Name 		string
	Seats		[]*Seat
}

