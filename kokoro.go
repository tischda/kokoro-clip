package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// synthesizeAndPlay sends a TTS request to the server and streams returned
// PCM audio to the local output device.
//
// The serverURL parameter is the full TTS endpoint URL.
// The req parameter contains voice, input text, and synthesis options.
// It returns an error if request encoding, HTTP communication, server-side
// validation, or audio playback initialization fails.
func synthesizeAndPlay(serverURL string, req TTSRequest) error {
	// Encode the request body once so it can be sent as JSON.
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	// Build a POST request to the local Kokoro FastAPI endpoint.
	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request and stream the response body.
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("reach TTS server at %s: %w", serverURL, err)
	}
	defer resp.Body.Close()

	// Surface non-200 responses with the server payload for debugging.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

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
