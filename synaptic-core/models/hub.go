package models

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	// WebSocket connection
	Conn *websocket.Conn

	// Unique session ID for this client
	SessionID string

	// Buffered channel of outbound messages
	Send chan *Message

	// The hub this client is connected to
	Hub *Hub

	// Client metadata
	UserAgent string
	IPAddress string
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan *Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the hub and handles client registration, unregistration, and broadcasting
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			h.mutex.Unlock()
			
			log.Printf("📱 Client connected: %s (Total: %d)", client.SessionID, len(h.Clients))
			
			// Send welcome message
			welcomeMsg := CreateChatMessage(
				"Welcome to Aether! I'm your AI assistant ready to help you analyze your data.",
				AuthorSystem,
				client.SessionID,
			)
			
			select {
			case client.Send <- welcomeMsg:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("📱 Client disconnected: %s (Total: %d)", client.SessionID, len(h.Clients))
			}
			h.mutex.Unlock()

		case message := <-h.Broadcast:
			h.mutex.RLock()
			// Broadcast to all clients (or implement per-session targeting)
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToSession sends a message to all clients with a specific session ID
func (h *Hub) BroadcastToSession(message *Message, sessionID string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	for client := range h.Clients {
		if client.SessionID == sessionID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.Clients)
}