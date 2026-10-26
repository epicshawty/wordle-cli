// SPDX-License-Identifier: AGPL-3.0-or-later

package sound

import (
	"encoding/binary"
	"math"
	"math/rand"
	"sync"

	"github.com/epicshawty/wordle-cli/internal/game"
)

const (
	sampleRate = 44100
	// stagger is the delay between tiles of one guess.
	stagger = 45
)

// ms converts milliseconds to samples.
func ms(n int) int { return sampleRate * n / 1000 }

// silence returns a zeroed buffer of n milliseconds.
func silence(n int) []float64 { return make([]float64, ms(n)) }

// thock mixes one soft mechanical hit into buf: a damped body tone plus a
// short low-passed noise burst. Everything here is percussive on purpose —
// the cue set has no melody.
func thock(buf []float64, offset, dur int, hz, gain float64) {
	end := offset + ms(dur)
	if end > len(buf) {
		end = len(buf)
	}
	rnd := rand.New(rand.NewSource(int64(offset + dur*31)))
	var lp float64
	for i := offset; i < end; i++ {
		t := float64(i-offset) / sampleRate
		body := math.Sin(2*math.Pi*hz*t) * math.Exp(-t/0.012)
		lp += 0.28 * ((rnd.Float64()*2 - 1) - lp)
		noise := lp * math.Exp(-t/0.004) * 0.6
		buf[i] += gain * (0.85*body + noise)
	}
}

// shimmer is a quiet, slowly decaying high tone that punctuates a win.
func shimmer(buf []float64, offset, dur int, hz, gain float64) {
	end := offset + ms(dur)
	if end > len(buf) {
		end = len(buf)
	}
	for i := offset; i < end; i++ {
		t := float64(i-offset) / sampleRate
		buf[i] += gain * math.Sin(2*math.Pi*hz*t) * math.Exp(-t/0.09)
	}
}

// cueSamples renders one standalone cue.
func cueSamples(c Cue) []float64 {
	switch c {
	case Type:
		b := silence(40)
		thock(b, 0, 40, 300, 0.10)
		return b
	case Erase:
		b := silence(50)
		thock(b, 0, 50, 200, 0.11)
		return b
	case Invalid:
		b := silence(130)
		thock(b, 0, 60, 150, 0.12)
		thock(b, ms(55), 70, 120, 0.10)
		return b
	case Win:
		b := silence(360)
		thock(b, 0, 60, 320, 0.10)
		thock(b, ms(70), 70, 430, 0.10)
		shimmer(b, ms(150), 200, 880, 0.06)
		return b
	case Lose:
		b := silence(300)
		thock(b, 0, 90, 190, 0.11)
		thock(b, ms(90), 140, 130, 0.10)
		return b
	}
	return nil
}

// rowSamples renders a whole guess: one blip per tile at 45 ms intervals, so
// the row reads as a single cascade, plus an optional win or loss tail.
func rowSamples(marks [game.WordLen]game.State, final Cue) []float64 {
	length := stagger*(game.WordLen-1) + 60
	switch final {
	case Win:
		length += 200
	case Lose:
		length += 160
	}
	buf := silence(length)

	for i, m := range marks {
		switch m {
		case game.Correct:
			thock(buf, ms(i*stagger), 60, 480, 0.11)
		case game.Present:
			thock(buf, ms(i*stagger), 55, 340, 0.09)
		default:
			thock(buf, ms(i*stagger), 50, 180, 0.07)
		}
	}

	end := ms(stagger*(game.WordLen-1) + 70)
	switch final {
	case Win:
		shimmer(buf, end, 200, 880, 0.07)
	case Lose:
		thock(buf, end, 150, 120, 0.10)
	}
	return buf
}

// encode writes a 16-bit mono PCM WAV image.
func encode(s []float64) []byte {
	buf := make([]byte, 44+2*len(s))
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+2*len(s)))
	copy(buf[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)           // fmt chunk size
	binary.LittleEndian.PutUint16(buf[20:], 1)            // PCM
	binary.LittleEndian.PutUint16(buf[22:], 1)            // mono
	binary.LittleEndian.PutUint32(buf[24:], sampleRate)   // samples per second
	binary.LittleEndian.PutUint32(buf[28:], sampleRate*2) // bytes per second
	binary.LittleEndian.PutUint16(buf[32:], 2)            // block align
	binary.LittleEndian.PutUint16(buf[34:], 16)           // bits per sample
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(2*len(s)))
	for i, v := range s {
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		binary.LittleEndian.PutUint16(buf[44+2*i:], uint16(int16(v*32767)))
	}
	return buf
}

var (
	cacheMu  sync.Mutex
	wavCache = map[Cue][]byte{}
)

// cueWAV returns the encoded bytes for c, synthesising once.
func cueWAV(c Cue) []byte {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if w, ok := wavCache[c]; ok {
		return w
	}
	s := cueSamples(c)
	if s == nil {
		return nil
	}
	w := encode(s)
	wavCache[c] = w
	return w
}
