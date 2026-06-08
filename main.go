package main

import (
	"log"

	"keyclub-api/internal"
	"keyclub-api/internal/logger"
)

func main() {
	if err := logger.Setup("logs/app.log"); err != nil {
		log.Fatal(err)
	}

	app := internal.NewApp()
	defer app.DB.Close()

	if err := app.Start(":8000"); err != nil {
		log.Fatal(err)
	}
}
