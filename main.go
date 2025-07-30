package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/server"
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

	// Create SSH server
	sshServer := server.NewSSHServer(cfg, db)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", cfg.Server.Port, err)
	}
	defer listener.Close()

	log.Printf("Searchlight BBS Server listening on port %d", cfg.Server.Port)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
			go sshServer.HandleConnection(conn)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")
}
