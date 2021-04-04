package karen

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/poopmail/canalization/internal/env"
)

// MessageType represents the type of an outgoing karen message
type MessageType string

const (
	MessageTypeDebug   = MessageType("DEBUG")
	MessageTypeInfo    = MessageType("INFO")
	MessageTypeSuccess = MessageType("SUCCESS")
	MessageTypeWarning = MessageType("WARNING")
	MessageTypeError   = MessageType("ERROR")
	MessageTypePanic   = MessageType("PANIC")
)

// Message represents a structured outgoing karen message
type Message struct {
	Type        MessageType `json:"type"`
	Service     string      `json:"service"`
	Topic       string      `json:"topic"`
	Description string      `json:"description"`
}

// Encode encodes a structured message into a Base64 string as required for karen
func (msg Message) Encode() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

// Send sends a structured message to the configured karen instance
func Send(rdb *redis.Client, msg Message) error {
	karenChannel := env.MustString("CANAL_KAREN_REDIS_CHANNEL", "")
	if karenChannel == "" {
		return nil
	}

	encoded, err := msg.Encode()
	if err != nil {
		return err
	}

	return rdb.Publish(context.Background(), karenChannel, encoded).Err()
}
