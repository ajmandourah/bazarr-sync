package tui

import (
	"fmt"
	"strings"
)

func (a App) stagedPanelHeight() int {
	if len(a.staged) == 0 {
		return 0
	}
	maxItems := 3
	if a.height < 24 {
		maxItems = 2
	}
	h := 1 // header line
	if len(a.staged) > maxItems {
		h += maxItems + 1 // items + "more" line
	} else {
		h += len(a.staged)
	}
	return h
}

func (a App) renderStagedList() string {
	if len(a.staged) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(stagedHeader.Render(fmt.Sprintf("  ── Staged: %d ──", len(a.staged))))

	maxItems := 3
	if a.height < 24 {
		maxItems = 2
	}

	show := a.staged
	if len(show) > maxItems {
		show = a.staged[:maxItems]
	}

	for _, job := range show {
		label := job.Title + " [" + job.Language + "]"
		maxLabelLen := a.width - 12
		if maxLabelLen < 10 {
			maxLabelLen = 10
		}
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen] + ".."
		}
		b.WriteString("\n" + stagedListItem.Render("  ▶ "+label))
	}

	if len(a.staged) > maxItems {
		b.WriteString("\n" + stagedMore.Render(fmt.Sprintf("  + %d more...", len(a.staged)-maxItems)))
	}

	return b.String()
}
