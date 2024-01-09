package main

import (
	"database/sql"
	"fmt"
	"os"

	// "fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) (*Account, error)
	DeleteAccount(int) error
	UpdateAccountBal(int, uint64) error
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(uint64) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) InitDB() error {
	return s.CreateAccountTable()
}

func NewPostgresStore() (*PostgresStore, error) {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) CreateAccountTable() error {
	q := `CREATE TABLE IF NOT EXISTS account (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(250) NOT NULL,
    last_name VARCHAR(250) NOT NULL,
    password_hash BYTEA NOT NULL,
    balance BIGINT,
    acc_number BIGINT NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);`

	_, err := s.db.Exec(q)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) (*Account, error) {
	q := `
		INSERT INTO account (first_name, last_name, password_hash, acc_number, balance)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, first_name, last_name, acc_number, balance, created_at
	`

	a := &Account{}
	res := s.db.QueryRow(q, acc.FirstName, acc.LastName, acc.PasswordHash, acc.AccNumber, acc.Balance)
	err := res.Scan(&a.ID, &a.FirstName, &a.LastName, &a.AccNumber, &a.Balance, &a.CreatedAt)

	if err != nil {
		return &Account{}, err
	}

	return a, nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	q := `DELETE FROM account WHERE id = $1;`
	_, err := s.db.Exec(q, id)

	return err
}

func (s *PostgresStore) UpdateAccountBal(id int, bal uint64) error {
	q := `
		UPDATE account
		SET balance = $2
		WHERE id = $1;
	`
	_, err := s.db.Exec(q, id, bal)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {

	q := `
		SELECT id, first_name, last_name, balance, acc_number, created_at
		FROM account WHERE id = $1;
	`
	a := &Account{}
	res := s.db.QueryRow(q, id)
	err := res.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Balance, &a.AccNumber, &a.CreatedAt)

	if err != nil {
		return &Account{}, err
	}

	return a, nil

}

func (s *PostgresStore) GetAccountByNumber(acc uint64) (*Account, error) {

	q := `
		SELECT * FROM account WHERE acc_number = $1;
	`
	a := &Account{}
	res := s.db.QueryRow(q, acc)
	err := res.Scan(&a.ID, &a.FirstName, &a.LastName, &a.PasswordHash, &a.Balance, &a.AccNumber, &a.CreatedAt)

	if err != nil {
		return &Account{}, err
	}

	return a, nil

}
