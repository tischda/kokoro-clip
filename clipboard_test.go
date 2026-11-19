package main

import "testing"

func TestCleanTextForSpeech(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "replace fenced code block",
			input: "Before ```go\nfmt.Println(\"hi\")\n``` after",
			want:  "Before Code snippet omitted. after",
		},
		{
			name:  "keep markdown link label and remove raw url",
			input: "See [docs](https://docs.example.com) and https://example.com/path",
			want:  "See docs and",
		},
		{
			name:  "strip markdown symbols and normalize whitespace",
			input: "*Hello* _world_ #tag `code`\n\nnext",
			want:  "Hello world tag code next",
		},
		{
			name:  "combined cleanup behavior",
			input: "# Title\nVisit [site](https://site.example) and https://raw.example\n```txt\nblock\n```\n*done*",
			want:  "Title Visit site and Code snippet omitted. done",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cleanTextForSpeech(tc.input)
			if got != tc.want {
				t.Fatalf("cleanTextForSpeech() = %q, want %q", got, tc.want)
			}
		})
	}
}
