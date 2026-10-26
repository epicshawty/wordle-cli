// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package sound

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	winmm         = syscall.NewLazyDLL("winmm.dll")
	procPlaySound = winmm.NewProc("PlaySoundW")
)

const (
	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndMemory    = 0x0004
)

// winBackend plays in-memory WAV images through winmm, with no cgo and no
// child process.
type winBackend struct{}

func newBackend() backend { return winBackend{} }

// play hands the buffer to PlaySound with SND_MEMORY|SND_ASYNC: it returns in
// about 5 ms and playback continues in the mixer. The buffer must outlive the
// call, which is why Player keeps the last image referenced.
func (winBackend) play(wav []byte) error {
	if len(wav) == 0 {
		return errors.New("empty wav")
	}
	ret, _, _ := procPlaySound.Call(
		uintptr(unsafe.Pointer(&wav[0])),
		0,
		sndMemory|sndAsync|sndNoDefault,
	)
	if ret == 0 {
		return errors.New("PlaySound failed")
	}
	return nil
}

func (winBackend) close() error { return nil }
