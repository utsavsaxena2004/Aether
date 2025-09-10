package models

import (
	"encoding/json"
	"time"
)

// Message represents a WebSocket message
type Message struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Content     string          `json:"content,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
	Author      string          `json:"author"`
	Timestamp   time.Time       `json:"timestamp"`
	SessionID   string          `json:"sessionId"`
	IsStreaming bool            `json:"isStreaming,omitempty"`
}

// MessageType constants
const (
	MessageTypeChat           = "chat"
	MessageTypeChartSpec      = "chart_spec"
	MessageTypeError          = "error"
	MessageTypeSystem         = "system"
	MessageTypeUserConnect    = "user_connect"
	MessageTypeUserDisconnect = "user_disconnect"
	MessageTypeVisualQuery    = "visual_query"
)

// Author constants
const (
	AuthorUser   = "user"
	AuthorAI     = "ai"
	AuthorSystem = "system"
)

// CreateChatMessage creates a new chat message
func CreateChatMessage(content, author, sessionID string) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeChat,
		Content:   content,
		Author:    author,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
}

// CreateChartMessage creates a new chart specification message
func CreateChartMessage(chartData json.RawMessage, sessionID string) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeChartSpec,
		Data:      chartData,
		Author:    AuthorAI,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
}

// CreateErrorMessage creates a new error message
func CreateErrorMessage(content, sessionID string) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeError,
		Content:   content,
		Author:    AuthorSystem,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
}

// CreateVisualQueryMessage creates a new visual query message
func CreateVisualQueryMessage(content string, data json.RawMessage, sessionID string) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeVisualQuery,
		Content:   content,
		Data:      data,
		Author:    AuthorUser,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// Simple timestamp-based ID for now
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}