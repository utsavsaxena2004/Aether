package models

import (
	"time"
)

// SessionData represents cached session information
type SessionData struct {
	SessionID    string                 `json:"sessionId"`
	CreatedAt    time.Time              `json:"createdAt"`
	LastActivity time.Time              `json:"lastActivity"`
	MessageCount int                    `json:"messageCount"`
	DataSummary  *DataSummary           `json:"dataSummary,omitempty"`
	ChatHistory  []MessageHistory       `json:"chatHistory"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DataSummary represents processed dataset information
type DataSummary struct {
	FileName    string                 `json:"fileName"`
	RowCount    int                    `json:"rowCount"`
	ColumnCount int                    `json:"columnCount"`
	Columns     []ColumnInfo           `json:"columns"`
	Stats       map[string]interface{} `json:"stats"`
	UploadedAt  time.Time              `json:"uploadedAt"`
}

// ColumnInfo represents information about a dataset column
type ColumnInfo struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Samples  []string    `json:"samples"`
	NullRate float64     `json:"nullRate"`
	Stats    interface{} `json:"stats,omitempty"`
}

// MessageHistory represents a simplified message for caching
type MessageHistory struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}
