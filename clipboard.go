package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.design/x/clipboard"
)

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
