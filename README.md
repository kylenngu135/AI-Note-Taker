# AI Note Taker

Upload documents or audio/video files and receive AI-powered study notes.

## Description

AI Note Taker is a full-stack web application that accepts document and media uploads, extracts or transcribes their content, and generates structured study sheets using OpenAI GPT-4o-mini. File processing runs asynchronously via a Redis-backed job queue. Metadata is stored in PostgreSQL; file blobs are stored in Cloudflare R2.

## Architecture

```
Browser
  │
  ▼
Go Backend (:8080) ──────────────────────────────────┐
  │                                                   │
  │  POST /api/uploads                                │
  │  ├── validates file                               │
  │  ├── stores raw file → R2                         │
  │  ├── creates upload row (status: pending) → PG    │
  │  └── enqueues job → Redis (Upstash)               │
  │                                                   │
  │  Background Worker (goroutine)                    │
  │  ├── dequeues job from Redis                      │
  │  ├── downloads raw file ← R2                      │
  │  ├── transcribes:                                 │
  │  │     documents (PDF/DOCX/TXT) → Go parser       │
  │  │     audio/video (MP3/MP4)   → OpenAI Whisper   │
  │  ├── generates study sheet → OpenAI GPT-4o-mini   │
  │  ├── stores transcription + notes → R2            │
  │  └── writes note + history → PostgreSQL           │
  │                                                   │
  └── Serves frontend static files from src/ui/       │
                                                      │
External Services                                     │
  ├── OpenAI API (Whisper + GPT-4o-mini) ────────────┘
  ├── Cloudflare R2 (S3-compatible object storage)
  ├── PostgreSQL (metadata, users, notes, tags)
  └── Redis / Upstash (job queue)
```

## Prerequisites

- Go 1.25+
- PostgreSQL database (local or hosted, e.g. Supabase)
- Upstash Redis instance
- Cloudflare R2 bucket
- OpenAI API key

## Environment Variables

Create a `.env` file in the project root:

```env
# S3 API
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key
R2_SECRET_ACCESS_KEY=your_r2_secret_key
R2_BUCKET_NAME=your_bucket_name

# Database URL (use the direct connection string, not the pgBouncer pooled URL)
DATABASE_URL=postgresql://user:password@host:5432/dbname

# Redis
REDIS_URL=https://your-instance.upstash.io
REDIS_TOKEN=your_upstash_token

# OpenAI API
OPENAI_API_KEY=your_openai_api_key

# JWT Secret Key
JWT_SECRET=your_jwt_secret
```

## Running

```bash
cd src/backend && go run .
```

The server starts on `http://localhost:8080`. Database migrations run automatically on startup.

## API Docs

Interactive API documentation is available at `http://localhost:8080/api-docs` once the server is running. The raw OpenAPI spec is at `http://localhost:8080/api-docs/openapi.yaml`.

## Usage

### Authentication

Register or sign in at `/auth.html` before uploading files.

### Uploading Files

Supported formats:
- **Documents**: PDF, DOCX, TXT (transcribed in Go, no external service required)
- **Audio**: MP3
- **Video**: MP4

Uploads are processed asynchronously. After submitting a file, the UI polls for completion using the returned `upload_id`.

### Study Notes

Each upload produces a structured notes sheet with:
- Summary
- Key concepts
- Detailed notes
- Important quotes or statements
- Action items / takeaways

### Follow-up Questions

Select any upload and use the message bar to ask follow-up questions. The full conversation history is maintained and sent to GPT-4o-mini with each request, so responses are context-aware.

### Export

Notes can be exported as TXT, PDF, or DOCX from the notes view.

## Features

- **Async processing** — uploads return immediately; background workers handle transcription and note generation
- **Multi-format support** — PDF, DOCX, TXT parsed natively in Go; MP3/MP4 sent to OpenAI Whisper
- **AI study sheets** — GPT-4o-mini produces structured notes with summary, key concepts, detailed notes, quotes, and action items
- **Conversation history** — follow-up questions build on the full prior context
- **Tags** — auto-generated topic tags and file-type tags on each upload; manual tags supported
- **Export** — download notes as TXT, PDF, or DOCX
- **JWT authentication** — HS256 tokens, bcrypt password hashing, HttpOnly cookie storage
- **Auto-migrations** — golang-migrate runs pending SQL migrations at startup

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go (standard library `net/http`) |
| Database | PostgreSQL + golang-migrate |
| Object storage | Cloudflare R2 (AWS SDK v2) |
| Job queue | Upstash Redis |
| Transcription | OpenAI Whisper API (`whisper-1`) |
| Note generation | OpenAI GPT-4o-mini |
| Frontend | Vanilla JS, HTML, CSS |
| Auth | JWT (HS256), bcrypt |
