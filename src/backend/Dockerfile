FROM golang:1.25.0

RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

RUN pip3 install openai-whisper --break-system-packages

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main .

CMD ["./main"]
