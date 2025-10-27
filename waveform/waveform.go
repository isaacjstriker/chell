package waveform

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/go-audio/wav"
	"golang.org/x/term"
)

// ReadWAVMono opens a WAV file and returns mono float64 samples normalized to [-1,1] and the sample rate.
func ReadWAVMono(path string) ([]float64, uint32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return nil, 0, fmt.Errorf("invalid WAV file")
	}

	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, 0, err
	}

	numCh := dec.Format().NumChannels
	bits := dec.BitDepth
	if bits == 0 {
		// default to 16 if unknown
		bits = 16
	}

	if buf == nil || len(buf.Data) == 0 {
		return nil, dec.SampleRate, nil
	}

	// mix to mono and convert to float64 in [-1,1]
	frameCount := len(buf.Data) / numCh
	out := make([]float64, frameCount)
	denom := int(1 << (bits - 1)) // signed PCM
	for i := 0; i < frameCount; i++ {
		sum := 0
		for c := 0; c < numCh; c++ {
			sum += buf.Data[i*numCh+c]
		}
		avg := float64(sum) / float64(numCh)
		out[i] = avg / float64(denom)
	}

	normalize(out)
	return out, dec.SampleRate, nil
}

func normalize(s []float64) {
	maxAbs := 0.0
	for _, v := range s {
		if a := math.Abs(v); a > maxAbs {
			maxAbs = a
		}
	}
	if maxAbs <= 0 {
		return
	}
	for i := range s {
		s[i] /= maxAbs
	}
}

// RenderBraille renders waveform using Unicode Braille characters.
// height is the number of character rows (each char row = 4 subpixels vertically).
func RenderBraille(samples []float64, sampleRate, charRows int, w io.Writer) error {
	cols, rows, err := term.GetSize(0)
	if err != nil {
		// fallback to typical size
		cols, rows = 80, 24
	}

	if charRows <= 0 || charRows > rows-1 {
		charRows = rows - 1
		if charRows < 1 {
			charRows = 1
		}
	}

	pixelHeight := charRows * 4
	pixelWidth := cols * 2

	grid := make([][]byte, charRows)
	for i := range grid {
		grid[i] = make([]byte, cols) // bitmask for braille dots
	}

	// Map samples into pixelWidth columns by averaging blocks
	total := len(samples)
	if total == 0 {
		return nil
	}
	block := total / pixelWidth
	if block < 1 {
		block = 1
	}

	// Precompute positions
	center := float64(pixelHeight-1) / 2.0
	positions := make([]int, pixelWidth)
	for x := 0; x < pixelWidth; x++ {
		start := x * block
		if start >= total {
			positions[x] = int(math.Round(center))
			continue
		}
		end := start + block
		if end > total {
			end = total
		}
		sum := 0.0
		for i := start; i < end; i++ {
			sum += samples[i]
		}
		avg := sum / float64(end-start)
		// map to pixel row (0..pixelHeight-1), top=0
		y := center - avg*(float64(pixelHeight)/2.0)
		if y < 0 {
			y = 0
		}
		if y > float64(pixelHeight-1) {
			y = float64(pixelHeight - 1)
		}
		positions[x] = int(math.Round(y))
	}

	// Braille bit layout:
	// rows (0..3) x cols (0..1) -> bit value
	bits := [][]byte{
		{0x01, 0x08}, // row0: dot1, dot4
		{0x02, 0x10}, // row1: dot2, dot5
		{0x04, 0x20}, // row2: dot3, dot6
		{0x40, 0x80}, // row3: dot7, dot8
	}

	for px := 0; px < pixelWidth; px++ {
		xChar := px / 2
		xSub := px % 2
		yIndex := positions[px]
		charRow := yIndex / 4
		subrow := yIndex % 4
		if charRow < 0 {
			charRow = 0
		}
		if charRow >= charRows {
			charRow = charRows - 1
		}
		grid[charRow][xChar] |= bits[subrow][xSub]
	}

	// Render
	base := rune(0x2800)
	for r := 0; r < charRows; r++ {
		line := make([]rune, cols)
		for c := 0; c < cols; c++ {
			val := grid[r][c]
			if val == 0 {
				line[c] = ' '
			} else {
				line[c] = base + rune(val)
			}
		}
		if _, err := fmt.Fprintln(w, string(line)); err != nil {
			return err
		}
	}
	return nil
}

// RenderASCII renders a simple ASCII/blocks waveform (centered).
func RenderASCII(samples []float64, sampleRate, charRows int, w io.Writer) error {
	cols, rows, err := term.GetSize(0)
	if err != nil {
		cols, rows = 80, 24
	}
	if charRows <= 0 || charRows > rows-1 {
		charRows = rows - 1
		if charRows < 1 {
			charRows = 1
		}
	}

	// We'll allocate rows x cols canvas
	canvasRows := charRows
	canvasCols := cols

	// chunk samples into cols
	total := len(samples)
	chunk := total / canvasCols
	if chunk < 1 {
		chunk = 1
	}
	peaks := make([]float64, canvasCols)
	for c := 0; c < canvasCols; c++ {
		start := c * chunk
		if start >= total {
			peaks[c] = 0
			continue
		}
		end := start + chunk
		if end > total {
			end = total
		}
		maxv := 0.0
		for i := start; i < end; i++ {
			if a := math.Abs(samples[i]); a > maxv {
				maxv = a
			}
		}
		peaks[c] = maxv
	}

	mid := canvasRows / 2
	// build canvas
	canvas := make([][]rune, canvasRows)
	for r := range canvas {
		canvas[r] = make([]rune, canvasCols)
		for c := 0; c < canvasCols; c++ {
			canvas[r][c] = ' '
		}
	}

	for c := 0; c < canvasCols; c++ {
		val := peaks[c]
		half := mid
		n := int(math.Round(val * float64(half)))
		if n == 0 {
			if mid >= 0 && mid < canvasRows {
				canvas[mid][c] = '·'
			}
			continue
		}
		for y := mid - n; y <= mid+n; y++ {
			if y >= 0 && y < canvasRows {
				canvas[y][c] = '█'
			}
		}
	}

	for r := 0; r < canvasRows; r++ {
		if _, err := fmt.Fprintln(w, string(canvas[r])); err != nil {
			return err
		}
	}
	return nil
}
