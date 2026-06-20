package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	ScreenMenu         int = iota
	ScreenBrowser          // Movie list or Show list
	ScreenMovieSubs        // Subtitles of selected movie
	ScreenShowEpisodes     // Episodes of selected show
	ScreenEpisodeSubs      // Subtitles of selected episode
	ScreenSyncing
	ScreenDone
	ScreenConfig
)

type SyncJob struct {
	Params   bazarr.SyncParams
	Title    string
	Language string
}

type SelectItem struct {
	Title     string
	Subtitle  string
	MediaType string
	MediaId   int
	Path      string
	Code2     string
	Selected  bool
}

type DataMessage struct {
	Movies []bazarr.Movie
	Shows  []bazarr.Show
	Errors error
}

type EpisodeMessage struct {
	Episodes []bazarr.Episode
	Errors   error
}

type SyncResultMessage struct {
	Index   int
	Success bool
	Title   string
}

type TickMessage struct{}

func (a App) Init() tea.Cmd {
	return func() tea.Msg { return TickMessage{} }
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		return a.HandleKey(msg)
	case DataMessage:
		return a.HandleData(msg)
	case EpisodeMessage:
		return a.HandleEpisodes(msg)
	case SyncResultMessage:
		return a.HandleSyncResult(msg)
	case []SyncResultMessage:
		return a.handleBatchResults(msg)
	case ConfigResult:
		return a.HandleConfigResult(msg)
	case TickMessage:
		a.frame++
		return a, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return TickMessage{} })
	}
	return a, nil
}

func (a App) View() string {
	var content string
	switch a.screen {
	case ScreenMenu:
		content = a.MenuView()
	case ScreenBrowser:
		content = a.BrowserView()
	case ScreenMovieSubs:
		content = a.MovieSubsView()
	case ScreenShowEpisodes:
		content = a.ShowEpisodesView()
	case ScreenEpisodeSubs:
		content = a.EpisodeSubsView()
	case ScreenSyncing:
		content = a.SyncingView()
	case ScreenDone:
		content = a.DoneView()
	case ScreenConfig:
		content = a.ConfigView()
	}
	contentHeight := lipgloss.Height(content)
	topPad := (a.height - contentHeight) / 2
	if topPad < 0 {
		topPad = 0
	}
	bottomPad := a.height - contentHeight - topPad
	if bottomPad < 0 {
		bottomPad = 0
	}
	return lipgloss.NewStyle().
		Background(base).
		Width(a.width).
		Padding(topPad, 0, bottomPad, 0).
		Align(lipgloss.Center).
		Render(content)
}

type App struct {
	cfg config.Config

	screen    int
	width     int
	height    int
	frame     int
	bazarrVer string

	// Menu
	menuIdx int

	// Browser (movies or shows)
	mediaType   string // "movie" or "show"
	movies      []bazarr.Movie
	shows       []bazarr.Show
	search      string
	browserIdx  int
	focusSearch bool
	loading     bool

	// Show episode navigation
	selectedShow bazarr.Show
	episodes     []bazarr.Episode
	episodeIdx   int
	epLoading    bool

	// Subtitle selection (for episodes)
	items  []SelectItem
	selIdx int

	// Staging area (accumulates selected subtitles across all screens)
	staged []SyncJob

	// Sync
	jobs    []SyncJob
	results []string
	summary string

	// Config screen
	cfgUrl               string
	cfgToken             string
	cfgIdx               int
	cfgValidating        bool
	cfgValidationResult  string
	cfgValidationSuccess bool
}

func Run() {
	c, err := config.LoadConfig()
	app := App{
		items:   make([]SelectItem, 0),
		results: make([]string, 0),
		staged:  make([]SyncJob, 0),
	}

	if err != nil || c.BaseUrl == "" || c.ApiToken == "" {
		config.SetConfig(c)
		app.cfg = c
		app.screen = ScreenConfig
		app.populateConfigFields()
	} else {
		config.SetConfig(c)
		app.cfg = c
		version, err := bazarr.CheckHealth(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		app.bazarrVer = version
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
