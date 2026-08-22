// Prints a bcrypt hash for seeding the admin user into the users table.
//
//	go run ./cmd/hashpw 'your-password'
package main

import (
	"fmt"
	"log"
	"os"

	"portfolio/internal/auth"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: hashpw <password>")
	}

	hash, err := auth.HashPassword(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Println(hash)
}
