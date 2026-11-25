package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	colBG       = "#14110E"
	colCream    = "#F4E8D4"
	colDim      = "#9A8B76"
	colFaint    = "#6E6256"
	colBrass    = "#E2B15A"
	colBrassHot = "#FFE7A8"
	colInk      = "#1C140C"
	colIvory    = "#FFF8EE"

	colLight = "#F0D9B5"
	colDark  = "#B58863"

	colLightRay = "#E7B4A4"
	colDarkRay  = "#A86A58"
	colRayDotL  = "#6E241C"
	colRayDotD  = "#FFE4DC"

	colTry       = "#F6D86A"
	colConflict  = "#E23B32"
	colAttackerL = "#E08B7C"
	colAttackerD = "#8C3A34"
	colRetreatL  = "#C5D4E0"
	colRetreatD  = "#6E8BA3"
	colPlacedL   = "#FFE3A8"
	colPlacedD   = "#C9955A"

	colFlashLight = "#FFE7A3"
	colFlashDark  = "#C48A3A"
	colFlashInk   = "#1A1208"

	colGood  = "#8FCB9B"
	colPause = "#E07A5F"
	colLine  = "#6E5A40"
	colEdge  = "#3A3228"
)

func paint(text, fg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(colBG)).
		Render(text)
}

func paintBold(text, fg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(colBG)).
		Bold(true).
		Render(text)
}

func badge(text, fg, bg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

func fillRight(line string, width int) string {
	gap := width - lipgloss.Width(line)
	if gap <= 0 {
		return line
	}
	return line + lipgloss.NewStyle().Background(lipgloss.Color(colBG)).Render(repeat(' ', gap))
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(colBG)).Render(repeat(' ', n))
}

func blankLine(width int) string {
	return lipgloss.NewStyle().Width(width).Background(lipgloss.Color(colBG)).Render("")
}

func vspace(width, height int) string {
	if height < 1 {
		height = 1
	}
	line := blankLine(width)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func repeat(r rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}

func truncatePlain(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
