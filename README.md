# AI Note Taker

Upload documents or audio/video files and receive AI-powered study notes instantly.

## Project Description

AI Note Taker is a full-stack web application that allows users to upload documents (PDF, DOCX, TXT) or audio/video files, transcribes the content, and generates comprehensive study sheets using Groq's Llama 3.3 LLM. The application features a Go backend that serves both the API and frontend, a Python Whisper service for audio/video transcription, PostgreSQL for metadata storage, and Cloudflare R2 for object storage.

### Architecture

```
Browser → Go Backend (:8080) → PostgreSQL (metadata)
                  ↓           R2 (files/blobs)
                  ↓
         Python Whisper (:8081) → audio/video transcription
                  ↓
           Groq API (Llama 3.3) → study sheet generation
```

## Installation

### Prerequisites

- Docker and Docker Compose
- Environment variables (see below)

### Environment Setup

Create a `.env` file in the project root with the following variables:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name

# Cloudflare R2
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key
R2_SECRET_ACCESS_KEY=your_r2_secret_key
R2_BUCKET_NAME=your_bucket_name

# Groq API
GROQ_API_KEY=your_groq_api_key

# JWT
JWT_SECRET=your_jwt_secret
```

### Running the Application

Start all services with Docker Compose:

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`.

## Usage

### Authentication

Before uploading files, register or sign in via the auth page at `/auth.html`.

### Uploading Files

1. Navigate to the main page at `/`
2. Sign in with your credentials
3. Drag and drop a file or click to browse
   - Supported document formats: PDF, DOCX, TXT
   - Supported audio formats: MP3, WAV, M4A
   - Supported video formats: MP4, MOV
4. Optionally provide a custom prompt for study sheet generation
5. Click "Upload" to process the file

### Viewing Notes

After upload, the generated study sheet is displayed on the main page. You can:
- Ask follow-up questions using the message bar
- Export notes in TXT, PDF, or DOCX format
- Delete the upload and associated notes

### Regenerating Notes

To regenerate notes with a custom prompt:
1. Click on an existing upload
2. Use the message bar to provide a specific prompt
3. The AI will regenerate the study sheet using your prompt and existing notes as context

## Features

- **Multi-format document support** — Upload PDF, DOCX, TXT documents or audio/video files
- **Automatic transcription** — Documents are parsed directly in Go; audio/video is transcribed via OpenAI Whisper
- **AI-powered study sheets** — Generates comprehensive study notes with summary, key concepts, main topics, and action items
- **Note regeneration** — Regenerate study sheets with custom prompts using existing notes as context
- **Multiple export formats** — Download notes as TXT, PDF, or DOCX
- **User authentication** — JWT-based authentication with bcrypt password hashing
- **Cloud storage** — Transcriptions and notes stored in Cloudflare R2
- **Note history** — Tracks all regenerated versions of study sheets
- **Responsive UI** — Clean, modern interface with sidebar navigation and export modals

## Tech Stack

- **Backend**: Go, PostgreSQL, Cloudflare R2 (AWS SDK v2)
- **Frontend**: Vanilla JS, HTML, CSS
- **Transcription**: Python Flask with OpenAI Whisper
- **AI**: Groq API with Llama 3.3 (70B)
- **Authentication**: JWT (HS256), bcrypt
