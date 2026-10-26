// SPDX-License-Identifier: AGPL-3.0-or-later

// Package plain plays a round on stdin/stdout with no TUI and no colour.
// It is used for --plain, for pipes, and by the smoke test.
package plain

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/epicshawty/wordle-cli/internal/game"
)

// Run plays g until it ends or stdin is exhausted, and returns an exit code.
func Run(g *game.Game, in io.Reader, out io.Writer) int {
	fmt.Fprintln(out, "Wordle: six guesses, five letters.")
	fmt.Fprintln(out, "# right spot   + wrong spot   . absent")
	fmt.Fprintln(out)

	sc := bufio.NewScanner(in)
	for g.Status == game.Playing {
		fmt.Fprintf(out, "%d> ", len(g.Guesses)+1)
		if !sc.Scan() {
			fmt.Fprintln(out)
			return 1
		}
		word := strings.ToLower(strings.TrimSpace(sc.Text()))
		if word == "" {
			continue
		}
		if err := g.Play(word); err != nil {
			fmt.Fprintf(out, "   %v\n", err)
			continue
		}
		fmt.Fprintf(out, "   %s\n", render(g.Guesses[len(g.Guesses)-1], g.Marks[len(g.Marks)-1]))
	}

	if g.Status == game.Won {
		fmt.Fprintf(out, "\nGot it in %d/%d.\n", len(g.Guesses), game.MaxGuesses)
		return 0
	}
	fmt.Fprintf(out, "\nOut of guesses. The word was %s.\n", strings.ToUpper(g.Answer))
	return 1
}

func render(word string, marks [game.WordLen]game.State) string {
	letters := make([]string, game.WordLen)
	scores := make([]string, game.WordLen)
	for i := 0; i < game.WordLen; i++ {
		letters[i] = strings.ToUpper(string(word[i]))
		switch marks[i] {
		case game.Correct:
			scores[i] = "#"
		case game.Present:
			scores[i] = "+"
		default:
			scores[i] = "."
		}
	}
	return strings.Join(letters, " ") + "   " + strings.Join(scores, " ")
}
