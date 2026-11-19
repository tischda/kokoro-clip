package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
)

type TTSRequest struct {
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
	LangCode       string  `json:"lang_code,omitempty"`
}

// postTTSRequest sends a JSON request to Kokoro and returns the HTTP response.
//
// The caller owns closing the returned response body.
func postTTSRequest(serverURL string, req TTSRequest) (*http.Response, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("reach TTS server at %s: %w", serverURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// synthesizeAndPlay sends a TTS request to the server and streams returned
// PCM audio to the local output device.
//
// The serverURL parameter is the full TTS endpoint URL.
// The req parameter contains voice, input text, and synthesis options.
// It returns an error if request encoding, HTTP communication, server-side
// validation, or audio playback initialization fails.
func synthesizeAndPlay(serverURL string, req TTSRequest) error {
	resp, err := postTTSRequest(serverURL, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Guard against common error payload types when audio was expected.
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected audio stream, got %s: %s", contentType, string(body[:min(len(body), 300)]))
	}

	// Configure audio output to match the expected PCM stream format.
	op := &oto.NewContextOptions{
		SampleRate:   24000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return fmt.Errorf("audio init failed: %w", err)
	}
	<-readyChan

	// Play the server response directly from the HTTP stream.
	player := otoCtx.NewPlayer(resp.Body)
	player.Play()

	// Give playback a brief head start, then wait until completion.
	time.Sleep(250 * time.Millisecond)
	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// synthesizeToMP3File sends a TTS request and writes the full MP3 response to
// outputPath.
//
// Unlike streaming playback, this method reads the whole audio response into
// memory before writing the file.
func synthesizeToMP3File(serverURL string, req TTSRequest, outputPath string) error {
	resp, err := postTTSRequest(serverURL, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected MP3 audio, got %s: %s", contentType, string(body[:min(len(body), 300)]))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read MP3 audio: %w", err)
	}
	if len(audioData) == 0 {
		return fmt.Errorf("empty audio response")
	}

	if err := os.WriteFile(outputPath, audioData, 0o644); err != nil {
		return fmt.Errorf("write MP3 file: %w", err)
	}

	return nil
}
