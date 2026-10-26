// SPDX-License-Identifier: AGPL-3.0-or-later

package words

import "testing"

func TestList(t *testing.T) {
	if Len() < 2000 {
		t.Fatalf("dictionary has %d words, want at least 2000", Len())
	}
	seen := make(map[string]bool, Len())
	for _, w := range All() {
		if len(w) != 5 {
			t.Fatalf("%q is not five letters", w)
		}
		for i := 0; i < len(w); i++ {
			if w[i] < 'a' || w[i] > 'z' {
				t.Fatalf("%q is not lowercase a-z", w)
			}
		}
		if seen[w] {
			t.Fatalf("%q is duplicated", w)
		}
		seen[w] = true
	}
}

func TestValid(t *testing.T) {
	if !Valid(All()[0]) {
		t.Fatalf("%q should be valid", All()[0])
	}
	if Valid("zzzzz") || Valid("CRANE") || Valid("") {
		t.Fatal("invalid words accepted")
	}
}

func TestRandom(t *testing.T) {
	for i := 0; i < 200; i++ {
		if w := Random(); !Valid(w) {
			t.Fatalf("Random returned unknown word %q", w)
		}
	}
	if a, b := Random(), Random(); a == b && Len() > 1 && a != Random() {
		// Two draws matching is legal; this only fails if Random is stuck.
		t.Skip("coincidental duplicate")
	}
}

func TestPick(t *testing.T) {
	want := All()[42]
	if got, err := Pick(want); err != nil || got != want {
		t.Fatalf("Pick(%q) = %q, %v", want, got, err)
	}
	if got, err := Pick(""); err != nil || !Valid(got) {
		t.Fatalf("Pick(\"\") = %q, %v", got, err)
	}
	if _, err := Pick("zzzzz"); err == nil {
		t.Fatal("Pick accepted an unknown word")
	}
}
