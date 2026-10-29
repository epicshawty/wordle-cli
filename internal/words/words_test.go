// SPDX-License-Identifier: AGPL-3.0-or-later

package words

import (
	"sort"
	"strings"
	"testing"
)

func TestLists(t *testing.T) {
	if Len() < 2000 {
		t.Fatalf("only %d answers", Len())
	}
	if GuessCount() < 15000 {
		t.Fatalf("only %d accepted guesses", GuessCount())
	}

	seen := make(map[string]bool, GuessCount())
	for _, w := range strings.Fields(rawGuesses) {
		if len(w) != 5 {
			t.Fatalf("guess %q is not five letters", w)
		}
		if seen[w] {
			t.Fatalf("guess %q duplicated", w)
		}
		seen[w] = true
	}

	// Every answer has to be playable.
	for _, a := range Answers() {
		if !Valid(a) {
			t.Fatalf("answer %q is not an accepted guess", a)
		}
		if !IsAnswer(a) {
			t.Fatalf("answer %q not recognised by IsAnswer", a)
		}
	}
	if IsAnswer("zzzzz") || IsAnswer("fjord") {
		t.Fatal("IsAnswer accepted a non-answer")
	}
}

// TestGuessesCoverRealWords pins the cases that used to be rejected while the
// dictionary was filtered against dictionary headwords.
func TestGuessesCoverRealWords(t *testing.T) {
	real := []string{
		"beers", "mails", "cared", "joked", "waved", "cooks", // inflections
		"fjord", "lymph", "nymph", "ouija", "xenon", "yodel", "yucky", "zonal", "axiom", "quirk", // rare but real
	}
	for _, w := range real {
		if !Valid(w) {
			t.Errorf("%q should be accepted as a guess", w)
		}
	}
	for _, w := range []string{"zzzzz", "CRANE", "aaaaa", ""} {
		if Valid(w) {
			t.Errorf("%q should not be accepted", w)
		}
	}
}

func TestRandom(t *testing.T) {
	for i := 0; i < 200; i++ {
		if w := Random(); !IsAnswer(w) {
			t.Fatalf("Random returned %q, which is not an answer", w)
		}
	}
}

func TestPick(t *testing.T) {
	want := Answers()[42]
	if got, err := Pick(want); err != nil || got != want {
		t.Fatalf("Pick(%q) = %q, %v", want, got, err)
	}
	if got, err := Pick(""); err != nil || !IsAnswer(got) {
		t.Fatalf("Pick(\"\") = %q, %v", got, err)
	}
	// A valid guess that is not an answer is allowed, for testing.
	if got, err := Pick("fjord"); err != nil || got != "fjord" {
		t.Fatalf("Pick(fjord) = %q, %v", got, err)
	}
	if _, err := Pick("zzzzz"); err == nil {
		t.Fatal("Pick accepted an unknown word")
	}
}

// TestGuessesSorted guards the binary search in Valid.
func TestGuessesSorted(t *testing.T) {
	if !sort.StringsAreSorted(guesses) {
		t.Fatal("guesses.txt is not sorted; Valid relies on it")
	}
}
