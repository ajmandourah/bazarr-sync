package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette
var (
	base      = lipgloss.Color("#1E1E2E")
	mantle    = lipgloss.Color("#181825")
	crust     = lipgloss.Color("#11111B")
	surface0  = lipgloss.Color("#313244")
	surface1  = lipgloss.Color("#45475A")
	surface2  = lipgloss.Color("#585B70")
	overlay0  = lipgloss.Color("#6C7086")
	overlay1  = lipgloss.Color("#7F849C")
	overlay2  = lipgloss.Color("#9399B2")
	subtext0  = lipgloss.Color("#A6ADC8")
	subtext1  = lipgloss.Color("#BAC2DE")
	text      = lipgloss.Color("#CDD6F4")
	rosewater = lipgloss.Color("#F5E0DC")
	flamingo  = lipgloss.Color("#F2CDDE")
	pink      = lipgloss.Color("#F5C2E7")
	mauve     = lipgloss.Color("#CBA6F7")
	red       = lipgloss.Color("#F38BA8")
	maroon    = lipgloss.Color("#EA76CB")
	peach     = lipgloss.Color("#FAB387")
	yellow    = lipgloss.Color("#F9E2AF")
	green     = lipgloss.Color("#A6E3A1")
	teal      = lipgloss.Color("#94E2D5")
	sky       = lipgloss.Color("#89DCEB")
	sapphire  = lipgloss.Color("#74C7EC")
	blue      = lipgloss.Color("#89B4FA")
	lavender  = lipgloss.Color("#B4BEFE")
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(text).Bold(true).Background(base)
	headerStyle   = lipgloss.NewStyle().Foreground(mauve).Bold(true).Background(base)
	subtitleStyle = lipgloss.NewStyle().Foreground(overlay1).Italic(true).Align(lipgloss.Center).Background(base)
	menuSel       = lipgloss.NewStyle().Foreground(mauve).Bold(true).Padding(0, 2).Background(base)
	menuUnsel     = lipgloss.NewStyle().Foreground(subtext0).Padding(0, 2).Background(base)
	panelStyle    = lipgloss.NewStyle().Background(base).Padding(1, 2)
	itemSel         = lipgloss.NewStyle().Foreground(pink).Bold(true).Background(base)
	itemUnsel       = lipgloss.NewStyle().Foreground(subtext0).Background(base)
	searchStyle     = lipgloss.NewStyle().Foreground(yellow).Background(base)
	searchBarActive = lipgloss.NewStyle().
			Foreground(yellow).
			Padding(0, 1).
			Background(base)
	searchBarInactive = lipgloss.NewStyle().
				Foreground(overlay0).Background(base)
	barStyle            = lipgloss.NewStyle().Foreground(lavender).Align(lipgloss.Center).Background(base)
	checkOn             = lipgloss.NewStyle().Foreground(green).Bold(true).Background(base)
	checkOff            = lipgloss.NewStyle().Foreground(overlay0).Background(base)
	versionStyle        = lipgloss.NewStyle().Foreground(overlay1).Background(base)
	syncSuccess         = lipgloss.NewStyle().Foreground(green).Bold(true).Background(base)
	syncError           = lipgloss.NewStyle().Foreground(red).Bold(true).Background(base)
	syncRunning         = lipgloss.NewStyle().Foreground(teal).Background(base)
	titleBar            = lipgloss.NewStyle().Foreground(text).Bold(true).Background(base)
	arrowStyle          = lipgloss.NewStyle().Foreground(pink).Bold(true).Width(2).Align(lipgloss.Center).Background(base)
	footerStyle         = lipgloss.NewStyle().Foreground(overlay0).Align(lipgloss.Center).Background(base)
	cheatSheet          = lipgloss.NewStyle().Foreground(mauve).Align(lipgloss.Center).Background(base)
	subtitleHeaderStyle = lipgloss.NewStyle().Foreground(pink).Bold(true).Background(base)
	stagedPanel         = lipgloss.NewStyle().Background(base).Padding(1, 2)
	stagedTitle = lipgloss.NewStyle().Foreground(peach).Bold(true).Background(base)
	stagedItem  = lipgloss.NewStyle().Foreground(subtext1).Background(base)
	stagedLang  = lipgloss.NewStyle().Foreground(green).Bold(true).Background(base)
)

func spinnerStr(n int) string {
	frames := []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇", "⠏", "⠟"}
	return syncRunning.Render(frames[n%len(frames)])
}

// bgLine wraps text in a full-width background line to prevent terminal color leaking.
func bgLine(text string, width int) string {
	return lipgloss.NewStyle().Background(base).Width(width).Render(text)
}
