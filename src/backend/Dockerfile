FROM golang:1.25.0

RUN --network=host apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY backend/ ./backend/
COPY frontend/ ./frontend/
COPY ui/ ./ui/

WORKDIR /app/backend

RUN --network=host go mod download

RUN go build -o main .

CMD ["./main"]
