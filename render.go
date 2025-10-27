package cmd

import (
	"fmt"
	"io"

	"github.com/isaacjstriker/chell/waveform"
)

func runFile(path, mode string, height int, out io.Writer) error {
	samples, sr, err := waveform.ReadWAVMono(path)
	if err != nil {
		return fmt.Errorf("read wav: %w", err)
	}
	if len(samples) == 0 {
		return fmt.Errorf("no audio data found")
	}

	switch mode {
	case "braille":
		return waveform.RenderBraille(samples, int(sr), height, out)
	case "ascii":
		return waveform.RenderASCII(samples, int(sr), height, out)
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
}