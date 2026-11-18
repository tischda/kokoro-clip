package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// https://goreleaser.com/cookbooks/using-main.version/
var (
	name    string
	version string
	date    string
	commit  string
)

func main() {
	helpShort := flag.Bool("?", false, "display this help message")
	helpLong := flag.Bool("help", false, "display this help message")
	versionShort := flag.Bool("v", false, "print version and exit")
	versionLong := flag.Bool("version", false, "print version and exit")

	model := flag.String("model", "kokoro", "TTS model name")
	voice := flag.String("voice", "", "Optional Voice ID override (auto-selected if empty)")
	speed := flag.Float64("speed", 1.0, "Speech speed multiplier (e.g. 0.8 to 1.5)")
	lang := flag.String("lang", "", "Optional language override (auto-detected if empty)")
	serverURL := flag.String("url", "http://localhost:8880/v1/audio/speech", "TTS API endpoint URL")
	apiKey := flag.String("api-key", "", "Optional Bearer API key")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: "+name+` [ OPTIONS ]

Reads Windows clipboard text aloud with automatic language detection.

OPTIONS:

  --model string
        TTS model name (default "kokoro")
  --voice string
        optional Voice ID override (auto-selected if empty)
  --speed float
        speech speed multiplier (e.g. 0.8 to 1.5) (default 1)
  --lang string
        optional language override (auto-detected if empty)
  --url string
        TTS API endpoint URL
        (default "http://localhost:8880/v1/audio/speech")
  --api-key string
        optional Bearer API key
  -?, --help
        display this help message
  -v, --version
        print version and exit`)
	}
	flag.Parse()

	if flag.Arg(0) == "version" || *versionShort || *versionLong {
		fmt.Printf("%s %s, built on %s (commit: %s)\n", name, version, date, commit)
		return
	}

	if *helpShort || *helpLong {
		flag.Usage()
		return
	}

	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(1)
	}

	cleanText, err := readClipboardText()
	if err != nil {
		log.Fatalf("%v", err)
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

	err = synthesizeAndPlay(*serverURL, *apiKey, TTSRequest{
		Model:          *model,
		Input:          cleanText,
		Voice:          finalVoice,
		ResponseFormat: "pcm",
		Speed:          *speed,
		LangCode:       finalLang,
	})
	if err != nil {
		log.Fatalf("%v", err)
	}
}
