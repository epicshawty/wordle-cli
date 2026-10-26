// SPDX-License-Identifier: AGPL-3.0-or-later

// Package game holds the rules of Wordle: scoring, guesses, game state.
// It knows nothing about terminals, word lists or randomness.
package game

import "errors"

const (
	// WordLen is the number of letters per word.
	WordLen = 5
	// MaxGuesses is how many tries a round allows.
	MaxGuesses = 6
)

// State is the score of one letter of a guess.
type State uint8

const (
	// Absent means the letter does not occur in the answer.
	Absent State = iota
	// Present means the letter occurs in the answer, elsewhere.
	Present
	// Correct means the letter is in the right position.
	Correct
)

// Status is the state of a round.
type Status uint8

const (
	Playing Status = iota
	Won
	Lost
)

var (
	// ErrLength is returned for guesses that are not five a-z letters.
	ErrLength = errors.New("needs 5 letters")
	// ErrNotAWord is returned for guesses outside the dictionary.
	ErrNotAWord = errors.New("not in word list")
	// ErrOver is returned when the round is already finished.
	ErrOver = errors.New("game is over")
)

// Game is one round of Wordle.
type Game struct {
	Answer  string
	Guesses []string
	Marks   [][WordLen]State
	Status  Status

	// Known reports whether a word may be guessed. Optional; nil accepts
	// any five letters.
	Known func(string) bool
}

// New starts a round with the given answer.
func New(answer string) *Game {
	return &Game{Answer: answer}
}

// Reset starts a fresh round with a new answer.
func (g *Game) Reset(answer string) {
	g.Answer = answer
	g.Guesses = nil
	g.Marks = nil
	g.Status = Playing
}

// Play scores guess w and advances the round.
func (g *Game) Play(w string) error {
	if g.Status != Playing {
		return ErrOver
	}
	if len(w) != WordLen {
		return ErrLength
	}
	for i := 0; i < len(w); i++ {
		if w[i] < 'a' || w[i] > 'z' {
			return ErrLength
		}
	}
	if g.Known != nil && !g.Known(w) {
		return ErrNotAWord
	}

	g.Guesses = append(g.Guesses, w)
	g.Marks = append(g.Marks, Evaluate(w, g.Answer))

	if w == g.Answer {
		g.Status = Won
	} else if len(g.Guesses) == MaxGuesses {
		g.Status = Lost
	}
	return nil
}

// Keyboard maps letters already tried to their best known score.
func (g *Game) Keyboard() map[byte]State {
	out := make(map[byte]State, len(g.Guesses)*WordLen)
	for i, guess := range g.Guesses {
		for j := 0; j < WordLen; j++ {
			c, st := guess[j], g.Marks[i][j]
			if prev, ok := out[c]; !ok || st > prev {
				out[c] = st
			}
		}
	}
	return out
}

// Evaluate scores guess against answer, handling repeated letters correctly:
// a letter is only marked Present while an unmatched copy remains in answer.
func Evaluate(guess, answer string) [WordLen]State {
	var out [WordLen]State
	var left [26]int

	for i := 0; i < WordLen; i++ {
		if guess[i] == answer[i] {
			out[i] = Correct
		} else {
			left[answer[i]-'a']++
		}
	}
	for i := 0; i < WordLen; i++ {
		if out[i] == Correct {
			continue
		}
		c := guess[i] - 'a'
		if left[c] > 0 {
			out[i] = Present
			left[c]--
		}
	}
	return out
}
