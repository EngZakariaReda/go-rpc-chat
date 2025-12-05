package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EngZakariaReda/go-rpc-chat/shared"
)

type ChatClient struct {
	userID     string
	serverAddr string
	client     *rpc.Client
	mu         sync.RWMutex
	connected  bool
	messages   chan string
	stopChan   chan bool
}

func NewChatClient(serverAddr, userID string) *ChatClient {
	return &ChatClient{
		userID:     userID,
		serverAddr: serverAddr,
		messages:   make(chan string, 100),
		stopChan:   make(chan bool),
		connected:  false,
	}
}

func (cc *ChatClient) Connect() error {
	client, err := rpc.Dial("tcp", cc.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server at %s: %v", cc.serverAddr, err)
	}

	cc.mu.Lock()
	cc.client = client
	cc.connected = true
	cc.mu.Unlock()

	var joinReply string
	err = cc.client.Call("ChatServer.Join", cc.userID, &joinReply)
	if err != nil {
		cc.mu.Lock()
		cc.client.Close()
		cc.connected = false
		cc.mu.Unlock()
		return fmt.Errorf("failed to join chat: %v", err)
	}

	fmt.Println(joinReply)

	var history []shared.Message
	err = cc.client.Call("ChatServer.GetHistory", cc.userID, &history)
	if err != nil {
		fmt.Printf("Note: Could not retrieve chat history: %v\n", err)
	} else {
		fmt.Println("\n=== Chat History ===")
		for _, msg := range history {
			switch msg.Type {
			case "join":
				fmt.Printf("--> %s\n", msg.Content)
			case "leave":
				fmt.Printf("<-- %s\n", msg.Content)
			default:
				fmt.Printf("[%s] %s\n", msg.UserID, msg.Content)
			}
		}
		fmt.Println("====================\n")
	}

	var listenReply bool
	err = cc.client.Call("ChatServer.Listen", cc.userID, &listenReply)
	if err != nil {
		fmt.Printf("Warning: Could not start listener: %v\n", err)
	}

	return nil
}

func (cc *ChatClient) StartMessageReceiver() {
	go func() {
		for {
			select {
			case <-cc.stopChan:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (cc *ChatClient) DisplayMessages() {
	go func() {
		for msg := range cc.messages {
			fmt.Println(msg)
			fmt.Print("> ")
		}
	}()
}

func (cc *ChatClient) SendMessage(content string) error {
	cc.mu.RLock()
	if !cc.connected || cc.client == nil {
		cc.mu.RUnlock()
		return fmt.Errorf("not connected to server")
	}
	client := cc.client
	cc.mu.RUnlock()

	msg := shared.Message{
		UserID:  cc.userID,
		Content: content,
	}

	var reply bool
	err := client.Call("ChatServer.SendMessage", msg, &reply)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

func (cc *ChatClient) Disconnect() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.connected && cc.client != nil {
		var reply bool
		cc.client.Call("ChatServer.Leave", cc.userID, &reply)
		cc.client.Close()
		cc.connected = false
		close(cc.stopChan)
		fmt.Println("Disconnected from server.")
	}
}

func (cc *ChatClient) StartCLI() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\nCommands:")
	fmt.Println("  /quit    - Exit the chat")
	fmt.Println("  /users   - Show online users (simulated)")
	fmt.Println("  /clear   - Clear screen")
	fmt.Println("  Type your message and press Enter to send")
	fmt.Println(strings.Repeat("-", 40))

	cc.DisplayMessages()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "/quit":
			return
		case "/users":
			fmt.Println("Online users feature is simulated in this version")
			cc.mu.RLock()
			fmt.Println("(You are connected as", cc.userID, ")")
			cc.mu.RUnlock()
			continue
		case "/clear":
			fmt.Print("\033[H\033[2J")
			continue
		case "/help":
			fmt.Println("Commands: /quit, /users, /clear, /help")
			continue
		}

		if err := cc.SendMessage(input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}