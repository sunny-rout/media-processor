# YouTube Downloader

A full-stack application for downloading YouTube videos and audio. Built with a Go backend using Gin and yt-dlp, and a Next.js frontend with TailwindCSS.

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
- YouTube URL validation
- Format selection (Video/Audio)
- Loading indicators
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
2. Enter a YouTube URL (e.g., `https://www.youtube.com/watch?v=...`)
3. Select format:
   - **Video** - Downloads as MP4
   - **Audio** - Downloads as MP3
4. Click **Download** button
5. File will be saved to your downloads folder

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

1. **User Input**: User enters YouTube URL and selects format
2. **Validation**: Frontend validates URL format
3. **API Request**: Frontend sends GET request to backend `/api/stream`
4. **Backend Processing**:
   - Backend validates URL
   - Spawns yt-dlp process
   - Streams output directly to response
5. **Frontend Download**:
   - Receives blob stream
   - Creates download link
   - Triggers browser download
6. **Cleanup**: Both frontend and backend clean up resources

### Error Handling

The application handles:
- Empty URL input
- Invalid YouTube URLs
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

## Troubleshooting

### Backend won't start
- Ensure Go is installed: `go version`
- Check if port 8080 is available
- Verify yt-dlp is installed: `yt-dlp --version`

### Frontend can't connect to backend
- Ensure backend is running on port 8080
- Check console for CORS errors
- Verify `BACKEND_URL` in `app/page.tsx`

### Downloads fail
- Ensure yt-dlp is up to date: `pip install -U yt-dlp`
- Check backend logs for errors
- Verify YouTube URL is valid and accessible

## License

MIT
