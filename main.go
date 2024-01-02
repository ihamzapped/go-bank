package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.InitDB(); err != nil {
		log.Fatal(err)
	}

	app := NewApiServer(":8000", store)
	app.Run()
}
