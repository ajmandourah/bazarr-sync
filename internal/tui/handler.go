package tui

import (
	"fmt"
	"strings"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func isSpaceKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeySpace || msg.String() == " "
}

func isEnterKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEnter || msg.String() == "\r"
}

func (a App) HandleKey(msg tea.KeyMsg) (App, tea.Cmd) {
	// Vim bindings: h/j/k/l → arrow keys (only for single runes, never q or space)
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'h', 'k':
			msg.Type = tea.KeyUp
			msg.Runes = nil
		case 'j', 'l':
			msg.Type = tea.KeyDown
			msg.Runes = nil
		}
	}

	switch a.screen {
	case ScreenMenu:
		return a.MenuHandler(msg)
	case ScreenBrowser:
		return a.BrowserHandler(msg)
	case ScreenMovieSubs:
		return a.MovieSubHandler(msg)
	case ScreenShowEpisodes:
		return a.ShowEpisodeHandler(msg)
	case ScreenEpisodeSubs:
		return a.EpisodeSubHandler(msg)
	case ScreenSyncing:
		// Wait for sync
	case ScreenDone:
		if isEnterKey(msg) {
			a.screen = ScreenMenu
			a.staged = make([]SyncJob, 0)
			a.jobs = nil
			a.results = nil
			a.summary = ""
			return a, nil
		}
		if msg.Type == tea.KeyEscape || msg.String() == "q" {
			return a, tea.Quit
		}
	case ScreenConfig:
		return a.ConfigHandler(msg)
	}
	return a, nil
}

// --- MENU ---

func (a App) MenuHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if isEnterKey(msg) {
		switch a.menuIdx {
		case 0:
			a.screen = ScreenBrowser
			a.mediaType = "movie"
			a.browserIdx = 0
			a.selIdx = 0
			a.staged = make([]SyncJob, 0)
			a.search = ""
			if len(a.movies) == 0 {
				a.loading = true
				return a, loadMovies()
			}
			a.loading = false
		case 1:
			a.screen = ScreenBrowser
			a.mediaType = "show"
			a.browserIdx = 0
			a.selectedShow = bazarr.Show{}
			a.episodes = nil
			a.staged = make([]SyncJob, 0)
			a.search = ""
			if len(a.shows) == 0 {
				a.loading = true
				return a, loadShowList()
			}
			a.loading = false
		case 2:
			a.screen = ScreenConfig
			a.populateConfigFields()
			a.cfgIdx = 0
			a.cfgValidationResult = ""
			a.cfgValidationSuccess = false
			a.cfgValidating = false
		case 3:
			return a, tea.Quit
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if a.menuIdx > 0 {
			a.menuIdx--
		}
	case tea.KeyDown:
		if a.menuIdx < len(menuItems)-1 {
			a.menuIdx++
		}
	case tea.KeyEscape:
		return a, tea.Quit
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			return a, tea.Quit
		}
	}
	return a, nil
}

func loadMovies() tea.Cmd {
	return func() tea.Msg {
		c := config.GetConfig()
		d, err := bazarr.QueryMovies(c)
		return DataMessage{Movies: d.Data, Errors: err}
	}
}

func loadShowList() tea.Cmd {
	return func() tea.Msg {
		c := config.GetConfig()
		d, err := bazarr.QuerySeries(c)
		return DataMessage{Shows: d.Data, Errors: err}
	}
}

func loadEpisodes(seriesId int) tea.Cmd {
	return func() tea.Msg {
		c := config.GetConfig()
		d, err := bazarr.QueryEpisodes(c, seriesId)
		return EpisodeMessage{Episodes: d.Data, Errors: err}
	}
}

func syncStagedCmd(staged []SyncJob) tea.Cmd {
	return func() tea.Msg {
		c := config.GetConfig()
		var results []SyncResultMessage
		for i, job := range staged {
			ok := bazarr.Sync(c, job.Params)
			results = append(results, SyncResultMessage{
				Index:   i,
				Success: ok,
				Title:   job.Title,
			})
		}
		return results
	}
}

// --- BROWSER ---

func (a App) BrowserHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if msg.Type == tea.KeyCtrlS {
		if len(a.staged) > 0 {
			a.screen = ScreenSyncing
			a.jobs = make([]SyncJob, len(a.staged))
			copy(a.jobs, a.staged)
			a.results = make([]string, len(a.staged))
			for i := range a.results {
				a.results[i] = "running"
			}
			return a, syncStagedCmd(a.staged)
		}
		return a, nil
	}

	if a.focusSearch {
		switch msg.Type {
		case tea.KeyEscape:
			a.focusSearch = false
			a.search = ""
			a.browserIdx = 0
			return a, nil
		case tea.KeyBackspace:
			if len(a.search) > 0 {
				a.search = a.search[:len(a.search)-1]
			}
			return a, nil
		}
		if isEnterKey(msg) {
			a.focusSearch = false
			a.browserIdx = 0
			return a, nil
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			a.search += string(msg.Runes)
			return a, nil
		}
		return a, nil
	}

	if a.search != "" {
		switch msg.Type {
		case tea.KeyEscape:
			a.search = ""
			a.browserIdx = 0
			return a, nil
		}
	}

	titles := a.filteredTitles()

	// Handle Space to enter subtitle selection
	if isSpaceKey(msg) {
		if a.mediaType == "movie" {
			a.screen = ScreenMovieSubs
			a.selIdx = 0
		} else if a.mediaType == "show" {
			if a.browserIdx < len(a.shows) {
				a.screen = ScreenShowEpisodes
				a.selectedShow = a.shows[a.browserIdx]
				a.episodeIdx = 0
				a.episodes = nil
				a.epLoading = true
				return a, loadEpisodes(a.selectedShow.SonarrId)
			}
		}
		return a, nil
	}

	if isEnterKey(msg) {
		if a.mediaType == "movie" {
			a.screen = ScreenMovieSubs
			a.selIdx = 0
		} else if a.mediaType == "show" {
			if a.browserIdx < len(a.shows) {
				a.screen = ScreenShowEpisodes
				a.selectedShow = a.shows[a.browserIdx]
				a.episodeIdx = 0
				a.episodes = nil
				a.epLoading = true
				return a, loadEpisodes(a.selectedShow.SonarrId)
			}
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if a.browserIdx > 0 {
			a.browserIdx--
		}
	case tea.KeyDown:
		if a.browserIdx < len(titles)-1 {
			a.browserIdx++
		}
	case tea.KeyEscape:
		a.screen = ScreenMenu
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == '/' {
			a.focusSearch = true
		}
		if len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			return a, tea.Quit
		}
	}
	return a, nil
}

// --- MOVIE SUBTITLES ---

func (a App) MovieSubHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if msg.Type == tea.KeyCtrlS {
		if len(a.staged) > 0 {
			a.screen = ScreenSyncing
			a.jobs = make([]SyncJob, len(a.staged))
			copy(a.jobs, a.staged)
			a.results = make([]string, len(a.staged))
			for i := range a.results {
				a.results[i] = "running"
			}
			return a, syncStagedCmd(a.staged)
		}
		return a, nil
	}

	titles := a.filteredTitles()
	if len(titles) == 0 {
		return a, nil
	}
	subs := a.getSubtitleListFiltered(a.browserIdx)
	if subs == nil {
		return a, nil
	}

	titleName := titles[a.browserIdx]
	var movieId int
	for _, m := range a.movies {
		if m.Title == titleName {
			movieId = m.RadarrId
			break
		}
	}

	if isSpaceKey(msg) {
		if a.selIdx < len(subs) {
			sub := subs[a.selIdx]
			a.toggleStage(sub, titleName, movieId, sub.Code2)
			if a.selIdx >= len(subs) {
				a.selIdx = len(subs) - 1
			}
			if a.selIdx < 0 {
				a.selIdx = 0
			}
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if a.selIdx > 0 {
			a.selIdx--
		}
	case tea.KeyDown:
		if a.selIdx < len(subs)-1 {
			a.selIdx++
		}
	case tea.KeyEscape:
		a.screen = ScreenBrowser
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'q':
				return a, tea.Quit
			case 'a':
				for _, sub := range subs {
					if !a.isStaged(sub.Path, sub.Code2) {
						a.toggleStage(sub, titleName, movieId, sub.Code2)
					}
				}
				return a, nil
			case 'S':
				a.staged = make([]SyncJob, 0)
				return a, nil
			}
		}
	}
	return a, nil
}

// --- SHOW EPISODES ---

func (a App) ShowEpisodeHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if msg.Type == tea.KeyCtrlS {
		if len(a.staged) > 0 {
			a.screen = ScreenSyncing
			a.jobs = make([]SyncJob, len(a.staged))
			copy(a.jobs, a.staged)
			a.results = make([]string, len(a.staged))
			for i := range a.results {
				a.results[i] = "running"
			}
			return a, syncStagedCmd(a.staged)
		}
		return a, nil
	}

	if isEnterKey(msg) {
		if len(a.episodes) > 0 && a.episodeIdx < len(a.episodes) {
			ep := a.episodes[a.episodeIdx]
			a.screen = ScreenEpisodeSubs
			a.items = buildItemsFromEpisode(a.selectedShow, ep)
			a.selIdx = 0
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if len(a.episodes) > 0 && a.episodeIdx > 0 {
			a.episodeIdx--
		}
	case tea.KeyDown:
		if a.episodeIdx < len(a.episodes)-1 {
			a.episodeIdx++
		}
	case tea.KeyEscape:
		a.screen = ScreenBrowser
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a App) ShowEpisodesView() string {
	headerHeight := 1 // titleBar
	footerHeight := 2 // barStyle + cheatSheet
	panelPadding := 2 // panelStyle padding top+bottom
	stagedHeight := a.stagedPanelHeight()
	contentRows := a.height - headerHeight - footerHeight - panelPadding - stagedHeight
	if contentRows < 1 {
		contentRows = 1
	}

	var b strings.Builder
	b.WriteString(titleBar.Render(fmt.Sprintf("  Episodes: %s", a.selectedShow.Title)))

	if a.epLoading || len(a.episodes) == 0 {
		b.WriteString("\n" + itemUnsel.Render("  "+spinnerStr(a.frame)+" Loading episodes..."))
		b.WriteString("\n" + cheatSheet.Render("  Waiting for data..."))
		return lipgloss.NewStyle().Background(base).Align(lipgloss.Center).Render(panelStyle.Render(b.String()))
	}

	// Calculate scroll window
	halfWindow := contentRows / 2
	startIdx := a.episodeIdx - halfWindow
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + contentRows
	if endIdx > len(a.episodes) {
		startIdx = len(a.episodes) - contentRows
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = len(a.episodes)
	}

	var list strings.Builder
	mainWidth := a.width
	if a.width >= 80 {
		mainWidth = a.width / 2
	}
	maxLabelLen := mainWidth - 10
	if maxLabelLen < 10 {
		maxLabelLen = 10
	}

	for i := startIdx; i < endIdx; i++ {
		ep := a.episodes[i]
		subCount := len(ep.Subtitles)
		label := fmt.Sprintf("S%02dE%02d - %s [%d subs]", ep.SeasonNumber, ep.EpisodeNumber, ep.Title, subCount)
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen] + ".."
		}
		if i == a.episodeIdx {
			list.WriteString("\n" + itemSel.Render("  >> "+label))
		} else {
			list.WriteString("\n" + itemUnsel.Render("     "+label))
		}
	}

	content := list.String() + "\n" + barStyle.Render(fmt.Sprintf("  %d / %d", a.episodeIdx+1, len(a.episodes)))
	b.WriteString(content)

	if stg := a.renderStagedList(); stg != "" {
		b.WriteString("\n" + stagedPanel.Render(stg))
	}
	b.WriteString("\n" + cheatSheet.Render("  ↑↓ h/j/k/l navigate  •  Enter select subs  •  Ctrl+S sync staged  •  Esc back  •  q quit"))

	return panelStyle.Render(b.String())
}

// --- EPISODE SUBTITLES ---

func buildItemsFromEpisode(show bazarr.Show, ep bazarr.Episode) []SelectItem {
	var items []SelectItem
	for _, sub := range ep.Subtitles {
		items = append(items, SelectItem{
			Title:     fmt.Sprintf("%s - S%02dE%02d - %s", show.Title, ep.SeasonNumber, ep.EpisodeNumber, ep.Title),
			Subtitle:  sub.Code2,
			MediaType: "series",
			MediaId:   show.SonarrId,
			Path:      sub.Path,
			Code2:     sub.Code2,
			Selected:  false,
		})
	}
	return items
}

func (a App) EpisodeSubHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if msg.Type == tea.KeyCtrlS {
		if len(a.staged) > 0 {
			a.screen = ScreenSyncing
			a.jobs = make([]SyncJob, len(a.staged))
			copy(a.jobs, a.staged)
			a.results = make([]string, len(a.staged))
			for i := range a.results {
				a.results[i] = "running"
			}
			return a, syncStagedCmd(a.staged)
		}
		return a, nil
	}

	if isSpaceKey(msg) {
		if a.selIdx < len(a.items) {
			item := a.items[a.selIdx]
			a.toggleEpisodeSub(item)
			if a.selIdx >= len(a.items) {
				a.selIdx = len(a.items) - 1
			}
			if a.selIdx < 0 {
				a.selIdx = 0
			}
		}
		return a, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		if a.selIdx > 0 {
			a.selIdx--
		}
	case tea.KeyDown:
		if a.selIdx < len(a.items)-1 {
			a.selIdx++
		}
	case tea.KeyEscape:
		a.screen = ScreenShowEpisodes
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'q':
				return a, tea.Quit
			case 'a':
				for _, item := range a.items {
					if !a.isStaged(item.Path, item.Code2) {
						a.toggleEpisodeSub(item)
					}
				}
				return a, nil
			case 'S':
				a.staged = make([]SyncJob, 0)
				return a, nil
			}
		}
	}
	return a, nil
}

func (a *App) toggleEpisodeSub(item SelectItem) {
	for i, s := range a.staged {
		if s.Params.Path == item.Path && s.Params.Lang == item.Code2 {
			a.staged = append(a.staged[:i], a.staged[i+1:]...)
			return
		}
	}
	param := bazarr.GetSyncParams(item.MediaType, item.MediaId, bazarr.Subtitle{Path: item.Path, Code2: item.Code2})
	a.staged = append(a.staged, SyncJob{
		Params:   param,
		Title:    item.Title,
		Language: item.Code2,
	})
}

func (a App) EpisodeSubsView() string {
	headerHeight := 1 // titleBar
	footerHeight := 2 // barStyle + cheatSheet
	panelPadding := 2 // panelStyle padding top+bottom
	stagedHeight := a.stagedPanelHeight()
	contentRows := a.height - headerHeight - footerHeight - panelPadding - stagedHeight
	if contentRows < 1 {
		contentRows = 1
	}

	var b strings.Builder
	b.WriteString(titleBar.Render("  Select Subtitles"))

	if len(a.items) == 0 {
		b.WriteString("\n" + itemUnsel.Render("  No subtitles available"))
		b.WriteString("\n" + cheatSheet.Render("  Esc back  •  Ctrl+S sync staged"))
		return lipgloss.NewStyle().Background(base).Align(lipgloss.Center).Render(panelStyle.Render(b.String()))
	}

	if a.selIdx >= len(a.items) {
		a.selIdx = 0
	}

	// Calculate scroll window
	halfWindow := contentRows / 2
	startIdx := a.selIdx - halfWindow
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + contentRows
	if endIdx > len(a.items) {
		startIdx = len(a.items) - contentRows
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = len(a.items)
	}

	mainWidth := a.width
	if a.width >= 80 {
		mainWidth = a.width / 2
	}
	maxLabelLen := mainWidth - 10
	if maxLabelLen < 10 {
		maxLabelLen = 10
	}

	for i := startIdx; i < endIdx; i++ {
		item := a.items[i]
		check := "[ ]"
		if a.isStaged(item.Path, item.Code2) {
			check = "[x]"
		}
		label := item.Title + " [" + item.Subtitle + "]"
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen] + ".."
		}
		if i == a.selIdx {
			b.WriteString("\n" + itemSel.Render("  >> "+check+" "+label))
		} else {
			b.WriteString("\n" + itemUnsel.Render("     "+check+" "+label))
		}
	}

	b.WriteString("\n" + barStyle.Render(fmt.Sprintf("  %d / %d", a.selIdx+1, len(a.items))))

	if stg := a.renderStagedList(); stg != "" {
		b.WriteString("\n" + stagedPanel.Render(stg))
	}
	b.WriteString("\n" + cheatSheet.Render("  ↑↓ h/j/k/l navigate  •  Space toggle  •  a stage all  •  S clear all  •  Ctrl+S sync  •  Esc back  •  q quit"))

	return panelStyle.Render(b.String())
}

// --- SYNC VIEWS ---

func (a App) SyncingView() string {
	var b strings.Builder
	b.WriteString(titleBar.Render("  Syncing Subtitles"))

	for i, job := range a.jobs {
		var status string
		if i < len(a.results) && a.results[i] != "running" && a.results[i] != "pending" {
			if a.results[i] == "ok" {
				status = syncSuccess.Render("✓")
			} else {
				status = syncError.Render("✗")
			}
		} else {
			status = syncRunning.Render(spinnerStr(a.frame))
		}
		b.WriteString("\n" + itemUnsel.Render("  "+status+" "+job.Title+" ["+job.Language+"]"))
	}

	b.WriteString("\n" + footerStyle.Render("  Syncing..."))

	return lipgloss.NewStyle().Background(base).Align(lipgloss.Center).Render(panelStyle.Render(b.String()))
}

func (a App) DoneView() string {
	var b strings.Builder
	b.WriteString(titleBar.Render("  Complete"))

	var okCount, failCount int
	for _, r := range a.results {
		if r == "ok" {
			okCount++
		} else {
			failCount++
		}
	}
	a.summary = fmt.Sprintf("%d synced, %d failed", okCount, failCount)

	b.WriteString("\n\n  " + subtitleStyle.Render(a.summary))
	b.WriteString("\n\n" + cheatSheet.Render("  Enter back to menu  •  q quit"))

	return panelStyle.Render(b.String())
}

// --- UPDATE HANDLERS ---

func (a App) HandleData(msg DataMessage) (App, tea.Cmd) {
	a.loading = false
	switch a.mediaType {
	case "movie":
		a.movies = msg.Movies
	case "show":
		a.shows = msg.Shows
	}
	if msg.Errors != nil {
		a.summary = "Error: " + msg.Errors.Error()
	}
	return a, nil
}

func (a App) HandleEpisodes(msg EpisodeMessage) (App, tea.Cmd) {
	a.episodes = msg.Episodes
	a.epLoading = false
	if msg.Errors != nil {
		a.summary = "Error loading episodes: " + msg.Errors.Error()
	}
	return a, nil
}

func (a App) HandleSyncResult(msg SyncResultMessage) (App, tea.Cmd) {
	if msg.Index >= 0 && msg.Index < len(a.results) {
		if msg.Success {
			a.results[msg.Index] = "ok"
		} else {
			a.results[msg.Index] = "fail"
		}
	}
	allDone := true
	for _, r := range a.results {
		if r == "running" || r == "pending" {
			allDone = false
			break
		}
	}
	if allDone {
		a.screen = ScreenDone
	}
	return a, nil
}

func (a App) handleBatchResults(results []SyncResultMessage) (App, tea.Cmd) {
	for _, r := range results {
		if r.Index >= 0 && r.Index < len(a.results) {
			if r.Success {
				a.results[r.Index] = "ok"
			} else {
				a.results[r.Index] = "fail"
			}
		}
	}
	a.screen = ScreenDone
	return a, nil
}

func (a App) HandleConfigResult(msg ConfigResult) (App, tea.Cmd) {
	a.cfgValidating = false
	a.cfgValidationResult = ""
	a.cfgValidationSuccess = false

	if msg.Success {
		a.cfgValidationSuccess = true
		a.cfgValidationResult = "Config saved!"
	} else {
		a.cfgValidationResult = msg.Error
	}
	return a, nil
}
