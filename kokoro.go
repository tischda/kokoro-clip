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
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
	LangCode       string  `json:"lang_code,omitempty"`
}

func synthesizeAndPlay(serverURL, apiKey string, req TTSRequest) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("reach TTS server at %s: %w", serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected audio stream, got %s: %s", contentType, string(body[:min(len(body), 300)]))
	}

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

	player := otoCtx.NewPlayer(resp.Body)
	player.Play()

	time.Sleep(250 * time.Millisecond)
	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
