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

	// Create default sysop user
	sysopUser := &database.User{
		Username:    "sysop",
		Password:    "password", // In production, use proper password hashing
		RealName:    cfg.BBS.SysopName,
		Email:       "sysop@localhost",
		AccessLevel: 255, // Maximum access level
		IsActive:    true,
	}

	err = db.CreateUser(sysopUser)
	if err != nil {
		fmt.Printf("Note: Sysop user may already exist: %v\n", err)
	} else {
		fmt.Println("Created default sysop user (username: sysop, password: password)")
	}

	// Create a test user
	testUser := &database.User{
		Username:    "test",
		Password:    "test",
		RealName:    "Test User",
		Email:       "test@localhost",
		AccessLevel: 10,
		IsActive:    true,
	}

	err = db.CreateUser(testUser)
	if err != nil {
		fmt.Printf("Note: Test user may already exist: %v\n", err)
	} else {
		fmt.Println("Created test user (username: test, password: test)")
	}

	// Create some sample bulletins
	bulletins := []database.Bulletin{
		{
			Title:  "Welcome to Searchlight BBS!",
			Body:   "This is a recreation of the classic Searchlight BBS software.\n\nFeatures:\n- SSH connectivity\n- Multi-user support\n- Message areas\n- File areas\n- Configurable menus\n\nEnjoy your stay!",
			Author: "Sysop",
		},
		{
			Title:  "System Information",
			Body:   "This BBS is running on Go with SQLite backend.\n\nTechnical details:\n- Concurrent connections via goroutines\n- SSH terminal interface\n- Flexible menu system\n- YAML configuration",
			Author: "Sysop",
		},
	}

	for _, bulletin := range bulletins {
		err = db.CreateBulletin(&bulletin)
		if err != nil {
			fmt.Printf("Error creating bulletin '%s': %v\n", bulletin.Title, err)
		} else {
			fmt.Printf("Created bulletin: %s\n", bulletin.Title)
		}
	}

	fmt.Println("\nDatabase setup complete!")
	fmt.Println("You can now run the BBS server with: go run main.go")
	fmt.Println("Connect via SSH: ssh -p 2323 sysop@localhost")
}
