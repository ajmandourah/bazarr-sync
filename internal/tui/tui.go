package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	ScreenMenu      int = iota
	ScreenBrowser
	ScreenSelecting
	ScreenSyncing
	ScreenDone
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
	Movies   []bazarr.Movie
	Shows    []bazarr.Show
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
	switch msg.(type) {
	case tea.WindowSizeMsg:
		// Size changed
	case tea.KeyMsg:
		return a.HandleKey(msg.(tea.KeyMsg))
	case DataMessage:
		return a.HandleData(msg.(DataMessage))
	case SyncResultMessage:
		return a.HandleSyncResult(msg.(SyncResultMessage))
	case TickMessage:
		a.frame++
		return a, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return TickMessage{} })
	}
	return a, nil
}

func (a App) View() string {
	switch a.screen {
	case ScreenMenu:
		return a.MenuView()
	case ScreenBrowser:
		return a.MenuView()
	case ScreenSelecting:
		return a.MenuView()
	case ScreenSyncing:
		return a.MenuView()
	case ScreenDone:
		return a.MenuView()
	}
	return ""
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

	// Browser
	mediaType  string
	movies     []bazarr.Movie
	shows      []bazarr.Show
	episodes   []bazarr.Episode
	search     string
	browserIdx int
	focusSearch bool

	// Selection
	items  []SelectItem
	selIdx int

	// Sync
	jobs    []SyncJob
	results []string
	summary string
}

func Run() {
	config.InitConfig()
	c := config.GetConfig()

	version, err := bazarr.CheckHealth(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := App{
		cfg:       c,
		bazarrVer: version,
		items:     make([]SelectItem, 0),
		results:   make([]string, 0, 1),
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
