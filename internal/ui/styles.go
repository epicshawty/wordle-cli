// SPDX-License-Identifier: AGPL-3.0-or-later

package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/epicshawty/wordle-cli/internal/game"
)

// The Wordle palette.
var (
	colorGreen  = lipgloss.Color("#538d4e")
	colorYellow = lipgloss.Color("#b59f3b")
	colorGray   = lipgloss.Color("#3a3a3c")
	colorBorder = lipgloss.Color("#3a3a3c")
	colorFilled = lipgloss.Color("#565758")
	colorCursor = lipgloss.Color("#818384")
	colorText   = lipgloss.Color("#f8f8f8")
	colorDim    = lipgloss.Color("#818384")
)

// tileBase is one 7x3 cell of the board.
var tileBase = lipgloss.NewStyle().
	Width(1).
	Height(1).
	Padding(0, 2).
	Align(lipgloss.Center, lipgloss.Center).
	Border(lipgloss.NormalBorder()).
	BorderForeground(colorBorder)

var (
	taglineStyle = lipgloss.NewStyle().Foreground(colorDim)
	hintStyle    = lipgloss.NewStyle().Foreground(colorDim)
	goodStyle    = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	badStyle     = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	helpStyle    = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorFilled).
			Padding(0, 2).
			Foreground(colorDim)
	wordmarkColors = []lipgloss.Color{colorGreen, colorGreen, colorYellow, colorYellow, colorGreen, colorGreen}
)

func stateColor(s game.State) lipgloss.Color {
	switch s {
	case game.Correct:
		return colorGreen
	case game.Present:
		return colorYellow
	default:
		return colorGray
	}
}
