package main

import (
	"fmt"
	"log"

	"bbs/internal/config"
	"bbs/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Load users from seed data
	fmt.Println("Loading users from seed data...")
	err = db.LoadUsersFromSeed()
	if err != nil {
		fmt.Printf("Error loading users from seed data: %v\n", err)
	} else {
		fmt.Println("Successfully loaded users from seed data")
	}

	// Load bulletins from seed data
	fmt.Println("Loading bulletins from seed data...")
	err = db.LoadBulletinsFromSeed()
	if err != nil {
		fmt.Printf("Error loading bulletins from seed data: %v\n", err)
	} else {
		fmt.Println("Successfully loaded bulletins from seed data")
	}

	// Load topics from seed data
	fmt.Println("Loading topics from seed data...")
	err = db.LoadTopicsFromSeed()
	if err != nil {
		fmt.Printf("Error loading topics from seed data: %v\n", err)
	} else {
		fmt.Println("Successfully loaded topics from seed data")
	}

	fmt.Println("\nDatabase setup complete!")
	fmt.Println("You can now run the BBS server with: go run main.go")
	fmt.Println("Connect via SSH: ssh -p 2323 sysop@localhost (password: password)")
	fmt.Println("Or connect as test user: ssh -p 2323 test@localhost (password: test)")
}
