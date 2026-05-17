package transcription

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	smallFileThreshold = 25 * 1024 * 1024 // 25MB
	defaultChunkSecs   = 600              // 10 minutes
	workerPoolSize     = 5
)

// TranscribeAudioLarge handles transcription for files of any size.
// Files under 25MB are sent directly to the Whisper API.
// Larger files are split into 10-minute chunks and transcribed in parallel
// with a worker pool capped at 5 concurrent Whisper API calls.
func TranscribeAudioLarge(filePath, filename string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	if info.Size() < smallFileThreshold {
		log.Printf("File under 25MB, transcribing directly")
		f, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("open file: %w", err)
		}
		defer func() { _ = f.Close() }()
		return TranscribeAudio(f, filename)
	}

	log.Printf("File over 25MB, splitting into chunks for parallel transcription")

	chunkPaths, err := SplitAudioIntoChunks(filePath, defaultChunkSecs)
	if err != nil {
		return "", fmt.Errorf("split audio: %w", err)
	}
	defer CleanupChunks(chunkPaths)

	log.Printf("Split into %d chunks, transcribing in parallel", len(chunkPaths))

	results := make([]string, len(chunkPaths))

	sem := make(chan struct{}, workerPoolSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var firstErr error
	var once sync.Once

	for i, chunkPath := range chunkPaths {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			log.Printf("Transcribing chunk %d of %d...", idx+1, len(chunkPaths))

			f, err := os.Open(path)
			if err != nil {
				once.Do(func() {
					firstErr = fmt.Errorf("open chunk %d: %w", idx, err)
					cancel()
				})
				return
			}
			defer func() { _ = f.Close() }()

			text, err := TranscribeAudio(f, fmt.Sprintf("chunk_%03d.mp3", idx))
			if err != nil {
				once.Do(func() {
					firstErr = fmt.Errorf("transcribe chunk %d: %w", idx, err)
					cancel()
				})
				return
			}
			results[idx] = text
		}(i, chunkPath)
	}

	wg.Wait()

	if firstErr != nil {
		return "", firstErr
	}

	log.Printf("All chunks transcribed, stitching results")
	return strings.Join(results, "\n"), nil
}
