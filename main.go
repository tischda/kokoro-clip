package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

// https://goreleaser.com/cookbooks/using-main.version/
var (
	name    string
	version string
	date    string
	commit  string
)

// printUsage prints CLI help text and examples.
func printUsage() {
	fmt.Println(name + " reads clipboard text aloud with automatic language detection.")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println()
	pflag.PrintDefaults()
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println()
	fmt.Println("  $ " + name)
}

// main parses CLI flags, reads clipboard text, and sends it to Kokoro TTS.
func main() {
	help := pflag.BoolP("help", "?", false, "display this help message")
	showVersion := pflag.BoolP("version", "v", false, "print version and exit")

	voice := pflag.String("voice", "", "Optional Voice ID override (auto-selected if empty)")
	speed := pflag.Float64("speed", 1.0, "Speech speed multiplier (e.g. 0.8 to 1.5)")
	lang := pflag.String("lang", "", "Optional language override (auto-detected if empty)")
	serverURL := pflag.String("url", "http://localhost:8880/v1/audio/speech", "TTS API endpoint URL")

	pflag.Usage = printUsage
	pflag.Parse()

	if pflag.Arg(0) == "version" || *showVersion {
		fmt.Printf("%s %s, built on %s (commit: %s)\n", name, version, date, commit)
		return
	}

	if *help {
		pflag.Usage()
		os.Exit(0)
	}

	if pflag.NArg() > 0 {
		pflag.Usage()
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

	err = synthesizeAndPlay(*serverURL, TTSRequest{
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
