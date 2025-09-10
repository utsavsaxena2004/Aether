# Synaptic Core - Aether Backend

This is the backend component of the Aether project, built with Go. It provides the API endpoints, WebSocket server, and AI integration for the conversational data analysis platform.

## Features

- **RESTful API**: HTTP endpoints for health checks and data operations
- **WebSocket Server**: Real-time communication with the frontend
- **AI Integration**: Google Gemini API for natural language processing
- **Session Management**: Redis-backed session storage
- **File Processing**: CSV file upload and processing
- **Chart Generation**: AI-powered chart specification generation

## Tech Stack

- [Go](https://golang.org/) - Programming language
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket library
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router
- [Redis](https://redis.io/) - In-memory data store
- [Google Gemini API](https://ai.google.dev/) - AI language model

## Getting Started

### Prerequisites

- Go >=1.22
- Redis server
- Google Gemini API Key

### Development

First, ensure you have the required dependencies:

```bash
go mod tidy
```

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Configure the environment variables:

```env
GEMINI_API_KEY=your_google_gemini_api_key_here
REDIS_ADDR=localhost:6379
PORT=8080
```

Run the application:

```bash
go run main.go
```

The server will start on http://localhost:8080

### Build for Production

```bash
go build -o aether-synaptic-core main.go
```

### Run Production Build

```bash
./aether-synaptic-core
```

## Project Structure

```
.
├── handlers/           # HTTP and WebSocket handlers
│   ├── upload.go       # File upload handler
│   └── websocket.go    # WebSocket connection handler
├── models/             # Data models
│   ├── hub.go          # WebSocket hub for connection management
│   ├── message.go      # Message structures
│   └── session.go      # Session data models
├── services/           # Business logic
│   ├── ai.go           # AI service integration
│   ├── data_processor.go # Data processing logic
│   └── redis.go        # Redis service
├── utils/              # Utility functions
│   └── generators.go   # ID generators and helpers
├── main.go             # Application entry point
├── go.mod              # Go module dependencies
└── go.sum              # Go module checksums
```

## API Endpoints

### Health Check

```
GET /health
```

Returns the health status of the service.

### Root Endpoint

```
GET /
```

Returns API information and status.

### File Upload

```
POST /upload
```

Upload a CSV file for analysis.

### Data Summary

```
GET /data-summary
```

Get the data summary for the current session.

## WebSocket Endpoint

```
GET /ws
```

Establishes a WebSocket connection for real-time communication.

### WebSocket Message Types

1. **Chat Messages** (`chat`): Text messages between user and AI
2. **Chart Specifications** (`chart_spec`): ECharts configuration from AI
3. **Visual Queries** (`visual_query`): Queries based on chart selections
4. **Error Messages** (`error`): Error information
5. **System Messages** (`system`): System-level notifications

## Services

### WebSocket Hub

Manages all active WebSocket connections and message broadcasting.

### AI Service

Integrates with Google Gemini API for:
- Natural language processing
- Chart specification generation
- Data analysis responses

### Redis Service

Handles session persistence and data caching:
- Session storage
- Message history
- Data summaries

### Data Processor

Processes uploaded CSV files:
- Data validation
- Column analysis
- Sample data extraction

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Google Gemini API key | Required |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (if required) | Empty |
| `REDIS_DB` | Redis database number | `0` |
| `PORT` | Server port | `8080` |
| `LOG_LEVEL` | Logging level | `info` |
| `DEBUG` | Debug mode | `false` |

## Logging

The application uses Go's standard logging package with structured logging for important events and errors.

## Error Handling

Comprehensive error handling with:
- Graceful degradation when optional services fail
- Detailed error messages for debugging
- Proper HTTP status codes

## Testing

Run tests with:

```bash
go test ./...
```

## Deployment

The application can be deployed using Docker with the provided Dockerfile, or as a standalone binary.