package tui

import (
	"strings"
	"github.com/charmbracelet/lipgloss"
)

var menuItems = []string{"Sync Movies", "Sync Shows", "Exit"}

func (a App) MenuView() string {
	asciiH := lipgloss.Height(asciiArt)
	subH := lipgloss.Height(subtitleStyle.Render(Tagline))
	availH := a.height - asciiH - subH - 10
	topPad := availH / 3
	if topPad < 1 {
		topPad = 2
	}

	var b strings.Builder
	for i := 0; i < topPad; i++ {
		b.WriteString("\n")
	}

	b.WriteString(titleStyle.Render(asciiArt))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render(Tagline))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		if i == a.menuIdx {
			b.WriteString(menuSel.Render(item))
		} else {
			b.WriteString(menuUnsel.Render(item))
		}
		if i < len(menuItems)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n" + footerStyle.Render("↑↓ navigate  Enter select  q quit"))

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, b.String())
}
