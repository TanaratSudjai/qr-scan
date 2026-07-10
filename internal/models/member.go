package models

import "time"

type Member struct {
	ID             int       `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	FullName       string    `json:"fullName"`
	Address        string    `json:"address"`
	Province       string    `json:"province"`
	PostalCode     string    `json:"postalCode"`
	PhoneNumber    string    `json:"phoneNumber"`
	Email          string    `json:"email"`
	Organization   string    `json:"organization"`
	Position       string    `json:"position"`
	Responsibility string    `json:"responsibility"`
	Expectation    string    `json:"expectation"`
	CountCheckin   int       `json:"countCheckin"`
}
