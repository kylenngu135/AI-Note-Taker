FROM python:3.11

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install -r requirements.txt --break-system-packages --root-user-action=ignore

COPY server.py .

CMD ["python", "server.py"]
