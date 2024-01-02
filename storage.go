package main

import (
	"database/sql"
	// "fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) (Account, error)
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccount(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) InitDB() error {
	return s.CreateAccountTable()
}

func NewPostgresStore() (*PostgresStore, error) {
	dsn := "user=admin password=password@331 dbname=fiber host=localhost port=5432 sslmode=disable"

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
    first_name VARCHAR(250),
    last_name VARCHAR(250),
    balance BIGINT,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);`

	_, err := s.db.Exec(q)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) (Account, error) {
	// fmt.Printf("Person: %+v\n", acc)

	q := `
		INSERT INTO account (first_name, last_name, balance)
		VALUES ($1, $2, $3)
		RETURNING id, first_name, last_name, balance, created_at
	`

	res := s.db.QueryRow(q, acc.FirstName, acc.LastName, acc.Balance)

	var a Account

	// Scan the query results into the fields of the custom type
	err := res.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Balance, &a.CreatedAt)
	if err != nil {
		return Account{}, err
	}

	return a, nil
}
func (s *PostgresStore) DeleteAccount(int) error          { return nil }
func (s *PostgresStore) UpdateAccount(*Account) error     { return nil }
func (s *PostgresStore) GetAccount(int) (*Account, error) { return nil, nil }
