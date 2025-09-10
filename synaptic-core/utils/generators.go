package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateSessionID generates a unique session ID
func GenerateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "session-" + hex.EncodeToString(bytes)
}

// GenerateMessageID generates a unique message ID
func GenerateMessageID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	timestamp := time.Now().Unix()
	return hex.EncodeToString(bytes) + "-" + time.Unix(timestamp, 0).Format("20060102150405")
}