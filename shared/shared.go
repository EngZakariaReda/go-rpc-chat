package shared

type Message struct {
	UserID  string
	Content string
	Type    string
}

type BroadcastMessage struct {
	Message Message
	Exclude string
}

type ChatService interface {
	Join(userID string, reply *string) error
	SendMessage(msg Message, reply *bool) error
	GetHistory(userID string, reply *[]Message) error
	Leave(userID string, reply *bool) error
	Listen(userID string, reply *bool) error
}