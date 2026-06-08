package main

import (
	"context"
	"flag"
	"fmt"
	"keyclub-api/auth"
	"keyclub-api/internal"
	"log"
	"os"
)

func main() {
	email := flag.String("email", "", "user email (required)")
	first := flag.String("first", "", "first name (required)")
	last := flag.String("last", "", "last name (required)")
	role := flag.String("role", "member", "user role (member, leader, officer)")
	flag.Parse()

	if *email == "" || *first == "" || *last == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./internal/cmd/adduser --email <email> --first <first> --last <last> [--role member]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := internal.LoadConfig()
	db, err := internal.LoadDatabase(config.DBConfig)
	if err != nil {
		log.Fatalf("failed to load database: %v", err)
	}
	defer db.Close()

	user, err := auth.CreateUser(context.Background(), db, *email, *first, *last, *role)
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	fmt.Printf("created user %s (%s %s, %s) with role %s\n",
		user.ID, user.FirstName, user.LastName, user.Email, user.Role)
}
