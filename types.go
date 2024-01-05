package main

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Account struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	PasswordHash []byte    `json:"-"`
	Balance      uint64    `json:"balance"`
	AccNumber    uint64    `json:"accNumber"`
	CreatedAt    time.Time `json:"createdAt"`
}

type LoginRequest struct {
	AccNumber uint64 `json:"accNumber"`
	Password  string `json:"password"`
}

type TransferRequest struct {
	Amount    uint64 `json:"amount"`
	Recipient uint64 `json:"recipient"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func NewAccount(fname, lname string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName:    fname,
		LastName:     lname,
		PasswordHash: hash,
		Balance:      10000,
		AccNumber:    GenRandNum(),
	}, nil
}
