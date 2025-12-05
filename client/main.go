package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	serverAddr := flag.String("server", "localhost:1234", "Server address (host:port)")
	userID := flag.String("user", "", "Your username (required)")
	flag.Parse()

	if *userID == "" {
		fmt.Println("Error: You must specify a username with -user flag")
		fmt.Println("Usage: go run . -user=YourName -server=localhost:1234")
		os.Exit(1)
	}

	fmt.Println("=== RPC Chat Client ===")
	fmt.Printf("Connecting as: %s\n", *userID)
	fmt.Printf("Server: %s\n", *serverAddr)
	fmt.Println("=======================")

	client := NewChatClient(*serverAddr, *userID)
	
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer client.Disconnect()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nDisconnecting...")
		client.Disconnect()
		os.Exit(0)
	}()

	client.StartMessageReceiver()
	client.StartCLI()

	fmt.Println("Goodbye!")
}