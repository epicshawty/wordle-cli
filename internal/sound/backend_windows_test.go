// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package sound

import (
	"os"
	"testing"
	"time"

	"github.com/epicshawty/wordle-cli/internal/game"
)

// TestWindowsBackend makes a real noise, so it only runs when asked:
//
//	WORDLE_SOUND_TEST=1 go test ./internal/sound -run WindowsBackend -v
//
// It is the only check that proves winmm accepted the buffer; the rest of the
// suite is silent on purpose.
func TestWindowsBackend(t *testing.T) {
	if os.Getenv("WORDLE_SOUND_TEST") != "1" {
		t.Skip("set WORDLE_SOUND_TEST=1 to play a real cue")
	}
	b := newBackend()
	if b == nil {
		t.Fatal("no backend on windows")
	}
	defer b.close()

	if err := b.play(cueWAV(Type)); err != nil {
		t.Fatalf("winmm rejected the type cue: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	row := encode(rowSamples([game.WordLen]game.State{
		game.Correct, game.Present, game.Absent, game.Present, game.Correct,
	}, Win))
	if err := b.play(row); err != nil {
		t.Fatalf("winmm rejected the row buffer: %v", err)
	}
	time.Sleep(400 * time.Millisecond)
}
