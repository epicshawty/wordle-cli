// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows && !darwin && !linux && !freebsd && !netbsd && !openbsd

package sound

// newBackend returns nil where we have no known player: the game stays silent
// rather than failing.
func newBackend() backend { return nil }
