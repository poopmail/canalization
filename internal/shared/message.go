package shared

import "github.com/bwmarrin/snowflake"

// Message represents an incoming email message
type Message struct {
	ID      snowflake.ID `json:"id"`
	Mailbox string       `json:"mailbox"`
	From    string       `json:"from"`
	Subject string       `json:"subject"`
	Content string       `json:"content"`
	Created int64        `json:"created"`
}

// MessageService represents a service which keeps track of messages
type MessageService interface {
	Count(mailbox string) (int, error)
	Messages(mailbox string, skip, limit int) ([]*Message, error)
	Message(id snowflake.ID) (*Message, error)
	CreateOrReplace(message *Message) error
	Delete(id snowflake.ID) error
	DeleteInMailbox(mailbox string) error
}
