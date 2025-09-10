package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aether-project/synaptic-core/handlers"
	"github.com/aether-project/synaptic-core/models"
	"github.com/aether-project/synaptic-core/services"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// healthHandler handles the health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

func main() {
	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create WebSocket hub
	hub := models.NewHub()
	go hub.Run()

	// Initialize Redis service
	redisService, err := services.NewRedisService()
	if err != nil {
		log.Printf("⚠️ Failed to initialize Redis service: %v", err)
		// Continue without Redis (optional)
	} else {
		log.Printf("✅ Redis service initialized successfully")
		// Perform health check
		if err := redisService.HealthCheck(); err != nil {
			log.Printf("⚠️ Redis health check failed: %v", err)
		} else {
			log.Printf("✅ Redis health check passed")
		}
	}

	// Initialize AI service
	aiService, err := services.NewAIService()
	if err != nil {
		log.Printf("⚠️ Failed to initialize AI service: %v (continuing without AI)", err)
		// For development, we can continue without AI
		// In production, you might want to fail here
	}

	// Initialize data processor with Redis
	dataProcessor := services.NewDataProcessor(redisService)

	// Create handlers with Redis
	webSocketHandler := handlers.NewWebSocketHandler(hub, aiService, redisService)
	uploadHandler := handlers.NewUploadHandler(dataProcessor, redisService)

	// Create router
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(loggingMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", webSocketHandler.HandleWebSocket)

	// Upload endpoints
	router.HandleFunc("/upload", uploadHandler.HandleUpload).Methods("POST")
	router.HandleFunc("/data-summary", uploadHandler.GetDataSummary).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		aiStatus := "disconnected"
		if aiService != nil {
			if err := aiService.ValidateConnection(); err == nil {
				aiStatus = "connected"
			}
		}

		redisStatus := "disconnected"
		if redisService != nil {
			if err := redisService.HealthCheck(); err == nil {
				redisStatus = "connected"
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":           "Aether Synaptic Core API",
			"version":           "1.0.0",
			"status":            "running",
			"connected_clients": hub.GetClientCount(),
			"ai_status":         aiStatus,
			"redis_status":      redisStatus,
			"ai_model":          "gemini-1.5-pro",
		})
	}).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Printf("🚀 Aether Synaptic Core starting on port %s", port)
	log.Printf("📡 Health check: http://localhost:%s/health", port)
	log.Printf("🔌 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("🌐 API endpoint: http://localhost:%s/", port)

	// Start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
