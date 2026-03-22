# Go Backend Service with Gin Framework

A simple HTTP server built with Go and the Gin framework that provides URL validation for popular social media platforms.

## Features

- RESTful API endpoints
- URL validation for YouTube and Instagram
- Request logging middleware
- Rate limiting protection
- Structured JSON responses
- Clean modular architecture

## Project Structure

```
/backend
├── main.go              # Entry point and server setup
├── go.mod               # Go module dependencies
├── routes/              # API route handlers
│   └── routes.go
├── middleware/          # HTTP middleware
│   ├── logger.go        # Request/response logging
│   └── ratelimit.go     # Rate limiting
└── utils/               # Utility functions
    └── validator.go     # URL validation logic
```

## Prerequisites

- Go 1.21 or higher

## Installation

1. Navigate to the backend directory:
```bash
cd backend
```

2. Download dependencies:
```bash
go mod download
```

## Running the Server

Start the server with:

```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### Health Check

**GET** `/health`

Returns the server health status.

**Response:**
```json
{
  "status": "ok"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

### Validate URL

**GET** `/api/validate?url=<url>`

Validates if a URL belongs to YouTube or Instagram.

**Query Parameters:**
- `url` (required): The URL to validate

**Response:**
```json
{
  "valid": true,
  "platform": "youtube"
}
```

**Supported Platforms:**
- `youtube` - youtube.com, youtu.be, and subdomains
- `instagram` - instagram.com and subdomains
- `unknown` - for unsupported or invalid URLs

**Examples:**

Valid YouTube URL:
```bash
curl "http://localhost:8080/api/validate?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```
Response:
```json
{
  "valid": true,
  "platform": "youtube"
}
```

Valid Instagram URL:
```bash
curl "http://localhost:8080/api/validate?url=https://instagram.com/example"
```
Response:
```json
{
  "valid": true,
  "platform": "instagram"
}
```

Invalid URL:
```bash
curl "http://localhost:8080/api/validate?url=https://example.com"
```
Response:
```json
{
  "valid": false,
  "platform": "unknown"
}
```

## Middleware

### Request Logger

Logs all incoming requests and outgoing responses with:
- HTTP method
- Request path
- Client IP
- Response status code
- Request duration

### Rate Limiter

Protects the API from abuse with:
- 10 requests per second per IP
- Burst capacity of 20 requests
- Automatic cleanup of inactive clients
- Returns 429 (Too Many Requests) when limit exceeded

## Building for Production

Build the binary:

```bash
go build -o server main.go
```

Run the binary:

```bash
./server
```

## Testing the API

You can test the endpoints using curl, Postman, or any HTTP client.

Example test script:

```bash
# Health check
curl http://localhost:8080/health

# Test YouTube URL
curl "http://localhost:8080/api/validate?url=https://youtube.com/watch?v=test"

# Test Instagram URL
curl "http://localhost:8080/api/validate?url=https://instagram.com/user"

# Test invalid URL
curl "http://localhost:8080/api/validate?url=https://example.com"

# Test short YouTube URL
curl "http://localhost:8080/api/validate?url=https://youtu.be/test"

# Test without protocol
curl "http://localhost:8080/api/validate?url=youtube.com"
```

## Dependencies

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) - Rate limiting

## License

MIT
