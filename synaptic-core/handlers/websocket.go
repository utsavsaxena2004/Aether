package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aether-project/synaptic-core/models"
	"github.com/aether-project/synaptic-core/services"
	"github.com/aether-project/synaptic-core/utils"
	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost during development
		// In production, implement proper origin checking
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	Hub   *models.Hub
	AI    *services.AIService
	Redis *services.RedisService
}

// NewWebSocketHandler creates a new WebSocket handler with Redis
func NewWebSocketHandler(hub *models.Hub, ai *services.AIService, redis *services.RedisService) *WebSocketHandler {
	return &WebSocketHandler{
		Hub:   hub,
		AI:    ai,
		Redis: redis,
	}
}

// HandleWebSocket upgrades HTTP requests to WebSocket connections
func (wsh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}

	// Generate session ID
	sessionID := utils.GenerateSessionID()

	// Create client
	client := &models.Client{
		Conn:      conn,
		SessionID: sessionID,
		Send:      make(chan *models.Message, 256),
		Hub:       wsh.Hub,
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: r.RemoteAddr,
	}

	// Register client with hub
	client.Hub.Register <- client

	// Create session in Redis if available
	if wsh.Redis != nil {
		if err := wsh.Redis.CreateSession(sessionID); err != nil {
			log.Printf("⚠️ Failed to create Redis session: %v", err)
		} else {
			log.Printf("✅ Created Redis session: %s", sessionID)
		}
	}

	// Start goroutines for reading and writing
	go wsh.writePump(client)
	go wsh.readPump(client)
}

// readPump pumps messages from the WebSocket connection to the hub
func (wsh *WebSocketHandler) readPump(client *models.Client) {
	defer func() {
		client.Hub.Unregister <- client
		client.Conn.Close()

		// Update session last activity in Redis if available
		if wsh.Redis != nil {
			if err := wsh.Redis.UpdateLastActivity(client.SessionID); err != nil {
				log.Printf("⚠️ Failed to update Redis session activity: %v", err)
			}
		}
	}()

	// Set read limits and timeouts
	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Update session last activity in Redis if available
		if wsh.Redis != nil {
			if err := wsh.Redis.UpdateLastActivity(client.SessionID); err != nil {
				log.Printf("⚠️ Failed to update Redis session activity: %v", err)
			}
		}

		return nil
	})

	for {
		var message models.Message
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		// Set session ID and timestamp
		message.SessionID = client.SessionID
		message.Timestamp = time.Now()

		// Generate message ID if not provided
		if message.ID == "" {
			message.ID = utils.GenerateMessageID()
		}

		log.Printf("📨 Received message: %s from %s", message.Content, client.SessionID)

		// Store message in Redis if available
		if wsh.Redis != nil {
			redisMessage := models.MessageHistory{
				ID:        message.ID,
				Type:      string(message.Type),
				Content:   message.Content,
				Author:    message.Author,
				Timestamp: message.Timestamp,
			}
			if err := wsh.Redis.AddMessage(client.SessionID, redisMessage); err != nil {
				log.Printf("⚠️ Failed to store message in Redis: %v", err)
			}
		}

		// Process the message (this is where we'll add AI processing later)
		wsh.processMessage(&message, client)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (wsh *WebSocketHandler) writePump(client *models.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("❌ WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage handles incoming messages from clients
func (wsh *WebSocketHandler) processMessage(message *models.Message, client *models.Client) {
	switch message.Type {
	case models.MessageTypeChat:
		// Generate AI response
		var responseContent string

		if wsh.AI != nil {
			// Check if user is asking for a visualization
			if wsh.AI.DetectChartIntent(message.Content) {
				// Generate chart specification
				chartSpec, chartErr := wsh.AI.GenerateChartSpec(message.Content, nil)
				if chartErr != nil {
					log.Printf("⚠️ Chart generation failed: %v", chartErr)
					responseContent = fmt.Sprintf("I'd like to create a visualization for you, but encountered an error: %v", chartErr)
				} else {
					// Send chart specification as a separate message
					chartMessage := models.CreateChartMessage(json.RawMessage(chartSpec), client.SessionID)
					wsh.Hub.BroadcastToSession(chartMessage, client.SessionID)

					responseContent = "I've created a visualization based on your request. You can see the chart above!"
				}
			} else {
				// Generate streaming text response
				streamResp, streamErr := wsh.AI.GenerateResponseStream(message.Content, nil)
				if streamErr != nil {
					log.Printf("⚠️ AI streaming generation failed: %v", streamErr)
					responseContent = fmt.Sprintf("I apologize, but I'm having trouble processing your request right now. Error: %v", streamErr)

					// Send as regular message
					response := models.CreateChatMessage(
						responseContent,
						models.AuthorAI,
						client.SessionID,
					)
					wsh.Hub.BroadcastToSession(response, client.SessionID)
				} else {
					// Stream the response
					wsh.streamAIResponse(streamResp, client)
					return
				}
			}
		} else {
			// Fallback response when AI is not available
			responseContent = fmt.Sprintf("I received your message: \"%s\". AI processing is temporarily unavailable.", message.Content)
		}

		// Only send regular response if we didn't stream
		if responseContent != "" {
			response := models.CreateChatMessage(
				responseContent,
				models.AuthorAI,
				client.SessionID,
			)

			// Send response back to the specific client session
			wsh.Hub.BroadcastToSession(response, client.SessionID)
		}

	case models.MessageTypeVisualQuery:
		// Handle visual query messages
		log.Printf("🔍 Processing visual query from session %s", client.SessionID)

		// Generate AI response for visual query
		if wsh.AI != nil {
			// Generate response based on visual selection and text query
			aiResponse, aiErr := wsh.AI.GenerateVisualQueryResponse(message.Content, message.Data, nil)
			if aiErr != nil {
				log.Printf("⚠️ Visual query AI generation failed: %v", aiErr)
				errorMsg := models.CreateErrorMessage(fmt.Sprintf("Failed to process visual query: %v", aiErr), client.SessionID)
				wsh.Hub.BroadcastToSession(errorMsg, client.SessionID)
				return
			}

			response := models.CreateChatMessage(
				aiResponse,
				models.AuthorAI,
				client.SessionID,
			)

			// Send response back to the specific client session
			wsh.Hub.BroadcastToSession(response, client.SessionID)
		} else {
			// Fallback response when AI is not available
			responseContent := "I received your visual query, but AI processing is temporarily unavailable."
			response := models.CreateChatMessage(
				responseContent,
				models.AuthorAI,
				client.SessionID,
			)
			wsh.Hub.BroadcastToSession(response, client.SessionID)
		}

	default:
		log.Printf("⚠️ Unknown message type: %s", message.Type)
	}
}

// streamAIResponse streams AI response tokens to the client
func (wsh *WebSocketHandler) streamAIResponse(stream *genai.GenerateContentResponseIterator, client *models.Client) {
	// Create a message ID for the streaming response
	messageID := utils.GenerateMessageID()

	// Track the accumulated content
	var accumulatedContent strings.Builder

	// Process the stream
	for {
		resp, err := stream.Next()
		if err != nil {
			// Stream ended or error occurred
			if err.Error() == "EOF" {
				log.Printf("✅ AI response stream completed")
			} else {
				log.Printf("⚠️ AI response stream error: %v", err)
			}
			break
		}

		// Extract text from response
		var responseText strings.Builder
		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					responseText.WriteString(string(txt))
				}
			}
		}

		token := responseText.String()
		if token != "" {
			accumulatedContent.WriteString(token)

			// Create streaming message
			streamingMessage := &models.Message{
				ID:          messageID,
				Type:        models.MessageTypeChat,
				Content:     accumulatedContent.String(),
				Author:      models.AuthorAI,
				Timestamp:   time.Now(),
				SessionID:   client.SessionID,
				IsStreaming: true,
			}

			// Send streaming message to client
			client.Send <- streamingMessage
		}
	}

	// Send final message to indicate stream completion
	finalMessage := &models.Message{
		ID:          messageID,
		Type:        models.MessageTypeChat,
		Content:     accumulatedContent.String(),
		Author:      models.AuthorAI,
		Timestamp:   time.Now(),
		SessionID:   client.SessionID,
		IsStreaming: false,
	}

	// Send final message to client
	client.Send <- finalMessage

	// Store final message in Redis if available
	if wsh.Redis != nil {
		redisMessage := models.MessageHistory{
			ID:        messageID,
			Type:      string(models.MessageTypeChat),
			Content:   accumulatedContent.String(),
			Author:    models.AuthorAI,
			Timestamp: time.Now(),
		}
		if err := wsh.Redis.AddMessage(client.SessionID, redisMessage); err != nil {
			log.Printf("⚠️ Failed to store final message in Redis: %v", err)
		}
	}
}
