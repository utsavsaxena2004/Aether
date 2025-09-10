package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aether-project/synaptic-core/models"
	"github.com/aether-project/synaptic-core/services"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	processor *services.DataProcessor
	redis     *services.RedisService
}

// NewUploadHandler creates a new upload handler with Redis
func NewUploadHandler(processor *services.DataProcessor, redis *services.RedisService) *UploadHandler {
	return &UploadHandler{
		processor: processor,
		redis:     redis,
	}
}

// UploadResponse represents the response from file upload
type UploadResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Summary   *models.DataSummary `json:"summary,omitempty"`
	SessionID string              `json:"sessionId,omitempty"`
	Error     string              `json:"error,omitempty"`
}

// HandleUpload processes CSV file uploads
func (uh *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "Only POST method allowed",
		})
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50 MB max
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "Failed to parse multipart form: " + err.Error(),
		})
		return
	}

	// Get session ID from form or header
	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		sessionID = r.Header.Get("X-Session-ID")
	}

	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "Session ID is required",
		})
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "No file provided or invalid file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Validate file type
	if !uh.isValidCSVFile(header.Filename) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "Invalid file type. Only CSV files are allowed",
		})
		return
	}

	// Check file size (50MB limit)
	if header.Size > 50<<20 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "File too large. Maximum size is 50MB",
		})
		return
	}

	log.Printf("📁 Processing file upload: %s (%.2f KB) for session %s",
		header.Filename, float64(header.Size)/1024, sessionID)

	// Process the CSV file
	summary, err := uh.processor.ProcessCSV(file, header, sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Error:   "Failed to process CSV file: " + err.Error(),
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully processed %s: %d rows, %d columns",
			header.Filename, summary.RowCount, summary.ColumnCount),
		Summary:   summary,
		SessionID: sessionID,
	})

	log.Printf("✅ File upload completed successfully for session %s", sessionID)
}

// GetDataSummary retrieves the data summary for a session
func (uh *UploadHandler) GetDataSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Only GET method allowed",
		})
		return
	}

	// Get session ID from query parameter or header
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = r.Header.Get("X-Session-ID")
	}

	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Session ID is required",
		})
		return
	}

	// Get data summary from Redis
	if uh.redis == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Redis service not available",
		})
		return
	}

	summary, err := uh.redis.GetDataSummary(sessionID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No data found for session: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}

// isValidCSVFile checks if the file has a valid CSV extension
func (uh *UploadHandler) isValidCSVFile(filename string) bool {
	lowerName := strings.ToLower(filename)
	return strings.HasSuffix(lowerName, ".csv") || strings.HasSuffix(lowerName, ".txt")
}
