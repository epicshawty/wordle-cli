// SPDX-License-Identifier: AGPL-3.0-or-later

package sound

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/epicshawty/wordle-cli/internal/game"
)

func TestCuesAreValidWAV(t *testing.T) {
	for _, c := range []Cue{Type, Erase, Invalid, Win, Lose} {
		wav := cueWAV(c)
		if len(wav) < 44 {
			t.Fatalf("cue %d: %d bytes, want a WAV header", c, len(wav))
		}
		if string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" || string(wav[36:40]) != "data" {
			t.Fatalf("cue %d: bad RIFF/WAVE/data markers", c)
		}
		if got := binary.LittleEndian.Uint32(wav[24:]); got != sampleRate {
			t.Fatalf("cue %d: sample rate %d, want %d", c, got, sampleRate)
		}
		if got := binary.LittleEndian.Uint32(wav[40:]); int(got) != len(wav)-44 {
			t.Fatalf("cue %d: data chunk %d, payload %d", c, got, len(wav)-44)
		}
		if peak(wav) == 0 {
			t.Fatalf("cue %d is silent", c)
		}
		if c == Type && ms(40)*2 != len(wav)-44 {
			t.Fatalf("type cue: %d payload bytes, want %d", len(wav)-44, ms(40)*2)
		}
	}
}

func TestPeakIsSane(t *testing.T) {
	for _, c := range []Cue{Type, Erase, Invalid, Win, Lose} {
		if p := peak(cueWAV(c)); p > 12000 || p < 800 {
			t.Fatalf("cue %d peak %d outside a pleasant range", c, p)
		}
	}
}

func TestRowSamples(t *testing.T) {
	marks := [game.WordLen]game.State{game.Correct, game.Present, game.Absent, game.Present, game.Correct}

	plain := rowSamples(marks, None)
	win := rowSamples(marks, Win)
	lose := rowSamples(marks, Lose)

	if len(win) <= len(plain) || len(lose) <= len(plain) {
		t.Fatalf("tails should extend the buffer: plain=%d win=%d lose=%d", len(plain), len(win), len(lose))
	}
	if len(plain) < ms(stagger*game.WordLen) {
		t.Fatalf("row buffer %d samples is shorter than the stagger", len(plain))
	}
	for name, s := range map[string][]float64{"plain": plain, "win": win, "lose": lose} {
		var max float64
		for _, v := range s {
			max = math.Max(max, math.Abs(v))
		}
		if max == 0 {
			t.Fatalf("%s row is silent", name)
		}
		if max > 1 {
			t.Fatalf("%s row clips at %f", name, max)
		}
	}
}

func TestSilentPlayerIsSafe(t *testing.T) {
	p := Open(false)
	if p.Enabled() {
		t.Fatal("disabled player reports enabled")
	}
	p.Play(Type)
	p.PlayRow([game.WordLen]game.State{}, Win)
	p.Close()

	var nilPlayer *Player
	if nilPlayer.Enabled() {
		t.Fatal("nil player reports enabled")
	}
	nilPlayer.Play(Erase)
	nilPlayer.PlayRow([game.WordLen]game.State{}, Lose)
	nilPlayer.Close()
}

// peak returns the largest absolute 16-bit sample in a WAV image.
func peak(wav []byte) int {
	max := 0
	for i := 44; i+1 < len(wav); i += 2 {
		v := int(int16(binary.LittleEndian.Uint16(wav[i:])))
		if v < 0 {
			v = -v
		}
		if v > max {
			max = v
		}
	}
	return max
}
