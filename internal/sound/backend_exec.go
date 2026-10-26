// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build darwin || linux || freebsd || netbsd || openbsd

package sound

import (
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// candidates in preference order; the first one on PATH wins.
var candidates = [][]string{
	{"afplay"},      // macOS
	{"paplay"},      // PulseAudio / PipeWire
	{"pw-play"},     // PipeWire
	{"aplay", "-q"}, // ALSA
}

// execBackend shells out to a system player. One temp file per distinct cue is
// written once and reused, so a keystroke costs only a fork.
type execBackend struct {
	bin []string
	dir string
	mu  sync.Mutex
}

func newBackend() backend {
	var bin []string
	for _, c := range candidates {
		if path, err := exec.LookPath(c[0]); err == nil {
			bin = append([]string{path}, c[1:]...)
			break
		}
	}
	if bin == nil {
		return nil
	}
	dir, err := os.MkdirTemp("", "wordle-sound")
	if err != nil {
		return nil
	}
	return &execBackend{bin: bin, dir: dir}
}

func (b *execBackend) play(wav []byte) error {
	if len(wav) == 0 {
		return errors.New("empty wav")
	}
	name := filepath.Join(b.dir, fmt.Sprintf("%08x.wav", crc32.ChecksumIEEE(wav)))
	if _, err := os.Stat(name); err != nil {
		if err := os.WriteFile(name, wav, 0o600); err != nil {
			return err
		}
	}
	args := append(append([]string{}, b.bin[1:]...), name)
	// The player outlives this call; reap it in the background.
	return exec.Command(b.bin[0], args...).Start()
}

func (b *execBackend) close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return os.RemoveAll(b.dir)
}
