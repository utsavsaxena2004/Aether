package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aether-project/synaptic-core/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AIService handles AI interactions using Google Gemini
type AIService struct {
	client *genai.Client
	model  *genai.GenerativeModel
	ctx    context.Context
}

// VisualQueryData represents the data structure for visual queries
type VisualQueryData struct {
	Selection struct {
		StartX float64 `json:"startX"`
		StartY float64 `json:"startY"`
		EndX   float64 `json:"endX"`
		EndY   float64 `json:"endY"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"selection"`
	ChartType string `json:"chartType,omitempty"`
}

// NewAIService creates a new AI service instance
func NewAIService() (*AIService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use Gemini 1.5 Pro model
	model := client.GenerativeModel("gemini-1.5-pro")

	// Configure model settings
	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(2048)

	log.Printf("✅ Gemini AI service initialized successfully")

	return &AIService{
		client: client,
		model:  model,
		ctx:    ctx,
	}, nil
}

// GenerateResponse generates an AI response based on user query and context
func (ai *AIService) GenerateResponse(userQuery string, sessionData *models.SessionData) (string, error) {
	// Build context-aware prompt
	prompt := ai.buildPrompt(userQuery, sessionData)

	log.Printf("🤖 Generating AI response for query: %s", userQuery)

	// Generate response
	resp, err := ai.model.GenerateContent(ai.ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates generated")
	}

	// Extract text from response
	var responseText strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText.WriteString(string(txt))
		}
	}

	response := responseText.String()
	if response == "" {
		return "", fmt.Errorf("empty response generated")
	}

	log.Printf("✅ AI response generated successfully (%d characters)", len(response))
	return response, nil
}

// GenerateResponseStream generates an AI response stream based on user query and context
func (ai *AIService) GenerateResponseStream(userQuery string, sessionData *models.SessionData) (*genai.GenerateContentResponseIterator, error) {
	// Build context-aware prompt
	prompt := ai.buildPrompt(userQuery, sessionData)

	log.Printf("🤖 Generating AI response stream for query: %s", userQuery)

	// Generate response stream
	resp := ai.model.GenerateContentStream(ai.ctx, genai.Text(prompt))

	return resp, nil
}

// GenerateChartSpec generates an ECharts JSON specification based on user query and data
func (ai *AIService) GenerateChartSpec(userQuery string, sessionData *models.SessionData) (string, error) {
	if sessionData == nil || sessionData.DataSummary == nil {
		return "", fmt.Errorf("no data available for chart generation")
	}

	// Build chart-specific prompt
	prompt := ai.buildChartPrompt(userQuery, sessionData)

	log.Printf("📈 Generating chart specification for query: %s", userQuery)

	// Generate response
	resp, err := ai.model.GenerateContent(ai.ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate chart specification: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no chart specification candidates generated")
	}

	// Extract text from response
	var responseText strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText.WriteString(string(txt))
		}
	}

	chartSpec := responseText.String()
	if chartSpec == "" {
		return "", fmt.Errorf("empty chart specification generated")
	}

	// Extract JSON from response (remove markdown formatting if present)
	chartSpec = ai.extractJSONFromResponse(chartSpec)

	log.Printf("✅ Chart specification generated successfully")
	return chartSpec, nil
}

// GenerateVisualQueryResponse generates an AI response based on a visual query and selection
func (ai *AIService) GenerateVisualQueryResponse(userQuery string, visualData json.RawMessage, sessionData *models.SessionData) (string, error) {
	// Parse visual query data
	var visualQuery VisualQueryData
	if err := json.Unmarshal(visualData, &visualQuery); err != nil {
		return "", fmt.Errorf("failed to parse visual query data: %w", err)
	}

	// Build visual query-specific prompt
	prompt := ai.buildVisualQueryPrompt(userQuery, visualQuery, sessionData)

	log.Printf("🔍 Generating AI response for visual query: %s", userQuery)

	// Generate response
	resp, err := ai.model.GenerateContent(ai.ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate visual query response: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no visual query response candidates generated")
	}

	// Extract text from response
	var responseText strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText.WriteString(string(txt))
		}
	}

	response := responseText.String()
	if response == "" {
		return "", fmt.Errorf("empty visual query response generated")
	}

	log.Printf("✅ Visual query response generated successfully (%d characters)", len(response))
	return response, nil
}

// DetectChartIntent determines if user query is asking for a visualization
func (ai *AIService) DetectChartIntent(userQuery string) bool {
	// Keywords that suggest chart/visualization intent
	chartKeywords := []string{
		"chart", "graph", "plot", "visualize", "show", "display",
		"bar chart", "line chart", "pie chart", "scatter plot",
		"histogram", "trend", "distribution", "correlation",
		"compare", "comparison", "over time", "breakdown",
	}

	queryLower := strings.ToLower(userQuery)
	for _, keyword := range chartKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}

// buildVisualQueryPrompt constructs a specialized prompt for visual query responses
func (ai *AIService) buildVisualQueryPrompt(userQuery string, visualQuery VisualQueryData, sessionData *models.SessionData) string {
	var prompt strings.Builder

	// System instructions for visual query responses
	prompt.WriteString("You are Aether, an expert data analysis AI with visual intelligence capabilities. ")
	prompt.WriteString("The user has selected an area on a data visualization and is asking a follow-up question. ")
	prompt.WriteString("Provide a detailed analysis of the selected area based on the user's question.\n\n")

	// Data context
	if sessionData != nil && sessionData.DataSummary != nil {
		ds := sessionData.DataSummary
		prompt.WriteString("## Dataset Information:\n")
		prompt.WriteString(fmt.Sprintf("- File: %s (%d rows × %d columns)\n", ds.FileName, ds.RowCount, ds.ColumnCount))

		// Column information with enhanced details
		prompt.WriteString("\n### Available Columns:\n")
		for _, col := range ds.Columns {
			prompt.WriteString(fmt.Sprintf("- **%s** (%s)", col.Name, col.Type))
			if col.NullRate > 0 {
				prompt.WriteString(fmt.Sprintf(" - %.1f%% missing", col.NullRate*100))
			}
			if len(col.Samples) > 0 {
				prompt.WriteString(fmt.Sprintf(" - samples: %s", strings.Join(col.Samples, ", ")))
			}
			prompt.WriteString("\n")
		}
	}

	// Visual selection context
	prompt.WriteString("\n## Visual Selection Context:\n")
	prompt.WriteString(fmt.Sprintf("- Selection area: (%.1f, %.1f) to (%.1f, %.1f)\n",
		visualQuery.Selection.StartX, visualQuery.Selection.StartY,
		visualQuery.Selection.EndX, visualQuery.Selection.EndY))
	prompt.WriteString(fmt.Sprintf("- Selection size: %.1f × %.1f pixels\n",
		visualQuery.Selection.Width, visualQuery.Selection.Height))

	// User's question
	prompt.WriteString(fmt.Sprintf("\n## User Question:\n%s\n\n", userQuery))

	// Instructions for response
	prompt.WriteString("## Response Instructions:\n")
	prompt.WriteString("1. Analyze the user's question in the context of the selected visual area\n")
	prompt.WriteString("2. Provide specific insights about the data points within the selected region\n")
	prompt.WriteString("3. Reference the dataset columns and their values where relevant\n")
	prompt.WriteString("4. If the question cannot be answered with the selection context, explain why\n")
	prompt.WriteString("5. Keep your response focused and data-driven\n")
	prompt.WriteString("6. Use markdown formatting for better readability\n\n")

	prompt.WriteString("Provide your analysis now:")

	return prompt.String()
}

// buildChartPrompt constructs a specialized prompt for chart generation
func (ai *AIService) buildChartPrompt(userQuery string, sessionData *models.SessionData) string {
	var prompt strings.Builder

	// System instructions for chart generation
	prompt.WriteString("You are Aether, an expert data visualization AI. Your task is to generate ECharts configuration JSON based on user requests and their data.\n\n")

	// Data context
	if sessionData.DataSummary != nil {
		ds := sessionData.DataSummary
		prompt.WriteString("## Dataset Information:\n")
		prompt.WriteString(fmt.Sprintf("- File: %s (%d rows × %d columns)\n", ds.FileName, ds.RowCount, ds.ColumnCount))

		// Column information with enhanced details
		prompt.WriteString("\n### Available Columns:\n")
		for _, col := range ds.Columns {
			prompt.WriteString(fmt.Sprintf("- **%s** (%s)", col.Name, col.Type))
			if col.NullRate > 0 {
				prompt.WriteString(fmt.Sprintf(" - %.1f%% missing", col.NullRate*100))
			}
			if len(col.Samples) > 0 {
				prompt.WriteString(fmt.Sprintf(" - samples: %s", strings.Join(col.Samples, ", ")))
			}
			prompt.WriteString("\n")
		}
	}

	// User's visualization request
	prompt.WriteString(fmt.Sprintf("\n## User Request:\n%s\n\n", userQuery))

	// Chart generation instructions
	prompt.WriteString("## Instructions:\n")
	prompt.WriteString("Generate a complete ECharts configuration object as valid JSON. Follow these rules:\n\n")
	prompt.WriteString("1. **JSON Format**: Return ONLY valid JSON, no markdown formatting\n")
	prompt.WriteString("2. **Data Structure**: Use realistic sample data based on the dataset columns\n")
	prompt.WriteString("3. **Chart Type Selection**:\n")
	prompt.WriteString("   - Numeric data over time: line chart\n")
	prompt.WriteString("   - Categories vs numeric: bar chart\n")
	prompt.WriteString("   - Parts of whole: pie chart\n")
	prompt.WriteString("   - Two numeric columns: scatter plot\n")
	prompt.WriteString("   - Single numeric distribution: histogram\n")
	prompt.WriteString("4. **Styling**: Use professional colors and clear typography\n")
	prompt.WriteString("5. **Interactivity**: Include tooltip and legend when appropriate\n")
	prompt.WriteString("6. **Responsiveness**: Configure for responsive design\n\n")

	prompt.WriteString("Generate the complete ECharts JSON configuration now:")

	return prompt.String()
}

// buildPrompt constructs a context-aware prompt for the AI
func (ai *AIService) buildPrompt(userQuery string, sessionData *models.SessionData) string {
	var prompt strings.Builder

	// System instructions
	prompt.WriteString("You are Aether, an advanced AI assistant specialized in data analysis and visualization. ")
	prompt.WriteString("You help users understand their data through conversational analysis and insights. ")
	prompt.WriteString("Always be helpful, accurate, and provide actionable insights.\n\n")

	// Context about the user's session
	if sessionData != nil {
		prompt.WriteString("## Current Session Context:\n")
		prompt.WriteString(fmt.Sprintf("- Session ID: %s\n", sessionData.SessionID))
		prompt.WriteString(fmt.Sprintf("- Session started: %s\n", sessionData.CreatedAt.Format("2006-01-02 15:04:05")))
		prompt.WriteString(fmt.Sprintf("- Total messages: %d\n", sessionData.MessageCount))

		// Add data context if available
		if sessionData.DataSummary != nil {
			ds := sessionData.DataSummary
			prompt.WriteString("\n## Uploaded Dataset Information:\n")
			prompt.WriteString(fmt.Sprintf("- File: %s\n", ds.FileName))
			prompt.WriteString(fmt.Sprintf("- Dimensions: %d rows × %d columns\n", ds.RowCount, ds.ColumnCount))
			prompt.WriteString(fmt.Sprintf("- Uploaded: %s\n", ds.UploadedAt.Format("2006-01-02 15:04:05")))

			// Add column information
			prompt.WriteString("\n### Column Details:\n")
			for _, col := range ds.Columns {
				prompt.WriteString(fmt.Sprintf("- **%s** (%s): ", col.Name, col.Type))
				if col.NullRate > 0 {
					prompt.WriteString(fmt.Sprintf("%.1f%% null values, ", col.NullRate*100))
				}
				if len(col.Samples) > 0 {
					prompt.WriteString(fmt.Sprintf("samples: %s", strings.Join(col.Samples, ", ")))
				}
				prompt.WriteString("\n")
			}

			// Add overall statistics
			if stats, ok := ds.Stats["completeness"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("\n### Data Quality: %.1f%% complete\n", stats*100))
			}
		}

		// Add recent conversation history
		if len(sessionData.ChatHistory) > 0 {
			prompt.WriteString("\n## Recent Conversation:\n")
			// Include last 5 messages for context
			start := len(sessionData.ChatHistory) - 5
			if start < 0 {
				start = 0
			}

			for i := start; i < len(sessionData.ChatHistory); i++ {
				msg := sessionData.ChatHistory[i]
				prompt.WriteString(fmt.Sprintf("- %s: %s\n",
					strings.Title(msg.Author), msg.Content))
			}
		}
	}

	// Current user query
	prompt.WriteString("\n## Current User Query:\n")
	prompt.WriteString(userQuery)

	// Instructions for response
	prompt.WriteString("\n\n## Response Instructions:\n")
	prompt.WriteString("- Provide a helpful, conversational response\n")
	prompt.WriteString("- If the user has uploaded data, reference it specifically\n")
	prompt.WriteString("- Suggest relevant analysis or visualizations when appropriate\n")
	prompt.WriteString("- If you need more information, ask clarifying questions\n")
	prompt.WriteString("- Keep responses concise but informative\n")

	return prompt.String()
}

// extractJSONFromResponse extracts JSON from AI response (removes markdown formatting)
func (ai *AIService) extractJSONFromResponse(response string) string {
	// Remove markdown code block formatting if present
	jsonRegex := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	matches := jsonRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		return matches[1]
	}

	return response
}

// ValidateConnection tests the AI service connection
func (ai *AIService) ValidateConnection() error {
	// Simple test prompt
	prompt := "Hello, this is a connection test. Please respond with 'OK'."

	resp, err := ai.model.GenerateContent(ai.ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("failed to validate AI connection: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return fmt.Errorf("no response candidates received during validation")
	}

	return nil
}
