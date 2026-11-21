[![Build Status](https://github.com/tischda/kokoro-clip/actions/workflows/build.yml/badge.svg)](https://github.com/tischda/kokoro-clip/actions/workflows/build.yml)
[![Test Status](https://github.com/tischda/kokoro-clip/actions/workflows/test.yml/badge.svg)](https://github.com/tischda/kokoro-clip/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/tischda/kokoro-clip/badge.svg)](https://coveralls.io/r/tischda/kokoro-clip)
[![Linter Status](https://github.com/tischda/kokoro-clip/actions/workflows/linter.yml/badge.svg)](https://github.com/tischda/kokoro-clip/actions/workflows/linter.yml)
[![License](https://img.shields.io/github/license/github.com/tischda/kokoro-clip.svg)](/LICENSE)
[![Release](https://img.shields.io/github/release/github.com/tischda/kokoro-clip.svg)](https://github.com/github.com/tischda/kokoro-clip/releases/latest)

# kokoro-clip

Reads Windows clipboard text aloud with automatic language detection.

## Install

Install [Kokoro-FastAPI](https://github.com/remsky/Kokoro-FastAPI) TTS engine and start it with:

```
pwsh start-gpu.ps1
```

Install kokoro-clip

```
go install github.com/tischda/kokoro-clip@latest
```

## Usage

```
kokoro-clip reads clipboard text aloud with automatic language detection.

FLAGS:

  -?, --help            display this help message
      --lang string     Optional language override (auto-detected if empty)
  -o, --output string   Write synthesized audio to an MP3 file instead of playing it
      --speed float     Speech speed multiplier (e.g. 0.8 to 1.5) (default 1)
      --url string      TTS API endpoint URL (default "http://localhost:8880/v1/audio/speech")
  -v, --version         print version and exit
      --voice string    Optional Voice ID override (auto-selected if empty)
```

## Examples

```
kokoro-clip
kokoro-clip --output clip.mp3
```
