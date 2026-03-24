# Media Downloader

A full-stack application for downloading YouTube and Instagram videos and audio. Built with a Go backend using Gin and yt-dlp, and a Next.js frontend with TailwindCSS.

## Project Structure

```
/
├── backend/          # Go backend service
│   ├── main.go       # Server entry point
│   ├── routes/       # API route handlers
│   ├── middleware/   # Middleware (logging, rate limiting)
│   └── utils/        # Utility functions
│
├── app/              # Next.js frontend (App Router)
│   ├── page.tsx      # Main download page
│   ├── layout.tsx    # Root layout
│   └── globals.css   # Global styles
│
└── package.json      # Frontend dependencies
```

## Features

### Backend
- RESTful API with Gin framework
- YouTube/Instagram URL validation
- Real-time media streaming with yt-dlp
- Request logging and rate limiting
- Client disconnect handling
- Process timeout protection

### Frontend
- Clean, minimal UI with TailwindCSS
- Real-time download streaming
- YouTube and Instagram URL validation
- Platform detection with visual feedback
- Format selection (Video/Audio)
- Loading indicators with progress tracking
- Comprehensive error handling
- Dark mode support

## Prerequisites

### Backend
- Go 1.21 or higher
- yt-dlp installed on system
  ```bash
  pip install yt-dlp
  # or
  brew install yt-dlp  # macOS
  ```
- gallery-dl installed on system (for Instagram)
  ```bash
  pip install gallery-dl
  # or
  brew install gallery-dl  # macOS
  ```

### Frontend
- Node.js 18+ and npm

## Installation

### 1. Install Frontend Dependencies

```bash
npm install
```

### 2. Install Backend Dependencies

```bash
cd backend
go mod download
```

## Running the Application

### Start Backend Server

```bash
cd backend
go run main.go
```

Server will start on `http://localhost:8080`

### Start Frontend Development Server

In a separate terminal:

```bash
npm run dev
```

Frontend will start on `http://localhost:3000`

## Usage

1. Open `http://localhost:3000` in your browser
2. Enter a YouTube or Instagram URL:
   - YouTube: `https://www.youtube.com/watch?v=...`
   - Instagram: `https://www.instagram.com/p/...` or `https://www.instagram.com/reel/...`
3. The platform will be automatically detected
4. Select format:
   - **Video** - Downloads as MP4
   - **Audio** - Downloads as MP3 (extracted)
5. Click **Download** button
6. File will be saved to your downloads folder

**Note**: Instagram posts must be public. Private content will return an error.

## API Endpoints

### Backend API (Port 8080)

#### Health Check
```
GET /health
```

#### Validate URL
```
GET /api/validate?url=<url>
```

#### Stream Media
```
GET /api/stream?url=<url>&format=video|audio
```

## How It Works

### Download Flow

1. **User Input**: User enters YouTube or Instagram URL and selects format
2. **Platform Detection**: Frontend detects platform and shows appropriate feedback
3. **Validation**: Frontend validates URL format for supported platforms
4. **API Request**: Frontend sends GET request to backend `/api/stream`
5. **Backend Processing**:
   - Backend validates URL and checks accessibility
   - Spawns yt-dlp process with platform-specific options
   - Streams output directly to response
6. **Frontend Download**:
   - Receives blob stream with progress tracking
   - Creates download link
   - Triggers browser download
7. **Cleanup**: Both frontend and backend clean up resources

### Error Handling

The application handles:
- Empty URL input
- Invalid YouTube or Instagram URLs
- Unsupported platforms
- Private Instagram content (403)
- Unavailable content (404)
- Backend connection failures
- yt-dlp processing errors
- Network timeouts
- Client disconnections

## Building for Production

### Frontend

```bash
npm run build
npm start
```

### Backend

```bash
cd backend
go build -o server main.go
./server
```

## Environment

Backend runs on port 8080 by default. To change, modify `backend/main.go`:

```go
router.Run(":8080")  // Change port here
```

Frontend connects to backend at `http://localhost:8080`. To change, modify `app/page.tsx`:

```typescript
const BACKEND_URL = "http://localhost:8080";  // Change URL here
```

## Technology Stack

### Backend
- Go 1.21+
- Gin Web Framework
- yt-dlp
- golang.org/x/time/rate

### Frontend
- Next.js 16.2+
- React 19
- TailwindCSS 3.4+
- TypeScript 5

## Monitoring with Prometheus + Grafana

### Prerequisites
- Docker and Docker Compose installed

### Step 1: Install Prometheus client for Go

```bash
cd backend
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Step 2: Start Prometheus and Grafana

```bash
docker-compose up -d
```

### Step 3: Access Services

| Service | URL | Default Credentials |
|---------|-----|---------------------|
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3001 | admin / admin |

### Step 4: Verify Prometheus is Scraping

1. Open http://localhost:9090
2. Go to **Status > Targets**
3. Confirm `media-processor` target shows `UP`

### Step 5: View Dashboards in Grafana

1. Open http://localhost:3001 and login
2. Go to **Dashboards > Manage**
3. Open the **Media Processor** dashboard

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `app_uptime_seconds` | Gauge | Application uptime in seconds |
| `stream_started_total` | Counter | Total streams started |
| `stream_success_total` | Counter | Total successful streams |
| `stream_errors_total` | Counter | Total errors by type (pool_full, private, unavailable, process, timeout) |
| `stream_duration_seconds` | Histogram | Stream duration buckets |
| `worker_pool_size` | Gauge | Current worker pool size |

### Stop Monitoring

```bash
docker-compose down
```

## Redis

Redis is used for rate limiting. It runs automatically in Docker.

| Setting | Value |
|---------|-------|
| Host | localhost (outside Docker) |
| Container | redis:6379 (inside Docker) |
| Persistence | AOF enabled |

**Environment variable:**
```bash
REDIS_URL=redis://redis:6379  # Docker
REDIS_URL=redis://localhost:6379  # Local
```

## Troubleshooting

### Backend won't start
- Ensure Go is installed: `go version`
- Check if port 8080 is available
- Verify yt-dlp is installed: `yt-dlp --version`
- Verify gallery-dl is installed: `gallery-dl --version`

### Frontend can't connect to backend
- Ensure backend is running on port 8080
- Check console for CORS errors
- Verify `BACKEND_URL` in `app/page.tsx`

### Downloads fail
- Ensure yt-dlp is up to date: `pip install -U yt-dlp`
- Ensure gallery-dl is up to date: `pip install -U gallery-dl`
- Check backend logs for errors
- Verify URL is valid and accessible
- For Instagram: Ensure the post is public (not private)
- Try the URL directly: `gallery-dl <url>` (Instagram) or `yt-dlp <url>` (YouTube)

## License

MIT
