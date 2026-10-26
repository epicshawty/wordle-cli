// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ui renders the interactive board with bubbletea and lipgloss.
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/epicshawty/wordle-cli/internal/game"
	"github.com/epicshawty/wordle-cli/internal/sound"
	"github.com/epicshawty/wordle-cli/internal/words"
)

// minWidth and minHeight are the space needed to draw the on-screen keyboard.
const (
	minWidth  = 40
	minHeight = 20
)

// Model is the interactive state of one session.
type Model struct {
	g     *game.Game
	snd   *sound.Player
	typed []byte
	msg   string
	bad   bool
	help  bool
	w, h  int
}

// New wires a round to the dictionary and to the sound player and returns the
// model. snd may be nil.
func New(g *game.Game, snd *sound.Player) Model {
	g.Known = words.Valid
	return Model{g: g, snd: snd}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m Model) key(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.help {
			m.help = false
			return m, nil
		}
		return m, tea.Quit
	case "?":
		m.help = !m.help
		return m, nil
	case "enter":
		return m.submit()
	case "backspace":
		if len(m.typed) > 0 {
			m.typed = m.typed[:len(m.typed)-1]
			m.msg, m.bad = "", false
			m.snd.Play(sound.Erase)
		}
		return m, nil
	case "q":
		if m.g.Status != game.Playing {
			return m, tea.Quit
		}
	case "r":
		if m.g.Status != game.Playing {
			m.g.Reset(words.Random())
			m.typed, m.msg, m.bad = nil, "", false
			return m, nil
		}
	}

	if len(k.Runes) == 1 {
		c := k.Runes[0]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c >= 'a' && c <= 'z' && m.g.Status == game.Playing && len(m.typed) < game.WordLen {
			m.typed = append(m.typed, byte(c))
			m.msg, m.bad = "", false
			m.snd.Play(sound.Type)
		}
	}
	return m, nil
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	if m.g.Status != game.Playing {
		return m, nil
	}
	if len(m.typed) < game.WordLen {
		m.msg, m.bad = "Not enough letters", true
		m.snd.Play(sound.Invalid)
		return m, nil
	}
	if err := m.g.Play(string(m.typed)); err != nil {
		m.msg, m.bad = err.Error(), true
		m.snd.Play(sound.Invalid)
		return m, nil
	}

	m.typed, m.bad = nil, false
	final := sound.None
	switch m.g.Status {
	case game.Won:
		m.msg = praise(len(m.g.Guesses))
		final = sound.Win
	case game.Lost:
		m.msg = "The word was " + strings.ToUpper(m.g.Answer)
		final = sound.Lose
	default:
		m.msg = ""
	}
	m.snd.PlayRow(m.g.Marks[len(m.g.Marks)-1], final)
	return m, nil
}

func praise(guesses int) string {
	switch guesses {
	case 1:
		return "Genius"
	case 2:
		return "Magnificent"
	case 3:
		return "Impressive"
	case 4:
		return "Splendid"
	case 5:
		return "Great"
	default:
		return "Phew"
	}
}

// View implements tea.Model.
func (m Model) View() string {
	var parts []string
	if m.fits(24) {
		parts = append(parts, wordmark(), "")
		if m.fits(30) {
			parts = append(parts, taglineStyle.Render("six guesses · five letters"), "")
		}
	}
	parts = append(parts, m.board())
	switch {
	case m.help:
		parts = append(parts, "", help())
	case m.fits(28) && m.wide():
		parts = append(parts, "", m.keyboard())
	}
	if m.fits(20) {
		parts = append(parts, "", m.status())
	}

	body := lipgloss.JoinVertical(lipgloss.Center, parts...)
	if m.w <= 0 {
		return body
	}
	if m.h > 0 && lipgloss.Height(body) < m.h {
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, body)
	}
	return lipgloss.Place(m.w, 0, lipgloss.Center, lipgloss.Top, body)
}

// fits reports whether the terminal is at least n rows tall (or unknown).
func (m Model) fits(n int) bool { return m.h == 0 || m.h >= n }

// wide reports whether the terminal is wide enough for the keyboard.
func (m Model) wide() bool { return m.w == 0 || m.w >= minWidth }

func wordmark() string {
	const word = "WORDLE"
	letters := make([]string, len(word))
	for i := 0; i < len(word); i++ {
		letters[i] = lipgloss.NewStyle().Bold(true).Foreground(wordmarkColors[i]).Render(string(word[i]))
	}
	return strings.Join(letters, " ")
}

func (m Model) board() string {
	rows := make([]string, 0, game.MaxGuesses)
	for i := 0; i < game.MaxGuesses; i++ {
		rows = append(rows, m.row(i))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) row(i int) string {
	cells := make([]string, 0, game.WordLen*2)
	for j := 0; j < game.WordLen; j++ {
		if j > 0 {
			cells = append(cells, " ")
		}
		switch {
		case i < len(m.g.Guesses):
			cells = append(cells, tile(letter(m.g.Guesses[i][j]), m.g.Marks[i][j], true, false))
		case i == len(m.g.Guesses) && j < len(m.typed):
			cells = append(cells, tile(letter(m.typed[j]), game.Absent, false, j == len(m.typed)-1))
		default:
			cursor := i == len(m.g.Guesses) && j == len(m.typed) && m.g.Status == game.Playing
			cells = append(cells, tile("", game.Absent, false, cursor))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

// letter renders a guess byte as the uppercase letter shown on the board.
func letter(b byte) string { return string(b - 'a' + 'A') }

func tile(letter string, score game.State, scored, active bool) string {
	switch {
	case scored:
		c := stateColor(score)
		return tileBase.Background(c).BorderForeground(c).Foreground(colorText).Bold(true).Render(letter)
	case letter != "":
		b := colorFilled
		if active {
			b = colorCursor
		}
		return tileBase.BorderForeground(b).Foreground(colorText).Bold(true).Render(letter)
	case active:
		return tileBase.BorderForeground(colorCursor).Render(" ")
	default:
		return tileBase.Render(" ")
	}
}

func (m Model) keyboard() string {
	known := m.g.Keyboard()
	rows := []string{
		keyRow(known, "qwertyuiop"),
		"",
		keyRow(known, "asdfghjkl"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			capKey("ENTER"), " ", keyRow(known, "zxcvbnm"), " ", capKey("BKSP")),
	}
	return lipgloss.JoinVertical(lipgloss.Center, rows...)
}

// capKey renders a wide action key such as enter or bksp.
func capKey(label string) string {
	return lipgloss.NewStyle().
		Background(colorCursor).
		Foreground(colorText).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

func keyRow(known map[byte]game.State, letters string) string {
	cells := make([]string, 0, len(letters)*2)
	for i := 0; i < len(letters); i++ {
		if i > 0 {
			cells = append(cells, " ")
		}
		style := lipgloss.NewStyle().
			Background(colorCursor).
			Foreground(colorText).
			Bold(true).
			Padding(0, 1)
		if st, ok := known[letters[i]]; ok {
			style = style.Background(stateColor(st))
		}
		cells = append(cells, style.Render(letter(letters[i])))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func (m Model) status() string {
	hint := "enter submit · ? help · ctrl+c quit"
	if m.g.Status != game.Playing {
		hint = "r new word · q quit"
	}
	if m.msg == "" {
		return hintStyle.Render(hint)
	}
	if m.bad {
		return badStyle.Render(m.msg)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		goodStyle.Render(m.msg), hintStyle.Render("   "+hint))
}

func help() string {
	return helpStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		"a-z       type a letter",
		"enter     submit guess",
		"bksp      delete letter",
		"?         toggle this help",
		"r         new word, once the round ends",
		"q  esc    quit",
		"",
		"green     right letter, right spot",
		"yellow    right letter, wrong spot",
		"grey      letter not in the word",
	))
}
