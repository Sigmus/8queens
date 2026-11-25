package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Sigmus/8queens/solver"
)

const (
	rankW       = 2
	panelOuterH = 17
	minTermW    = 28
	minTermH    = 15
)

type layout struct {
	cellW    int
	cellH    int
	padX     int
	panelW   int
	subtitle bool
	gap      int
}

func (l layout) boardSize() (int, int) {
	innerW := rankW + boardN*l.cellW
	w := innerW + l.padX*2 + 2
	h := boardN*l.cellH + 1 + 2 + 1
	return w, h
}

func (l layout) need() (int, int) {
	bw, bh := l.boardSize()
	w := bw
	if l.panelW > 0 {
		w += 2 + l.panelW
	}
	h := bh + 3 + l.gap
	if l.subtitle {
		h++
	}
	if l.panelW > 0 && bh < panelOuterH {
		h += panelOuterH - bh
	}
	return w, h
}

func pickLayout(width, height int) (layout, bool) {
	options := []layout{
		{7, 3, 1, 34, true, 1},
		{5, 3, 1, 34, true, 1},
		{7, 2, 1, 34, true, 1},
		{5, 2, 1, 32, false, 1},
		{5, 2, 1, 32, false, 0},
		{7, 1, 1, 32, false, 0},
		{5, 1, 1, 32, false, 0},
		{5, 2, 1, 0, false, 0},
		{3, 1, 1, 0, false, 0},
		{3, 1, 0, 0, false, 0},
	}
	for _, opt := range options {
		w, h := opt.need()
		if width >= w && height >= h {
			return opt, true
		}
	}
	return layout{}, false
}

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Eight Queens\n"
	}
	l, ok := pickLayout(m.width, m.height)
	if !ok {
		return place(m, m.tooSmall())
	}
	board := m.renderBoard(l)
	var main string
	if l.panelW > 0 {
		panel := m.renderPanel(l.panelW)
		// Keep the panel card aligned with the board frame, under the caption.
		panel = lipgloss.JoinVertical(lipgloss.Left, blankLine(lipgloss.Width(panel)), panel)
		gapH := lipgloss.Height(board)
		if h := lipgloss.Height(panel); h > gapH {
			gapH = h
		}
		main = lipgloss.JoinHorizontal(lipgloss.Top, board, vspace(2, gapH), panel)
	} else {
		main = board
	}
	inner := lipgloss.Width(main)
	parts := []string{m.renderTitle(inner, l.subtitle), m.renderStats(inner)}
	if l.gap > 0 {
		parts = append(parts, blankLine(inner))
	}
	parts = append(parts, main, m.renderFooter(inner))
	body := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return place(m, body)
}

func place(m Model, body string) string {
	hPos := lipgloss.Center
	vPos := lipgloss.Center
	if lipgloss.Width(body) > m.width {
		hPos = lipgloss.Left
	}
	if lipgloss.Height(body) > m.height {
		vPos = lipgloss.Top
	}
	return lipgloss.Place(
		m.width, m.height,
		hPos, vPos,
		body,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(colBG)),
	)
}

func (m Model) tooSmall() string {
	title := paintBold("Eight Queens", colBrass)
	msg := paint("Need a terminal at least 28×15.", colCream)
	now := paint(fmt.Sprintf("This one is %d×%d.", m.width, m.height), colDim)
	return lipgloss.JoinVertical(lipgloss.Left, title, msg, now)
}

func (m Model) renderTitle(width int, subtitle bool) string {
	title := paintBold("♛  Eight Queens", colBrass)
	if !subtitle {
		return fillRight(title, width)
	}
	sub := paint("No two queens share a rank, file, or diagonal.", colDim)
	block := lipgloss.JoinVertical(lipgloss.Left, fillRight(title, width), fillRight(sub, width))
	return block
}

func (m Model) renderStats(width int) string {
	modeBadge := badge(m.modeName(), colInk, colBrass)
	var status string
	if m.paused {
		status = badge("paused", colInk, colPause)
	} else {
		status = badge("running", colInk, colGood)
	}
	sol := paint(fmt.Sprintf("%d/%d", m.solutions, m.total), colCream)
	attempts := paint(fmt.Sprintf("attempts %s", withCommas(m.attempts)), colDim)
	speed := paint(fmt.Sprintf("speed %s %d", speedMeter(m.speed), m.speed), colDim)
	pass := paint(fmt.Sprintf("pass %d", m.cycles+1), colDim)

	candidates := [][]string{
		{modeBadge, status, sol, attempts, speed, pass},
		{modeBadge, status, sol, attempts, speed},
		{modeBadge, status, sol, attempts},
		{modeBadge, status, sol},
		{sol, status},
		{sol},
	}
	for _, parts := range candidates {
		line, ok := joinFit(width, parts...)
		if ok {
			return fillRight(line, width)
		}
	}
	return fillRight(sol, width)
}

func joinFit(width int, parts ...string) (string, bool) {
	total := 0
	for i, p := range parts {
		total += lipgloss.Width(p)
		if i > 0 {
			total += 2
		}
	}
	if total > width {
		return "", false
	}
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteString(spaces(2))
		}
		b.WriteString(p)
	}
	return b.String(), true
}

func (m Model) renderFooter(width int) string {
	verb := "pause"
	if m.paused {
		verb = "resume"
	}
	candidates := []string{
		fmt.Sprintf("space %s    +/- speed    ← → speed    m mode    r restart    q quit", verb),
		fmt.Sprintf("space %s   +/- speed   m mode   r restart   q quit", verb),
		fmt.Sprintf("space %s   +/- speed   m mode   q quit", verb),
		fmt.Sprintf("space %s   m mode   q quit", verb),
		fmt.Sprintf("space %s   q quit", verb),
	}
	text := candidates[len(candidates)-1]
	for _, c := range candidates {
		if lipgloss.Width(c) <= width {
			text = c
			break
		}
	}
	return fillRight(paint(text, colFaint), width)
}

func (m Model) renderBoard(l layout) string {
	var rows []string
	for r := 0; r < boardN; r++ {
		cells := make([]string, 0, boardN+1)
		cells = append(cells, rankLabel(fmt.Sprintf("%d", boardN-r), l.cellH))
		for c := 0; c < boardN; c++ {
			cells = append(cells, m.renderCell(r, c, l.cellW, l.cellH))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	rows = append(rows, fileLabels(l.cellW))
	inner := strings.Join(rows, "\n")

	borderFG := colLine
	switch {
	case m.celebrating && m.blinkOn:
		borderFG = colBrassHot
	case m.celebrating || m.kind == solver.KindSolution:
		borderFG = colBrass
	}
	frame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderFG)).
		BorderBackground(lipgloss.Color(colBG)).
		Background(lipgloss.Color(colBG)).
		Padding(0, l.padX).
		Render(inner)

	caption := fillRight(m.captionLine(), lipgloss.Width(frame))
	return caption + "\n" + frame
}

func rankLabel(text string, cellH int) string {
	lines := make([]string, cellH)
	mid := (cellH - 1) / 2
	for i := range lines {
		glyph := ""
		if i == mid {
			glyph = text
		}
		lines[i] = lipgloss.NewStyle().
			Width(rankW).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color(colDim)).
			Background(lipgloss.Color(colBG)).
			Render(glyph)
	}
	return strings.Join(lines, "\n")
}

func fileLabels(cellW int) string {
	parts := make([]string, 0, boardN+1)
	parts = append(parts, lipgloss.NewStyle().
		Width(rankW).
		Background(lipgloss.Color(colBG)).
		Render(""))
	for c := 0; c < boardN; c++ {
		parts = append(parts, lipgloss.NewStyle().
			Width(cellW).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(colDim)).
			Background(lipgloss.Color(colBG)).
			Render(string(rune('a'+c))))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) renderCell(r, c, w, h int) string {
	light := (r+c)%2 == 0
	bg := colDark
	fg := colIvory
	glyph := " "
	if light {
		bg = colLight
		fg = colInk
	}
	queen := m.board != nil && r < len(m.board) && m.board[r] == c

	flash := m.celebrating && m.blinkOn
	switch {
	case flash:
		glyph = " "
		fg = colFlashInk
		if light {
			bg = colFlashLight
		} else {
			bg = colFlashDark
		}
		if queen {
			glyph = "♕"
		}
	case m.kind == solver.KindConflict && r == m.hiRow && c == m.hiCol:
		bg, fg, glyph = colConflict, colIvory, "×"
	case m.kind == solver.KindConflict && hasPos(m.attacks, r, c):
		if light {
			bg, fg = colAttackerL, colInk
		} else {
			bg, fg = colAttackerD, colIvory
		}
		glyph = "♛"
	case m.kind == solver.KindConflict && hasPos(m.rays, r, c) && !queen:
		if light {
			bg, fg = colLightRay, colRayDotL
		} else {
			bg, fg = colDarkRay, colRayDotD
		}
		glyph = "●"
	case m.kind == solver.KindPlace && r == m.hiRow && c == m.hiCol:
		if light {
			bg, fg = colPlacedL, colInk
		} else {
			bg, fg = colPlacedD, colIvory
		}
		glyph = "♛"
	case m.kind == solver.KindBacktrack && r == m.hiRow && c == m.hiCol:
		if light {
			bg, fg = colRetreatL, colInk
		} else {
			bg, fg = colRetreatD, colIvory
		}
		glyph = "·"
	case queen:
		glyph = "♛"
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Bold(true).
		Render(glyph)
}

func (m Model) captionLine() string {
	fg := colCream
	text := m.caption()
	if m.celebrating && m.blinkOn {
		fg = colBrassHot
	} else if m.celebrating || m.kind == solver.KindSolution {
		fg = colBrass
	}
	return paintBold(text, fg)
}

func (m Model) caption() string {
	if m.celebrating || m.kind == solver.KindSolution {
		mark := "★"
		if m.celebrating && m.blinkOn {
			mark = "✦"
		}
		return fmt.Sprintf("%s  Solution %d of %d  %s", mark, m.solutions, m.total, mark)
	}
	switch m.kind {
	case solver.KindPlace:
		return "Placed a queen on " + m.square(m.hiRow, m.hiCol)
	case solver.KindConflict:
		return "Blocked on " + m.square(m.hiRow, m.hiCol)
	case solver.KindBacktrack:
		return "Backtracked from " + m.square(m.hiRow, m.hiCol)
	default:
		return "Setting up the board"
	}
}

func (m Model) detail() string {
	switch m.kind {
	case solver.KindConflict:
		return m.attackPhrase()
	case solver.KindPlace:
		return fmt.Sprintf("%d of 8 queens placed", m.queensOn())
	case solver.KindBacktrack:
		return "Queen lifted, next file"
	case solver.KindSolution:
		return "No queen attacks another"
	default:
		return ""
	}
}

func (m Model) attackPhrase() string {
	if len(m.attacks) == 0 {
		return "Attacked"
	}
	parts := make([]string, len(m.attacks))
	for i, p := range m.attacks {
		how := "diagonal"
		if p.Col == m.hiCol {
			how = "file"
		}
		parts[i] = fmt.Sprintf("%s (%s)", m.square(p.Row, p.Col), how)
	}
	return "by " + strings.Join(parts, ", ")
}

func (m Model) renderPanel(outerW int) string {
	textW := outerW - 4
	if textW < 8 {
		textW = 8
	}
	status := "running"
	if m.paused {
		status = "paused"
	}
	lines := []string{
		kv("Mode", m.modeName(), textW),
		kv("Status", status, textW),
		kv("Speed", speedMeter(m.speed)+"  "+fmt.Sprint(m.speed), textW),
		"",
		kv("Solutions", fmt.Sprintf("%d/%d", m.solutions, m.total), textW),
		progressBar(m.solutions, m.total, textW),
		kv("Attempts", withCommas(m.attempts), textW),
		kv("Queens", fmt.Sprintf("%d/8", m.queensOn()), textW),
		kv("Pass", fmt.Sprint(m.cycles+1), textW),
		"",
		"This step",
		truncatePlain(m.caption(), textW),
		truncatePlain(m.detail(), textW),
		"",
		truncatePlain("♛ queen   × blocked   ● line", textW),
	}
	painted := make([]string, len(lines))
	for i, line := range lines {
		fg := colCream
		switch {
		case line == "":
			fg = colBG
		case i == 10 || i == 14:
			fg = colDim
		case strings.HasPrefix(line, "Status") && m.paused:
			fg = colPause
		case strings.HasPrefix(line, "Status"):
			fg = colGood
		case i == 11 && (m.celebrating || m.kind == solver.KindSolution):
			fg = colBrass
		}
		painted[i] = fillRight(paint(line, fg), textW)
	}
	body := strings.Join(painted, "\n")
	return lipgloss.NewStyle().
		Width(textW+2).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colEdge)).
		BorderBackground(lipgloss.Color(colBG)).
		Background(lipgloss.Color(colBG)).
		Render(body)
}

func kv(label, value string, width int) string {
	gap := width - lipgloss.Width(label) - lipgloss.Width(value)
	if gap < 1 {
		return truncatePlain(label+" "+value, width)
	}
	return label + repeat(' ', gap) + value
}

func progressBar(done, total, width int) string {
	if width <= 0 {
		return ""
	}
	filled := 0
	if total > 0 {
		filled = done * width / total
		if done > 0 && filled == 0 {
			filled = 1
		}
		if filled > width {
			filled = width
		}
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func speedMeter(speed int) string {
	return strings.Repeat("█", speed) + strings.Repeat("░", maxSpeed-speed)
}

func withCommas(n int) string {
	s := fmt.Sprint(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func hasPos(ps []solver.Pos, r, c int) bool {
	for _, p := range ps {
		if p.Row == r && p.Col == c {
			return true
		}
	}
	return false
}
