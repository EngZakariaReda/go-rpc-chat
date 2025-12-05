package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	port := flag.String("port", ":1234", "Port to listen on (e.g., :1234)")
	flag.Parse()

	fmt.Println("=== RPC Chat Server with Broadcasting ===")
	fmt.Println("Features:")
	fmt.Println("- Real-time message broadcasting")
	fmt.Println("- User join/leave notifications")
	fmt.Println("- No self-echo for messages")
	fmt.Println("- Thread-safe with mutex synchronization")
	fmt.Println("- Full chat history for new clients")
	fmt.Println("=========================================")

	fmt.Printf("\nStarting server on port %s...\n", *port)
	if err := StartServer(*port); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}