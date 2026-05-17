package transcription

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// SplitAudioIntoChunks splits inputPath into chunks of at most chunkDuration seconds,
// preferring silence boundaries to avoid mid-word cuts.
// Returns ordered chunk file paths (all .mp3) in a temp directory.
func SplitAudioIntoChunks(inputPath string, chunkDuration int) ([]string, error) {
	tmpDir, err := os.MkdirTemp("", "audio-chunks-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	duration, err := getAudioDuration(inputPath)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("get audio duration: %w", err)
	}

	// Build ideal boundary points at each chunkDuration interval.
	var boundaries []float64
	for t := float64(chunkDuration); t < duration; t += float64(chunkDuration) {
		boundaries = append(boundaries, t)
	}

	if len(boundaries) == 0 {
		// File shorter than one chunk — convert to a single mp3 chunk.
		outPath := filepath.Join(tmpDir, "chunk_000.mp3")
		cmd := exec.Command("ffmpeg", "-loglevel", "error",
			"-i", inputPath,
			"-c:a", "libmp3lame", "-q:a", "3",
			outPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("ffmpeg convert: %w: %s", err, out)
		}
		return []string{outPath}, nil
	}

	// Use silence detection to snap boundaries to natural pause points.
	silenceEnds := detectSilencePoints(inputPath)

	const tolerance = 30.0
	splitPoints := make([]float64, len(boundaries))
	for i, boundary := range boundaries {
		best := boundary
		bestDist := tolerance + 1
		for _, s := range silenceEnds {
			if d := math.Abs(s - boundary); d < bestDist {
				bestDist = d
				best = s
			}
		}
		splitPoints[i] = best
	}

	starts := append([]float64{0}, splitPoints...)
	ends := append(splitPoints, duration)

	chunkPaths := make([]string, len(starts))
	for i, start := range starts {
		outPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%03d.mp3", i))
		args := []string{
			"-loglevel", "error",
			"-i", inputPath,
			"-ss", strconv.FormatFloat(start, 'f', 3, 64),
			"-to", strconv.FormatFloat(ends[i], 'f', 3, 64),
			"-c:a", "libmp3lame", "-q:a", "3",
			outPath,
		}
		cmd := exec.Command("ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("ffmpeg chunk %d: %w: %s", i, err, out)
		}
		chunkPaths[i] = outPath
	}

	return chunkPaths, nil
}

// CleanupChunks removes all chunk temp files and their parent directory.
func CleanupChunks(chunkPaths []string) {
	if len(chunkPaths) == 0 {
		return
	}
	_ = os.RemoveAll(filepath.Dir(chunkPaths[0]))
}

func getAudioDuration(inputPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	d, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", strings.TrimSpace(string(out)), err)
	}
	return d, nil
}

// detectSilencePoints returns silence_end timestamps parsed from ffmpeg's silencedetect filter.
// Errors from ffmpeg are ignored; the result may be empty.
func detectSilencePoints(inputPath string) []float64 {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-af", "silencedetect=n=-40dB:d=0.5",
		"-f", "null", "-")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run() // non-zero exit is expected with -f null

	var points []float64
	scanner := bufio.NewScanner(&stderr)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, "silence_end:")
		if idx < 0 {
			continue
		}
		rest := line[idx+len("silence_end:"):]
		rest = strings.SplitN(rest, "|", 2)[0]
		val, err := strconv.ParseFloat(strings.TrimSpace(rest), 64)
		if err == nil {
			points = append(points, val)
		}
	}
	return points
}
