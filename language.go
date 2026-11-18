package main

import (
	"strings"

	"github.com/pemistahl/lingua-go"
)

// detectKokoroLanguage detects supported languages and returns (lang_code, default_voice).
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
		case lingua.German:
			return "d", "df_xiaobei"
		case lingua.Spanish:
			return "e", "ef_dora"
		}
	}

	// Fallback to US English if uncertain.
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
