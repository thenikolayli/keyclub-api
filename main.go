package main

import (
	"log"

	"keyclub-api/internal"
)

func main() {
	app := internal.NewApp()
	defer app.DB.Close()

	if err := app.Start(":8000"); err != nil {
		log.Fatal(err)
	}
}
