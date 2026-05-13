package main

import (
	"fmt"
	"keyclub-api/auth"
	"keyclub-api/internal"
)

func main() {
	fmt.Println("Hello, World!")
	app := internal.NewApp()

	if err := auth.SendInvite("nikolay.li2008@gmail.com", "Nikolay", 0, app.Config); err != nil {
		fmt.Println("Error sending invite:", err)
	}
}
