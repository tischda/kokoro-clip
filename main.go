package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/pemistahl/lingua-go"
	"golang.design/x/clipboard"
)

type TTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
	LangCode       string  `json:"lang_code,omitempty"`
}

func cleanTextForSpeech(text string) string {
	reCodeBlock := regexp.MustCompile("(?s)```.*?```")
	text = reCodeBlock.ReplaceAllString(text, " Code snippet omitted. ")

	reMdLink := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	text = reMdLink.ReplaceAllString(text, "$1")

	reURL := regexp.MustCompile(`https?://[^\s\)]+`)
	text = reURL.ReplaceAllString(text, "")

	reMdSymbols := regexp.MustCompile(`[*_#` + "`" + `]+`)
	text = reMdSymbols.ReplaceAllString(text, "")

	reSpaces := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reSpaces.ReplaceAllString(text, " "))
}

// detectKokoroLanguage detects English or French and returns (lang_code, default_voice).
func detectKokoroLanguage(text string) (string, string) {
	languages := []lingua.Language{
        lingua.English,
        lingua.French,
        lingua.German,
        lingua.Spanish,
    }

    detector := lingua.NewLanguageDetectorBuilder().
        FromLanguages(languages...).
        Build()

	if lang, exists := detector.DetectLanguageOf(text); exists {
		switch lang {
		case lingua.French:
			return "f", "ff_siwis"
		case lingua.English:
			return "a", "af_heart"
		}
	}
	// Fallback to US English if uncertain
	return "a", "af_heart"
}

func normalizeKokoroLang(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "en", "en-us", "us":
		return "a"
	case "en-gb", "en-uk", "uk":
		return "b"
	case "es":
		return "e"
	case "fr":
		return "f"
	case "hi":
		return "h"
	case "it":
		return "i"
	case "ja", "jp":
		return "j"
	case "pt", "pt-br":
		return "p"
	case "zh", "cn":
		return "z"
	default:
		return code
	}
}

func main() {
	help := flag.Bool("help", false, "Show usage instructions and available flags")
	model := flag.String("model", "kokoro", "TTS model name")
	voice := flag.String("voice", "", "Optional Voice ID override (auto-selected if empty)")
	speed := flag.Float64("speed", 1.0, "Speech speed multiplier (e.g. 0.8 to 1.5)")
	lang := flag.String("lang", "", "Optional language override (auto-detected if empty)")
	serverURL := flag.String("url", "http://localhost:8880/v1/audio/speech", "TTS API endpoint URL")
	apiKey := flag.String("api-key", "", "Optional Bearer API key")
	flag.Parse()

	if *help {
		fmt.Println("kokoro-clip - Read Windows clipboard text aloud with automatic language detection")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if err := clipboard.Init(); err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	rawBytes, err := clipboard.Read(context.Background(), clipboard.FmtText)
	if err != nil {
		log.Fatalf("Failed to read clipboard: %v", err)
	}

	cleanText := cleanTextForSpeech(string(rawBytes))
	if cleanText == "" {
		log.Fatal("Clipboard is empty or has no speakable text.")
	}

	// Auto-detect language and default voice unless overridden via CLI flags
	detectedLang, detectedVoice := detectKokoroLanguage(cleanText)

	finalLang := detectedLang
	if *lang != "" {
		finalLang = normalizeKokoroLang(*lang)
	}

	finalVoice := detectedVoice
	if *voice != "" {
		finalVoice = *voice
	}

	fmt.Printf("Speaking (%s / %s)...\n", finalLang, finalVoice)

	reqBody, err := json.Marshal(TTSRequest{
		Model:          *model,
		Input:          cleanText,
		Voice:          finalVoice,
		ResponseFormat: "pcm",
		Speed:          *speed,
		LangCode:       finalLang,
	})
	if err != nil {
		log.Fatalf("Failed to encode request: %v", err)
	}

	req, err := http.NewRequest("POST", *serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to build HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if *apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+*apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to reach TTS server at %s: %v", *serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Server error (%d): %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Expected audio stream, got %s: %s", contentType, string(body[:min(len(body), 300)]))
	}

	op := &oto.NewContextOptions{
		SampleRate:   24000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		log.Fatalf("Audio init failed: %v", err)
	}
	<-readyChan

	player := otoCtx.NewPlayer(resp.Body)
	player.Play()

	time.Sleep(250 * time.Millisecond)

	for player.IsPlaying() {
		time.Sleep(100 * time.Millisecond)
	}
}
