package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aether-project/synaptic-core/models"
	"github.com/go-redis/redis/v8"
)

// RedisService handles Redis operations for session caching
type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisService creates a new Redis service instance
func NewRedisService() (*RedisService, error) {
	// Support both REDIS_URL and REDIS_ADDR environment variables
	redisURL := os.Getenv("REDIS_URL")
	redisAddr := os.Getenv("REDIS_ADDR")

	if redisURL == "" && redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	var opt *redis.Options
	var err error

	if redisURL != "" {
		// Parse Redis URL if provided
		opt, err = redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
	} else {
		// Use Redis address directly
		opt = &redis.Options{
			Addr: redisAddr,
		}
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Redis connected successfully: %s", pong)

	return &RedisService{
		client: client,
		ctx:    ctx,
	}, nil
}

// CreateSession creates a new session in Redis
func (r *RedisService) CreateSession(sessionID string) error {
	sessionData := &models.SessionData{
		SessionID:    sessionID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		MessageCount: 0,
		ChatHistory:  []models.MessageHistory{},
		Metadata:     make(map[string]interface{}),
	}

	return r.SetSession(sessionID, sessionData)
}

// GetSession retrieves session data from Redis
func (r *RedisService) GetSession(sessionID string) (*models.SessionData, error) {
	key := r.getSessionKey(sessionID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var sessionData models.SessionData
	err = json.Unmarshal([]byte(val), &sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &sessionData, nil
}

// SetSession stores session data in Redis
func (r *RedisService) SetSession(sessionID string, data *models.SessionData) error {
	key := r.getSessionKey(sessionID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Set with 24 hour expiration
	err = r.client.Set(r.ctx, key, jsonData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	return nil
}

// AddMessage adds a message to the session's chat history
func (r *RedisService) AddMessage(sessionID string, message models.MessageHistory) error {
	sessionData, err := r.GetSession(sessionID)
	if err != nil {
		// If session doesn't exist, create it
		if err := r.CreateSession(sessionID); err != nil {
			return err
		}
		sessionData, err = r.GetSession(sessionID)
		if err != nil {
			return err
		}
	}

	// Add message to history
	sessionData.ChatHistory = append(sessionData.ChatHistory, message)
	sessionData.MessageCount++
	sessionData.LastActivity = time.Now()

	// Keep only last 100 messages to prevent memory issues
	if len(sessionData.ChatHistory) > 100 {
		sessionData.ChatHistory = sessionData.ChatHistory[len(sessionData.ChatHistory)-100:]
	}

	return r.SetSession(sessionID, sessionData)
}

// SetDataSummary stores processed dataset information
func (r *RedisService) SetDataSummary(sessionID string, summary *models.DataSummary) error {
	sessionData, err := r.GetSession(sessionID)
	if err != nil {
		if err := r.CreateSession(sessionID); err != nil {
			return err
		}
		sessionData, err = r.GetSession(sessionID)
		if err != nil {
			return err
		}
	}

	sessionData.DataSummary = summary
	sessionData.LastActivity = time.Now()

	return r.SetSession(sessionID, sessionData)
}

// GetDataSummary retrieves dataset summary for a session
func (r *RedisService) GetDataSummary(sessionID string) (*models.DataSummary, error) {
	sessionData, err := r.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return sessionData.DataSummary, nil
}

// DeleteSession removes a session from Redis
func (r *RedisService) DeleteSession(sessionID string) error {
	key := r.getSessionKey(sessionID)
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// UpdateLastActivity updates the last activity timestamp
func (r *RedisService) UpdateLastActivity(sessionID string) error {
	sessionData, err := r.GetSession(sessionID)
	if err != nil {
		return err
	}

	sessionData.LastActivity = time.Now()
	return r.SetSession(sessionID, sessionData)
}

// GetAllSessions returns all active session IDs (for debugging)
func (r *RedisService) GetAllSessions() ([]string, error) {
	pattern := r.getSessionKey("*")
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	// Extract session IDs from keys
	sessions := make([]string, len(keys))
	for i, key := range keys {
		// Remove "session:" prefix
		sessions[i] = key[8:] // len("session:") = 8
	}

	return sessions, nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// getSessionKey generates the Redis key for a session
func (r *RedisService) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// Health check for Redis service
func (r *RedisService) HealthCheck() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}
