package main

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/EngZakariaReda/go-rpc-chat/shared"
)

type ChatServer struct {
	mu          sync.RWMutex
	clients     map[string]net.Conn
	messages    []shared.Message
	broadcast   chan shared.BroadcastMessage
	messageSubs map[string]chan string
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:     make(map[string]net.Conn),
		messages:    make([]shared.Message, 0),
		broadcast:   make(chan shared.BroadcastMessage, 100),
		messageSubs: make(map[string]chan string),
	}
}

func (cs *ChatServer) Join(userID string, reply *string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.clients[userID]; exists {
		return fmt.Errorf("user %s already exists", userID)
	}

	cs.clients[userID] = nil
	cs.messageSubs[userID] = make(chan string, 100)

	joinMsg := shared.Message{
		UserID:  userID,
		Content: fmt.Sprintf("User %s joined", userID),
		Type:    "join",
	}

	cs.messages = append(cs.messages, joinMsg)

	go func() {
		cs.broadcast <- shared.BroadcastMessage{
			Message: joinMsg,
			Exclude: userID,
		}
	}()

	*reply = fmt.Sprintf("Welcome %s! You have joined the chat.", userID)
	fmt.Printf("[SERVER] User %s joined\n", userID)

	return nil
}

func (cs *ChatServer) SendMessage(msg shared.Message, reply *bool) error {
	msg.Type = "message"
	
	cs.mu.Lock()
	cs.messages = append(cs.messages, msg)
	cs.mu.Unlock()

	go func() {
		cs.broadcast <- shared.BroadcastMessage{
			Message: msg,
			Exclude: msg.UserID,
		}
	}()

	cs.mu.RLock()
	if ch, ok := cs.messageSubs[msg.UserID]; ok {
		select {
		case ch <- fmt.Sprintf("[You] %s", msg.Content):
		default:
		}
	}
	cs.mu.RUnlock()

	fmt.Printf("[MESSAGE] %s: %s\n", msg.UserID, msg.Content)
	*reply = true
	return nil
}

func (cs *ChatServer) GetHistory(userID string, reply *[]Message) error {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	*reply = cs.messages
	fmt.Printf("[HISTORY] Sent history to %s (%d messages)\n", userID, len(cs.messages))
	return nil
}

func (cs *ChatServer) Leave(userID string, reply *bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if conn, exists := cs.clients[userID]; exists {
		if conn != nil {
			conn.Close()
		}
		
		delete(cs.clients, userID)
		
		if ch, ok := cs.messageSubs[userID]; ok {
			close(ch)
			delete(cs.messageSubs, userID)
		}

		leaveMsg := shared.Message{
			UserID:  userID,
			Content: fmt.Sprintf("User %s left", userID),
			Type:    "leave",
		}

		cs.messages = append(cs.messages, leaveMsg)

		go func() {
			cs.broadcast <- shared.BroadcastMessage{
				Message: leaveMsg,
				Exclude: userID,
			}
		}()

		fmt.Printf("[SERVER] User %s left\n", userID)
	}

	*reply = true
	return nil
}

func (cs *ChatServer) Listen(userID string, reply *bool) error {
	cs.mu.Lock()
	if _, exists := cs.clients[userID]; exists {
		fmt.Printf("[LISTEN] %s started listening\n", userID)
	}
	cs.mu.Unlock()

	*reply = true
	return nil
}

func (cs *ChatServer) StartBroadcaster() {
	fmt.Println("[BROADCASTER] Starting broadcaster...")
	for bm := range cs.broadcast {
		cs.mu.RLock()
		for userID, ch := range cs.messageSubs {
			if userID == bm.Exclude {
				continue
			}
			
			formattedMsg := fmt.Sprintf("[%s] %s", bm.Message.UserID, bm.Message.Content)
			
			select {
			case ch <- formattedMsg:
			default:
				fmt.Printf("[WARNING] Channel full for user %s\n", userID)
			}
		}
		cs.mu.RUnlock()
		
		time.Sleep(10 * time.Millisecond)
	}
}

func StartServer(port string) error {
	chatServer := NewChatServer()
	
	go chatServer.StartBroadcaster()
	
	rpc.Register(chatServer)
	
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	
	fmt.Printf("[SERVER] Listening on %s\n", port)
	fmt.Println("[SERVER] Ready to accept connections...")
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[ERROR] Failed to accept connection: %v\n", err)
			continue
		}
		
		go func() {
			fmt.Printf("[CONNECTION] New client connected from %s\n", conn.RemoteAddr())
			rpc.ServeConn(conn)
		}()
	}
}