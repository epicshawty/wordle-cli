// SPDX-License-Identifier: AGPL-3.0-or-later

// Package words is the embedded dictionary: 2,315 common five-letter words,
// used both as answers and as accepted guesses.
package words

import (
	_ "embed"
	"errors"
	"math/rand/v2"
	"strings"
)

//go:embed words.txt
var raw string

var (
	list = strings.Fields(raw)
	set  = make(map[string]struct{}, len(list))
)

func init() {
	for _, w := range list {
		set[w] = struct{}{}
	}
}

// All returns every word, in file order.
func All() []string { return list }

// Len reports how many words the dictionary holds.
func Len() int { return len(list) }

// Valid reports whether w is an accepted guess.
func Valid(w string) bool {
	_, ok := set[w]
	return ok
}

// Random returns a uniformly random answer.
func Random() string { return list[rand.IntN(len(list))] }

// Pick returns forced when it is a word in the list, otherwise a random one.
func Pick(forced string) (string, error) {
	w := strings.ToLower(strings.TrimSpace(forced))
	if w == "" {
		return Random(), nil
	}
	if len(w) != 5 || !Valid(w) {
		return "", errors.New("wordle: " + forced + " is not a five-letter word in the list")
	}
	return w, nil
}
