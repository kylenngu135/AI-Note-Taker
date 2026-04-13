FROM python:3.11

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install -r requirements.txt --break-system-packages --root-user-action=ignore

COPY server.py .

# Download whisper model at build time so runtime doesn't need network
RUN python -c "import whisper; whisper.load_model('small', download_root='/root/.cache/whisper')"

CMD ["python", "server.py"]
