package tui

import "github.com/charmbracelet/lipgloss"

var (
	cyan    = lipgloss.Color("#00D8FF")
	purple  = lipgloss.Color("#A855F7")
	green   = lipgloss.Color("#00FF9D")
	red     = lipgloss.Color("#FF5555")
	yellow  = lipgloss.Color("#FBBF24")
	white   = lipgloss.Color("#E0E6ED")
	dim     = lipgloss.Color("#8892A4")
	bg      = lipgloss.Color("#0A0E17")
	panelBg = lipgloss.Color("#0D1220")
	border  = lipgloss.Color("#1E2D3D")
	darkCyan = lipgloss.Color("#009ACD")
)

var (
	titleStyle   = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Padding(0, 1)
	subtitleStyle = lipgloss.NewStyle().Foreground(dim).Italic(true).Align(lipgloss.Center)
	menuSel      = lipgloss.NewStyle().Foreground(darkCyan).Background(cyan).Bold(true).Padding(0, 2)
	menuUnsel    = lipgloss.NewStyle().Foreground(dim).Padding(0, 2)
	panelStyle   = lipgloss.NewStyle().Background(panelBg).Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(1, 2)
	itemSel      = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	itemUnsel    = lipgloss.NewStyle().Foreground(dim)
	searchStyle  = lipgloss.NewStyle().Foreground(yellow).Padding(0, 1)
	barStyle     = lipgloss.NewStyle().Foreground(cyan).Padding(0, 1).Align(lipgloss.Center)
	checkOn      = lipgloss.NewStyle().Foreground(green).Bold(true).Width(3).Align(lipgloss.Center)
	checkOff     = lipgloss.NewStyle().Foreground(dim).Width(3).Align(lipgloss.Center)
	versionStyle = lipgloss.NewStyle().Foreground(dim)
	syncSuccess  = lipgloss.NewStyle().Foreground(green).Bold(true)
	syncError    = lipgloss.NewStyle().Foreground(red).Bold(true)
	syncRunning  = lipgloss.NewStyle().Foreground(cyan)
	titleBar     = lipgloss.NewStyle().Background(panelBg).Foreground(cyan).Bold(true).Padding(0, 1)
	arrowStyle   = lipgloss.NewStyle().Foreground(cyan).Bold(true).Width(2).Align(lipgloss.Center)
	footerStyle  = lipgloss.NewStyle().Foreground(dim).Padding(0, 1).Align(lipgloss.Center)
)

func spinnerStr(n int) string {
	frames := []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇", "⠏", "⠟"}
	return syncRunning.Render(frames[n%len(frames)])
}
