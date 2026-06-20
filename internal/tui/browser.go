package tui

import (
	"fmt"
	"strings"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

func (a App) browserTitles() []string {
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

func (a App) getSubtitleListFiltered(filteredIdx int) []bazarr.Subtitle {
	filtered := a.filteredTitles()
	if filteredIdx < 0 || filteredIdx >= len(filtered) {
		return nil
	}
	titleName := filtered[filteredIdx]
	for _, m := range a.movies {
		if m.Title == titleName {
			return m.Subtitles
		}
	}
	return nil
}

func (a App) isStaged(path string, code string) bool {
	for _, s := range a.staged {
		if s.Params.Path == path && s.Params.Lang == code {
			return true
		}
	}
	return false
}

func (a *App) toggleStage(sub bazarr.Subtitle, title string, mediaId int, code2 string) {
	for i, s := range a.staged {
		if s.Params.Path == sub.Path && s.Params.Lang == code2 {
			a.staged = append(a.staged[:i], a.staged[i+1:]...)
			return
		}
	}
	param := bazarr.GetSyncParams("movie", mediaId, sub)
	a.staged = append(a.staged, SyncJob{
		Params:   param,
		Title:    title,
		Language: code2,
	})
}

// renderSearchBar renders a permanent search bar at the top
func (a App) renderSearchBar() string {
	if a.focusSearch || a.search != "" {
		return searchBarActive.Render(" / " + a.search + " [Esc clear] ")
	}
	return searchBarInactive.Render(" / type to search... ")
}

// BrowserView - unified list view for movies and shows
func (a App) BrowserView() string {
	var header string
	switch a.mediaType {
	case "movie":
		header = "Movies"
	case "show":
		header = "Shows"
	}

	titles := a.filteredTitles()

	// Compute layout dimensions - account for staged panel dynamically
	headerHeight := 2 // titleBar + searchBar
	footerHeight := 2 // barStyle + cheatSheet
	panelPadding := 2 // panelStyle padding top+bottom
	stagedHeight := a.stagedPanelHeight()
	contentRows := a.height - headerHeight - footerHeight - panelPadding - stagedHeight
	if contentRows < 1 {
		contentRows = 1
	}

	var b strings.Builder
	b.WriteString(titleBar.Render("  " + header + "  |  Staged: " + fmt.Sprintf("%d", len(a.staged))))
	b.WriteString("\n" + a.renderSearchBar())

	if len(titles) == 0 {
		if a.loading {
			b.WriteString("\n" + itemUnsel.Render("  "+spinnerStr(a.frame)+" Loading items..."))
		} else {
			b.WriteString("\n" + itemUnsel.Render("  No items found"))
		}
		b.WriteString("\n" + cheatSheet.Render("  / search  •  Esc back  •  q quit"))
		return lipgloss.NewStyle().Background(base).Align(lipgloss.Center).Render(panelStyle.Render(b.String()))
	}

	if a.browserIdx >= len(titles) {
		a.browserIdx = len(titles) - 1
	}

	// Calculate scroll window
	halfWindow := contentRows / 2
	startIdx := a.browserIdx - halfWindow
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + contentRows
	if endIdx > len(titles) {
		startIdx = len(titles) - contentRows
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = len(titles)
	}

	// Determine max title width for truncation
	mainWidth := a.width
	if a.width >= 80 {
		mainWidth = a.width / 2
	}
	maxTitleLen := mainWidth - 10
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}

	for i := startIdx; i < endIdx; i++ {
		displayTitle := titles[i]
		if len(displayTitle) > maxTitleLen {
			displayTitle = displayTitle[:maxTitleLen] + ".."
		}
		if i == a.browserIdx {
			b.WriteString("\n" + itemSel.Render("  >> "+displayTitle))
		} else {
			b.WriteString("\n" + itemUnsel.Render("     "+displayTitle))
		}
	}

	b.WriteString("\n" + barStyle.Render(fmt.Sprintf("  %d / %d", a.browserIdx+1, len(titles))))

	if stg := a.renderStagedList(); stg != "" {
		b.WriteString("\n" + stagedPanel.Render(stg))
	}
	b.WriteString("\n" + cheatSheet.Render("  ↑↓ j/k navigate  •  Enter/Space subs  •  / search  •  Ctrl+S sync  •  Esc back  •  q quit"))

	return panelStyle.Render(b.String())
}

// MovieSubsView - subtitle selection for a selected movie
func (a App) MovieSubsView() string {
	titles := a.filteredTitles()
	if len(titles) == 0 {
		return ""
	}
	titleName := titles[a.browserIdx]
	subs := a.getSubtitleListFiltered(a.browserIdx)

	// Compute layout dimensions - account for staged panel dynamically
	headerHeight := 1 // titleBar only
	footerHeight := 2 // barStyle + cheatSheet
	panelPadding := 2 // panelStyle padding top+bottom
	stagedHeight := a.stagedPanelHeight()
	contentRows := a.height - headerHeight - footerHeight - panelPadding - stagedHeight
	if contentRows < 1 {
		contentRows = 1
	}

	var b strings.Builder
	b.WriteString(titleBar.Render("  Subtitles: " + titleName))

	if len(subs) == 0 {
		b.WriteString("\n" + itemUnsel.Render("  No subtitles available"))
		b.WriteString("\n" + cheatSheet.Render("  Esc back  •  Ctrl+S sync staged  •  q quit"))
		return lipgloss.NewStyle().Background(base).Align(lipgloss.Center).Render(panelStyle.Render(b.String()))
	}

	if a.selIdx >= len(subs) {
		a.selIdx = 0
	}

	// Calculate scroll window
	halfWindow := contentRows / 2
	startIdx := a.selIdx - halfWindow
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + contentRows
	if endIdx > len(subs) {
		startIdx = len(subs) - contentRows
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = len(subs)
	}

	for i := startIdx; i < endIdx; i++ {
		check := "[ ]"
		if a.isStaged(subs[i].Path, subs[i].Code2) {
			check = "[x]"
		}
		if i == a.selIdx {
			b.WriteString("\n" + itemSel.Render("  >> "+check+" "+subs[i].Code2))
		} else {
			b.WriteString("\n" + itemUnsel.Render("     "+check+" "+subs[i].Code2))
		}
	}

	b.WriteString("\n" + barStyle.Render(fmt.Sprintf("  %d / %d", a.selIdx+1, len(subs))))

	if stg := a.renderStagedList(); stg != "" {
		b.WriteString("\n" + stagedPanel.Render(stg))
	}
	b.WriteString("\n" + cheatSheet.Render("  ↑↓ j/k navigate  •  Space toggle  •  a stage all  •  S clear all  •  Ctrl+S sync  •  Esc back  •  q quit"))

	return panelStyle.Render(b.String())
}
