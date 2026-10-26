// SPDX-License-Identifier: AGPL-3.0-or-later

// Command wordle is a small Wordle clone for the terminal: one random
// five-letter answer, six guesses, no runtime dependencies.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/epicshawty/wordle-cli/internal/game"
	"github.com/epicshawty/wordle-cli/internal/plain"
	"github.com/epicshawty/wordle-cli/internal/sound"
	"github.com/epicshawty/wordle-cli/internal/ui"
	"github.com/epicshawty/wordle-cli/internal/words"
)

// version is set at build time with -X main.version.
var version = "dev"

func main() {
	var (
		showVersion = flag.Bool("version", false, "print version and exit")
		plainMode   = flag.Bool("plain", false, "line-based mode: no TUI, no colour")
		noSound     = flag.Bool("no-sound", false, "mute the sound cues")
		answer      = flag.String("word", "", "play a specific five-letter `word`")
	)
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("wordle %s\n", version)
		return
	}

	pick, err := words.Pick(*answer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	g := game.New(pick)
	g.Known = words.Valid
	if *plainMode || !isTerminal(os.Stdout) {
		os.Exit(plain.Run(g, os.Stdin, os.Stdout))
	}

	snd := sound.Open(!*noSound && soundEnabled())
	defer snd.Close()
	if _, err := tea.NewProgram(ui.New(g, snd), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "wordle:", err)
		os.Exit(1)
	}
}

// soundEnabled reports whether WORDLE_SOUND allows audio. Anything except an
// explicit off value leaves sound on.
func soundEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("WORDLE_SOUND"))) {
	case "0", "false", "no", "off":
		return false
	}
	return true
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprint(out, `wordle — a five-letter word game for the terminal.

usage: wordle [flags]

flags:
  --word WORD   play a specific answer instead of a random one
  --plain       line-based mode, no TUI and no colour
  --no-sound    mute the sound cues (also WORDLE_SOUND=0)
  --version     print the version and exit
  -h, --help    print this help

keys: a-z type · enter submit · backspace erase · ? help · ctrl+c quit
`)
}
