package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSynthesizeToMP3File(t *testing.T) {
	t.Parallel()

	wantAudio := []byte("ID3-test-audio")
	wantReq := TTSRequest{
		Input:          "hello from clipboard",
		Voice:          "af_heart",
		ResponseFormat: "mp3",
		Speed:          1.0,
		LangCode:       "a",
	}

	ttsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want %q", got, "application/json")
		}

		var gotReq TTSRequest
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if gotReq != wantReq {
			t.Fatalf("request = %#v, want %#v", gotReq, wantReq)
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		if _, err := w.Write(wantAudio); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer ttsServer.Close()

	outputFile := filepath.Join(t.TempDir(), "clip.mp3")
	if err := synthesizeToMP3File(ttsServer.URL, wantReq, outputFile); err != nil {
		t.Fatalf("synthesizeToMP3File() error = %v", err)
	}

	gotAudio, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	if !bytes.Equal(gotAudio, wantAudio) {
		t.Fatalf("audio bytes = %v, want %v", gotAudio, wantAudio)
	}
}
