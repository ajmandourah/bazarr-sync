package tui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

func (a App) browserTitles() []string {
	// Return titles for fuzzy matching
	var titles []string
	switch a.mediaType {
	case "movie":
		for _, m := range a.movies {
			titles = append(titles, m.Title)
		}
	case "show":
		for _, s := range a.shows {
			titles = append(titles, s.Title)
		}
	}
	return titles
}

func (a App) filteredTitles() []string {
	all := a.browserTitles()
	if a.search == "" {
		return all
	}
	matches := fuzzy.Find(a.search, all)
	var filtered []string
	for _, m := range matches {
		filtered = append(filtered, m.Str)
	}
	return filtered
}

func (a App) BrowserView() string {
	var header string
	switch a.mediaType {
	case "movie":
		header = "Movies"
	case "show":
		header = "Shows"
	}

	var b strings.Builder
	b.WriteString(titleBar.Render("  " + header))
	
	if a.search != "" {
		b.WriteString("\n" + searchStyle.Render("  ▸ " + a.search + " [Escape clear]"))
	}
	
	titles := a.filteredTitles()
	if len(titles) == 0 {
		return panelStyle.Render(b.String() + "\n  No items found")
	}
	
	for i, t := range titles {
		if i == a.browserIdx {
			b.WriteString("\n" + itemSel.Render(">> " + t))
		} else {
			b.WriteString("\n" + itemUnsel.Render("   " + t))
		}
	}
	
	b.WriteString("\n\n" + barStyle.Render(fmt.Sprintf("  %d/%d", a.browserIdx+1, len(titles))))
	
	return lipgloss.Place(a.width, a.height, lipgloss.Top, lipgloss.Center, b.String())
}
