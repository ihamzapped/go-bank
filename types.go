package main

import "math/rand"

type Account struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Number    int64  `json:"number"`
	Balance   int64  `json:"balance"`
}

func NewAccount(fname, lname string) *Account {
	return &Account{
		ID:        rand.Intn(10000),
		FirstName: fname,
		LastName:  lname,
		Number:    int64(rand.Intn(10000)),
	}
}
