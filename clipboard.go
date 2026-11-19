package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.design/x/clipboard"
)

// readClipboardText initializes the clipboard backend, reads plain text,
// and returns sanitized text suitable for speech synthesis.
//
// It returns an error if clipboard initialization or reading fails, or if the
// resulting text is empty after sanitization.
func readClipboardText() (string, error) {
	if err := clipboard.Init(); err != nil {
		return "", fmt.Errorf("initialize clipboard: %w", err)
	}

	rawBytes, err := clipboard.Read(context.Background(), clipboard.FmtText)
	if err != nil {
		return "", fmt.Errorf("read clipboard: %w", err)
	}

	cleanText := cleanTextForSpeech(string(rawBytes))
	if cleanText == "" {
		return "", fmt.Errorf("clipboard is empty or has no speakable text")
	}

	return cleanText, nil
}

// cleanTextForSpeech removes markdown artifacts and URLs and normalizes
// whitespace to produce cleaner spoken output.
//
// The text parameter is raw clipboard text that may include markdown and URLs.
// It returns a trimmed string with normalized spacing for speech output.
func cleanTextForSpeech(text string) string {
	// reCodeBlock matches fenced markdown code blocks, including multiline
	// content, and replaces them with a short spoken placeholder.
	reCodeBlock := regexp.MustCompile("(?s)```.*?```")
	text = reCodeBlock.ReplaceAllString(text, " Code snippet omitted. ")

	// reMdLink matches markdown links and keeps only the visible link text.
	reMdLink := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	text = reMdLink.ReplaceAllString(text, "$1")

	// reURL matches raw HTTP(S) URLs and removes them from spoken output.
	reURL := regexp.MustCompile(`https?://[^\s\)]+`)
	text = reURL.ReplaceAllString(text, "")

	// reMdSymbols matches common markdown formatting symbols and strips them.
	reMdSymbols := regexp.MustCompile(`[*_#` + "`" + `]+`)
	text = reMdSymbols.ReplaceAllString(text, "")

	// reSpaces collapses all whitespace runs into single spaces and trims ends.
	reSpaces := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reSpaces.ReplaceAllString(text, " "))
}
