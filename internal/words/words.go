// SPDX-License-Identifier: AGPL-3.0-or-later

// Package words is the embedded dictionary: 2,315 common answers and the
// 20,141 five-letter words accepted as guesses.
//
// The split is Wordle's own: answers stay fair and guessable, while any real
// English five-letter word — inflections and rare words included — is allowed
// as a guess.
//
// guesses.txt is sorted, so membership is a binary search over the embedded
// blob: no map and no per-word allocation at startup.
package words

import (
	_ "embed"
	"errors"
	"math/rand/v2"
	"sort"
	"strings"
)

//go:embed answers.txt
var rawAnswers string

//go:embed guesses.txt
var rawGuesses string

var (
	answers = strings.Fields(rawAnswers)
	guesses = strings.Fields(rawGuesses) // sorted; Valid relies on it
)

// Answers returns every word that can be the answer, in file order.
func Answers() []string { return answers }

// Len reports how many words can be the answer.
func Len() int { return len(answers) }

// GuessCount reports how many words are accepted as guesses.
func GuessCount() int { return len(guesses) }

// Valid reports whether w is an accepted guess.
func Valid(w string) bool {
	i := sort.SearchStrings(guesses, w)
	return i < len(guesses) && guesses[i] == w
}

// IsAnswer reports whether w can be the answer to a round. Every answer is
// also a valid guess.
func IsAnswer(w string) bool {
	for _, a := range answers {
		if a == w {
			return true
		}
	}
	return false
}

// Random returns a uniformly random answer.
func Random() string { return answers[rand.IntN(len(answers))] }

// Pick returns forced when it is an accepted guess, otherwise a random answer.
// A forced word does not have to be in the answer list, which is what makes
// --word useful for testing.
func Pick(forced string) (string, error) {
	w := strings.ToLower(strings.TrimSpace(forced))
	if w == "" {
		return Random(), nil
	}
	if len(w) != 5 || !Valid(w) {
		return "", errors.New("wordle: " + forced + " is not a five-letter word in the dictionary")
	}
	return w, nil
}
