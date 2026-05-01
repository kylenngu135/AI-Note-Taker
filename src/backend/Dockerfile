FROM golang:1.25.0

RUN apt-get update && apt-get install -y \
    ffmpeg \
    libmupdf-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY backend/ ./backend/
COPY frontend/ ./frontend/
COPY ui/ ./ui/

WORKDIR /app/backend

RUN go mod download

RUN go build -o main .

CMD ["./main"]
