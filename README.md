# chell

A CLI that renders waveform data, and allows the user to edit the audio (WIP).

## Features
- WAV (PCM) input (mixes to mono)
- Two rendering modes:
  - braille — high vertical resolution using Unicode Braille characters
  - ascii — simple block/centered waveform
- Automatically sizes to your terminal

## Install
1. go build
2. or: go install github.com/isaacjstriker/chell@latest

## Usage
```
chell [--mode braille|ascii] [--height N] path/to/file.wav
```

## Examples
- Render with braille (default):
  chell track.wav

- Render with ASCII, 20 character rows:
  chell --mode ascii --height 20 track.wav

## Contribute
- Please feel free to fork the repo and contribute your own ideas! I will be adding CI/CD integration in the near future. Thanks for checking this out!
