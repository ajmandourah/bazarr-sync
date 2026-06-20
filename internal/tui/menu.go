package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var menuItems = []string{"Sync Movies", "Sync Shows", "Settings", "Exit"}

var (
	menuBoxStyle = lipgloss.NewStyle().
			Background(base).
			PaddingTop(1).
			PaddingBottom(1).
			PaddingLeft(3).
			PaddingRight(3)
	menuItemSelStyle = lipgloss.NewStyle().
				Foreground(pink).
				Bold(true).Background(base)
	menuItemUnselStyle = lipgloss.NewStyle().
				Foreground(subtext0).Background(base)
)

var menuIcons = []string{"🎬", "📺", "⚙️", "🚪"}

func (a App) MenuView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(asciiArt))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render(Tagline))
	b.WriteString("\n\n")

	var items []string
	for i, item := range menuItems {
		icon := menuIcons[i]
		if i == a.menuIdx {
			items = append(items, menuItemSelStyle.Render("▶ "+icon+" "+item))
		} else {
			items = append(items, menuItemUnselStyle.Render("  "+icon+" "+item))
		}
	}

	menuContent := strings.Join(items, "\n")
	b.WriteString(menuBoxStyle.Render(menuContent))
	b.WriteString("\n\n")

	b.WriteString(cheatSheet.Render("  ↑↓ h/j/k/l  •  Enter select  •  q quit"))

	return b.String()
}
