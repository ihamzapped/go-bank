package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {
	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal("Failed to connect db:\n", err)
	}

	if err := store.InitDB(); err != nil {
		log.Fatal("Failed to init db:\n", err)
	}

	app := NewApiServer(":8000", store)
	app.Run()
}

// Utils

// Generates a unique 10 digit random number
func GenRandNum() uint64 {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	min := 1000000000
	max := 9999999999

	return uint64(rand.Intn(max-min+1) + min)
}
