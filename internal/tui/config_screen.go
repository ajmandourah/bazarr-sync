package tui

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigResult struct {
	Success bool
	Error   string
}

func (a App) ConfigView() string {
	fields := []struct {
		Label string
		Value string
	}{
		{Label: "Bazarr URL", Value: a.cfgUrl},
		{Label: "Api Token", Value: maskString(a.cfgToken)},
	}

	var b strings.Builder
	b.WriteString(titleBar.Render("  Settings"))
	b.WriteString("\n")

	for i, field := range fields {
		inputStyle := cfgInputStyle
		labelStyle := cfgLabelStyle
		if i == a.cfgIdx {
			inputStyle = cfgInputFocusedStyle
			labelStyle = cfgLabelFocusedStyle
		}

		label := fmt.Sprintf("  %s ", field.Label)
		label = strings.ReplaceAll(label, " ", "\u00A0")

		var displayValue string
		if field.Value == "" {
			if field.Label == "Bazarr URL" {
				displayValue = cfgPlaceholderStyle.Render("https://bazarr.example.com")
			} else {
				displayValue = cfgPlaceholderStyle.Render("...")
			}
		} else {
			displayValue = field.Value
		}

		b.WriteString("\n" + lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(label),
			inputStyle.Render(displayValue),
		))
	}

	if a.cfgValidating {
		b.WriteString("\n\n  " + cfgValidatingStyle.Render(spinnerStr(a.frame)+" Validating connection..."))
	} else if a.cfgValidationResult != "" {
		if a.cfgValidationSuccess {
			b.WriteString("\n\n  " + cfgSuccessStyle.Render("✓ Configuration saved and validated!"))
		} else {
			b.WriteString("\n\n  " + cfgErrorStyle.Render("✗ "+a.cfgValidationResult))
		}
	}

	b.WriteString("\n\n" + cheatSheet.Render("  Tab/↑↓ navigate  •  Enter save  •  q/Esc back"))

	return panelStyle.Render(b.String())
}

func maskString(s string) string {
	if s == "" {
		return ""
	}
	return strings.Repeat("•", len(s))
}

func (a App) ConfigHandler(msg tea.KeyMsg) (App, tea.Cmd) {
	if a.cfgValidating {
		return a, nil
	}

	switch msg.Type {
	case tea.KeyEscape:
		a.screen = ScreenMenu
		a.cfgValidationResult = ""
		a.cfgValidationSuccess = false
		a.cfgValidating = false
		return a, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			a.screen = ScreenMenu
			a.cfgValidationResult = ""
			a.cfgValidationSuccess = false
			a.cfgValidating = false
			return a, nil
		}
	case tea.KeyEnter:
		if a.cfgValidationResult != "" {
			if a.cfgValidationSuccess {
				a.screen = ScreenMenu
				a.cfgValidationResult = ""
				a.cfgValidationSuccess = false
				return a, nil
			}
			a.cfgValidationResult = ""
			a.cfgValidationSuccess = false
			return a, nil
		}
		a.cfgValidating = true
		return a, a.saveConfigCmd()
	case tea.KeyTab:
		a.cfgIdx = (a.cfgIdx + 1) % 2
		return a, nil
	case tea.KeyUp:
		if a.cfgIdx > 0 {
			a.cfgIdx--
		}
		return a, nil
	case tea.KeyDown:
		if a.cfgIdx < 1 {
			a.cfgIdx++
		}
		return a, nil
	}

	if msg.Type == tea.KeyBackspace {
		return a.handleConfigBackspace()
	}

	if msg.Type == tea.KeyRunes {
		return a.handleConfigInput(msg)
	}

	return a, nil
}

func (a App) handleConfigBackspace() (App, tea.Cmd) {
	switch a.cfgIdx {
	case 0:
		if len(a.cfgUrl) > 0 {
			a.cfgUrl = a.cfgUrl[:len(a.cfgUrl)-1]
		}
	case 1:
		if len(a.cfgToken) > 0 {
			a.cfgToken = a.cfgToken[:len(a.cfgToken)-1]
		}
	}
	return a, nil
}

func (a App) handleConfigInput(msg tea.KeyMsg) (App, tea.Cmd) {
	if len(msg.Runes) != 1 {
		return a, nil
	}
	r := msg.Runes[0]
	switch a.cfgIdx {
	case 0:
		a.cfgUrl += string(r)
	case 1:
		a.cfgToken += string(r)
	}
	return a, nil
}

func (a App) saveConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := config.GetConfig()
		cfg.BaseUrl = a.cfgUrl
		cfg.ApiToken = a.cfgToken
		cfg.ApiUrl = ""

		parsedUrl, err := url.Parse(a.cfgUrl)
		if err != nil {
			return ConfigResult{
				Success: false,
				Error:   "Invalid URL: " + err.Error(),
			}
		}
		cfg.ApiUrl = parsedUrl.Scheme + "://" + parsedUrl.Host + "/api/"
		config.SetConfig(cfg)

		err = a.writeConfigFile()
		if err != nil {
			return ConfigResult{
				Success: false,
				Error:   err.Error(),
			}
		}

		_, err = bazarr.CheckHealth(cfg)
		if err != nil {
			return ConfigResult{
				Success: false,
				Error:   err.Error(),
			}
		}

		return ConfigResult{Success: true}
	}
}

func (a *App) populateConfigFields() {
	c := config.GetConfig()
	a.cfgUrl = c.BaseUrl
	a.cfgToken = c.ApiToken
}

func (a *App) writeConfigFile() error {
	c := config.GetConfig()
	if c.BaseUrl == "" || c.ApiToken == "" {
		return fmt.Errorf("bazarr url and api token are required")
	}

	cfgFile := config.GetConfigFile()
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}

	return os.WriteFile(cfgFile, []byte(fmt.Sprintf("bazarr_url: %s\nbazarr_token: %s\n",
		c.BaseUrl, c.ApiToken)), 0644)
}
