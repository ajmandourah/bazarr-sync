package tui

import (
	"fmt"
	"strings"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a App) HandleKey(msg tea.KeyMsg) (App, tea.Cmd) {
	switch a.screen {
	case ScreenMenu:
		return a.MenuHandler(msg)
	case ScreenBrowser:
		return a.BrowserHandler(msg)
	case ScreenSelecting:
		return a.SelectionHandler(msg)
	case ScreenSyncing:
		if msg.Type == tea.KeyEnter {
			// Done - show results
			var okCount, failCount int
			for _, r := range a.results {
				if r == "ok" {
					okCount++
				} else {
					failCount++
				}
			}
			a.summary = fmt.Sprintf("%d synced, %d failed", okCount, failCount)
			a.screen = ScreenDone
		}
	case ScreenDone:
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEscape || msg.String() == "q" {
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) MenuHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if a.menuIdx > 0 {
			a.menuIdx--
		}
	case tea.KeyDown:
		if a.menuIdx < len(menuItems)-1 {
			a.menuIdx++
		}
	case tea.KeyEnter:
		switch a.menuIdx {
		case 0:
			a.screen = ScreenBrowser
			a.mediaType = "movie"
			a.browserIdx = 0
			return a, func() tea.Msg {
				c := config.GetConfig()
				d, err := bazarr.QueryMovies(c)
				return DataMessage{Movies: d.Data, Errors: err}
			}
		case 1:
			a.screen = ScreenBrowser
			a.mediaType = "show"
			a.browserIdx = 0
			return a, func() tea.Msg {
				c := config.GetConfig()
				d, err := bazarr.QuerySeries(c)
				return DataMessage{Shows: d.Data, Errors: err}
			}
		case 2:
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) BrowserHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	count := len(a.movies)
	if a.mediaType == "show" {
		count = len(a.shows)
	}

	if a.focusSearch {
		switch msg.Type {
		case tea.KeyEscape:
			a.focusSearch = false
			a.search = ""
		case tea.KeyBackspace:
			if len(a.search) > 0 {
				a.search = a.search[:len(a.search)-1]
			}
		case tea.KeyRunes:
			a.search += string(msg.Runes)
		default:
			a.focusSearch = false
			return a.BrowserHandler(msg)
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if a.browserIdx > 0 {
			a.browserIdx--
		}
	case tea.KeyDown:
		if a.browserIdx < count-1 {
			a.browserIdx++
		}
	case tea.KeyRunes:
		a.focusSearch = true
	case tea.KeyEnter:
		a.screen = ScreenSelecting
		a.selIdx = 0
		a.items = a.buildItems()
	case tea.KeyEscape:
		if msg.String() == "q" {
			return a, tea.Quit
		}
		a.screen = ScreenMenu
	}
	return a, nil
}

func (a App) SelectionHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if a.selIdx > 0 {
			a.selIdx--
		}
	case tea.KeyDown:
		if a.selIdx < len(a.items)-1 {
			a.selIdx++
		}
	case tea.KeySpace:
		if a.selIdx < len(a.items) {
			it := a.items[a.selIdx]
			it.Selected = !it.Selected
			a.items[a.selIdx] = it
		}
	case tea.KeyEnter:
		var sel []SelectItem
		for _, it := range a.items {
			if it.Selected {
				sel = append(sel, it)
			}
		}
		if len(sel) > 0 {
			a.jobs = nil
			a.results = make([]string, len(sel))
			for i, it := range sel {
				a.jobs = append(a.jobs, SyncJob{
					Params: bazarr.SyncParams{
						Action:   "sync",
						Path:     it.Path,
						Id:       it.MediaId,
						Lang:     it.Code2,
						Type:     it.MediaType,
						Gss:      "False",
						NoFramerateFix: "False",
					},
					Title:    it.Title,
					Language: it.Code2,
				})
				a.results[i] = "running"
			}
			a.screen = ScreenSyncing
			return a, syncJobs(a.jobs, a.cfg)
		}
	case tea.KeyEscape:
		a.screen = ScreenBrowser
	}
	return a, nil
}

func (a App) HandleData(msg DataMessage) (App, tea.Cmd) {
	switch a.mediaType {
	case "movie":
		a.movies = msg.Movies
	case "show":
		a.shows = msg.Shows
	}
	return a, nil
}

func (a App) HandleSyncResult(msg SyncResultMessage) (App, tea.Cmd) {
	if msg.Index < len(a.results) {
		a.results[msg.Index] = "ok"
	}
	return a, nil
}

func (a App) buildItems() []SelectItem {
	switch a.mediaType {
	case "movie":
		if a.browserIdx >= len(a.movies) {
			return nil
		}
		return buildFromMovie(a.movies[a.browserIdx])
	case "show":
		if a.browserIdx >= len(a.shows) {
			return nil
		}
		return buildFromShow(a.shows[a.browserIdx])
	}
	return nil
}

func buildFromMovie(m bazarr.Movie) []SelectItem {
	var items []SelectItem
	for _, sub := range m.Subtitles {
		items = append(items, SelectItem{
			Title:     m.Title,
			Subtitle:  "lang: " + sub.Code2,
			MediaType: "movie",
			MediaId:   m.RadarrId,
			Path:      sub.Path,
			Code2:     sub.Code2,
		})
	}
	return items
}

func buildFromShow(s bazarr.Show) []SelectItem {
	var items []SelectItem
	return items
}

func syncJobs(jobs []SyncJob, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if len(jobs) == 0 {
			return SyncResultMessage{Index: -1}
		}
		job := jobs[0]
		ok := bazarr.Sync(cfg, job.Params)
		return SyncResultMessage{
			Index:   0,
			Success: ok,
			Title:   job.Title,
		}
	}
}


func (a App) SelectingView() string {
	var count int
	for _, it := range a.items {
		if it.Selected {
			count++
		}
	}

	var b strings.Builder
	b.WriteString(titleBar.Render(fmt.Sprintf("  Select Subtitles [%d]", count)))

	for i, item := range a.items {
		if i == a.selIdx {
			b.WriteString("\n>> " + itemSel.Render(item.Title + " " + item.Subtitle))
		} else {
			b.WriteString("\n   " + itemUnsel.Render(item.Title + " " + item.Subtitle))
		}
	}

	b.WriteString("\n" + footerStyle.Render("↑↓ nav  Space select  Enter sync  Esc back"))
	return lipgloss.Place(a.width, a.height, lipgloss.Top, lipgloss.Center, b.String())
}

func (a App) SyncingView() string {
	var b strings.Builder
	b.WriteString(titleBar.Render("  Syncing"))

	for i, job := range a.jobs {
		status := spinnerStr(a.frame)
		if i < len(a.results) && a.results[i] != "running" {
			if a.results[i] == "ok" {
				status = syncSuccess.Render("✓")
			} else {
				status = syncError.Render("✗")
			}
		}
		b.WriteString("\n  " + status + " " + itemUnsel.Render(job.Title))
	}

	b.WriteString("\n\n" + footerStyle.Render("Syncing... press Enter when done"))
	return lipgloss.Place(a.width, a.height, lipgloss.Top, lipgloss.Center, b.String())
}

func (a App) DoneView() string {
	var b strings.Builder
	b.WriteString(titleBar.Render("  Complete"))
	b.WriteString("\n\n  " + subtitleStyle.Render(a.summary))
	b.WriteString("\n\n" + footerStyle.Render("Press Enter to exit"))

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, b.String())
}
