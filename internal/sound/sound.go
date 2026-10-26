// SPDX-License-Identifier: AGPL-3.0-or-later

// Package sound plays the short percussive cues behind the board.
//
// Standard library only. On Windows the WAV data goes straight to winmm; on
// macOS, Linux and the BSDs it hands the buffer to a system player. With no
// backend available every call is a silent no-op, so a headless machine or a
// CI runner never sees an error.
package sound

import (
	"sync"

	"github.com/epicshawty/wordle-cli/internal/game"
)

// Cue is one short sound.
type Cue int

const (
	// None is the absence of a tail sound.
	None Cue = iota - 1
	// Type is a letter landing on the board.
	Type
	// Erase is a letter taken back.
	Erase
	// Invalid is a guess the game refused.
	Invalid
	// Win is the flourish after a winning guess.
	Win
	// Lose is the thud after the last guess.
	Lose
)

// backend plays one complete WAV image and returns immediately.
type backend interface {
	play(wav []byte) error
	close() error
}

// Player plays cues. A nil Player is silent.
type Player struct {
	mu   sync.Mutex
	b    backend
	on   bool
	last []byte // keeps the async Windows buffer alive while it plays
}

// Open returns a player. A disabled player, or an environment with no audio
// backend, is silent and never errors.
func Open(enabled bool) *Player {
	p := &Player{on: enabled}
	if enabled {
		p.b = newBackend()
		p.on = p.b != nil
	}
	return p
}

// Enabled reports whether cues will make a sound.
func (p *Player) Enabled() bool { return p != nil && p.on && p.b != nil }

// Close releases the backend. Safe on nil and safe to call twice.
func (p *Player) Close() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.b != nil {
		_ = p.b.close()
	}
	p.b, p.on, p.last = nil, false, nil
}

// Play emits cue c.
func (p *Player) Play(c Cue) {
	if !p.Enabled() {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playLocked(cueWAV(c))
}

// PlayRow plays one guess: a blip per tile, staggered, pre-mixed into a single
// buffer so a one-channel backend cannot cut itself off. final adds a win or
// loss tail to the same buffer.
func (p *Player) PlayRow(marks [game.WordLen]game.State, final Cue) {
	if !p.Enabled() {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playLocked(encode(rowSamples(marks, final)))
}

func (p *Player) playLocked(wav []byte) {
	if len(wav) == 0 || p.b == nil {
		return
	}
	if err := p.b.play(wav); err != nil {
		return
	}
	p.last = wav
}
