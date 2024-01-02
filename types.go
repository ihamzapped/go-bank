package main

import (
	"time"
)

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Balance   uint64    `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// tb rem
func NewAccount(fname, lname string) *Account {
	return &Account{
		FirstName: fname,
		LastName:  lname,
	}
}
