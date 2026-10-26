// SPDX-License-Identifier: AGPL-3.0-or-later

package game

import "testing"

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name, guess, answer string
		want                [WordLen]State
	}{
		{"perfect", "crane", "crane", [WordLen]State{Correct, Correct, Correct, Correct, Correct}},
		{"none", "doubt", "crane", [WordLen]State{Absent, Absent, Absent, Absent, Absent}},
		{
			// answer has two b's and one y: the third b in the guess has no
			// copy left, so it must be Absent, not Present.
			"repeated guess letter", "bobby", "abbey",
			[WordLen]State{Present, Absent, Correct, Absent, Correct},
		},
		{
			"repeated answer letters", "erase", "speed",
			[WordLen]State{Present, Absent, Absent, Present, Present},
		},
		{
			// the green e must not be consumed twice by the others.
			"green wins", "eerie", "every",
			[WordLen]State{Correct, Present, Present, Absent, Absent},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Evaluate(tc.guess, tc.answer); got != tc.want {
				t.Fatalf("Evaluate(%q, %q) = %v, want %v", tc.guess, tc.answer, got, tc.want)
			}
		})
	}
}

func TestPlay(t *testing.T) {
	g := New("abbey")
	g.Known = func(w string) bool { return w != "zzzzz" }

	if err := g.Play("zzzzz"); err != ErrNotAWord {
		t.Fatalf("unknown word: got %v, want ErrNotAWord", err)
	}
	if err := g.Play("toolong"); err != ErrLength {
		t.Fatalf("long word: got %v, want ErrLength", err)
	}
	if err := g.Play("bobby"); err != nil {
		t.Fatalf("valid word: %v", err)
	}
	if len(g.Guesses) != 1 || g.Status != Playing {
		t.Fatalf("got %d guesses, status %v", len(g.Guesses), g.Status)
	}

	if err := g.Play("abbey"); err != nil {
		t.Fatalf("winning word: %v", err)
	}
	if g.Status != Won {
		t.Fatalf("status = %v, want Won", g.Status)
	}
	if err := g.Play("crane"); err != ErrOver {
		t.Fatalf("after win: got %v, want ErrOver", err)
	}
}

func TestLose(t *testing.T) {
	g := New("abbey")
	for i := 0; i < MaxGuesses; i++ {
		if err := g.Play("crane"); err != nil {
			t.Fatalf("guess %d: %v", i, err)
		}
	}
	if g.Status != Lost {
		t.Fatalf("status = %v, want Lost", g.Status)
	}
}

func TestKeyboard(t *testing.T) {
	g := New("abbey")
	if err := g.Play("bobby"); err != nil {
		t.Fatal(err)
	}
	keys := g.Keyboard()
	if keys['b'] != Correct {
		t.Fatalf("b = %v, want Correct (best score wins)", keys['b'])
	}
	if keys['y'] != Correct {
		t.Fatalf("y = %v, want Correct", keys['y'])
	}
	if _, ok := keys['q']; ok {
		t.Fatal("q should be unknown")
	}
}

func TestReset(t *testing.T) {
	g := New("abbey")
	_ = g.Play("abbey")
	g.Reset("crane")
	if g.Status != Playing || len(g.Guesses) != 0 || g.Answer != "crane" {
		t.Fatalf("reset incomplete: %+v", g)
	}
}
