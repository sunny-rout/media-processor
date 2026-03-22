# Go Backend Service with Gin Framework

A powerful HTTP server built with Go and the Gin framework that provides URL validation and media streaming capabilities for popular social media platforms.

## Features

- RESTful API endpoints
- URL validation for YouTube and Instagram
- YouTube media streaming (video/audio) using yt-dlp
- Request logging middleware
- Rate limiting protection
- Structured JSON responses
- Clean modular architecture
- Client disconnect handling
- Process timeout protection

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
- yt-dlp installed on the system (for media streaming)
  - Install via: `pip install yt-dlp` or download from [yt-dlp releases](https://github.com/yt-dlp/yt-dlp/releases)

## Installation

1. Navigate to the backend directory:
```bash
cd backend
```

2. Download dependencies:
```bash
go mod download

go mod tidy
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

### Stream Media

**GET** `/api/stream?url=<url>&format=<format>`

Streams YouTube videos or audio directly without storing files on disk.

**Query Parameters:**
- `url` (required): The YouTube URL to stream
- `format` (required): Either `video` or `audio`

**Response:**
- Streams media file with appropriate headers
- Content-Disposition: `attachment; filename="media.mp4"` or `"media.mp3"`
- Content-Type: `application/octet-stream`

**Features:**
- No disk storage - streams directly from yt-dlp output
- 3-minute timeout protection
- Client disconnect handling (cancels yt-dlp process)
- Detailed error logging

**Examples:**

Stream video:
```bash
curl "http://localhost:8080/api/stream?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=video" -o video.mp4
```

Stream audio:
```bash
curl "http://localhost:8080/api/stream?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=audio" -o audio.mp3
```

Browser download (open in browser):
```
http://localhost:8080/api/stream?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=video
```

**Error Responses:**

Missing parameters (400):
```json
{
  "error": "URL parameter is required"
}
```

Invalid format (400):
```json
{
  "error": "format parameter must be 'video' or 'audio'"
}
```

Non-YouTube URL (400):
```json
{
  "error": "Only valid YouTube URLs are supported"
}
```

yt-dlp failure (500):
```json
{
  "error": "Failed to start media processing"
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

# Test YouTube URL validation
curl "http://localhost:8080/api/validate?url=https://youtube.com/watch?v=test"

# Test Instagram URL validation
curl "http://localhost:8080/api/validate?url=https://instagram.com/user"

# Test invalid URL
curl "http://localhost:8080/api/validate?url=https://example.com"

# Test short YouTube URL
curl "http://localhost:8080/api/validate?url=https://youtu.be/test"

# Test without protocol
curl "http://localhost:8080/api/validate?url=youtube.com"

# Stream YouTube video
curl "http://localhost:8080/api/stream?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=video" -o test-video.mp4

# Stream YouTube audio
curl "http://localhost:8080/api/stream?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=audio" -o test-audio.mp3
```

## Dependencies

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) - Rate limiting
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube media downloader (external dependency)

## Important Notes

### Media Streaming
- The `/api/stream` endpoint requires yt-dlp to be installed on the system
- Video streaming uses the best available format from YouTube
- Audio streaming extracts and converts to MP3 format
- No files are stored on disk - everything streams directly to the client
- Each streaming request has a 3-minute timeout to prevent hanging connections
- If the client disconnects, the yt-dlp process is automatically terminated

### Performance Considerations
- Media streaming is resource-intensive and may consume significant bandwidth
- Consider implementing additional rate limiting specifically for the streaming endpoint in production
- Monitor server resources when handling multiple concurrent streaming requests

### Security
- Only YouTube URLs are accepted for streaming
- URL validation is performed before initiating any yt-dlp process
- Process timeouts prevent indefinite resource consumption
- Client disconnect detection ensures proper cleanup

## License

MIT
