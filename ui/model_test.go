package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"eightqueens/solver"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	m.Run()
}

func TestNewRejectsBadFlags(t *testing.T) {
	if _, err := New("nope", 5); err == nil {
		t.Fatal("expected an error for an unknown mode")
	}
	if _, err := New("step", 0); err == nil {
		t.Fatal("expected an error for speed 0")
	}
	if _, err := New("step", 9); err == nil {
		t.Fatal("expected an error for speed 9")
	}
	if _, err := New("solutions", 1); err != nil {
		t.Fatal(err)
	}
}

func TestSpaceIsTheOnlyPause(t *testing.T) {
	m := mustModel(t, "step", 8)
	before := m.attempts
	m = press(t, m, " ")
	if !m.paused {
		t.Fatal("space did not pause")
	}
	for i := 0; i < 6; i++ {
		m = frame(t, m)
	}
	if m.attempts != before || m.kind != solver.KindPlace {
		t.Fatalf("search advanced while paused: attempts %d → %d, kind %s", before, m.attempts, m.kind)
	}
	for _, k := range []string{"m", "+", "-", "r", "up", "left", "q"} {
		if k == "q" {
			continue
		}
		other := mustModel(t, "step", 5)
		other = press(t, other, k)
		if other.paused {
			t.Fatalf("key %q paused the search", k)
		}
	}
	m = press(t, m, " ")
	if m.paused {
		t.Fatal("second space did not resume")
	}
	resumed := m.attempts
	m = frame(t, m)
	if m.attempts <= resumed && m.kind == solver.KindPlace && m.hiRow == 0 && m.hiCol == 0 {
		t.Fatal("resume did not take the next step")
	}
}

func TestFramesKeepComingWithoutInput(t *testing.T) {
	m := mustModel(t, "step", 8)
	start := m.attempts
	for i := 0; i < 40; i++ {
		next, cmd := m.Update(frameMsg{})
		if cmd == nil {
			t.Fatal("frame tick was not rescheduled")
		}
		m = next.(Model)
	}
	if m.attempts <= start {
		t.Fatal("search did not advance on its own")
	}
	if m.paused {
		t.Fatal("search paused itself")
	}
}

func TestQuitKeys(t *testing.T) {
	for _, k := range []string{"q", "Q", "ctrl+c"} {
		m := mustModel(t, "step", 5)
		_, cmd := m.Update(keyMsg(k))
		if !isQuit(cmd) {
			t.Fatalf("key %q did not quit", k)
		}
	}
}

func TestStepModeShowsConflictsAndPlacements(t *testing.T) {
	m := sized(mustModel(t, "step", 8), 100, 36)
	if m.kind != solver.KindPlace || m.square(m.hiRow, m.hiCol) != "a8" {
		t.Fatalf("opening step = %s %s, want place a8", m.kind, m.square(m.hiRow, m.hiCol))
	}
	view := strip(m.View())
	for _, want := range []string{"Eight Queens", "step", "running", "a8", "Placed a queen", "0/92", "space pause", "q quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q", want)
		}
	}
	var conflict Model
	found := false
	for i := 0; i < 40; i++ {
		m = frame(t, m)
		if m.kind == solver.KindConflict && len(m.attacks) > 0 && len(m.rays) > 0 {
			conflict = m
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no conflict with an attack line in the first 40 steps")
	}
	cv := conflict.View()
	if !strings.Contains(cv, truecolorBG(colConflict)) {
		t.Fatal("tried square is not drawn in the conflict color")
	}
	if !strings.Contains(cv, truecolorBG(colAttackerL)) && !strings.Contains(cv, truecolorBG(colAttackerD)) {
		t.Fatal("attacking queen is not highlighted")
	}
	ray := conflict.rays[0]
	marked := conflict.renderCell(ray.Row, ray.Col, 5, 1)
	if !strings.Contains(strip(marked), "●") {
		t.Fatalf("attack line square = %q", strip(marked))
	}
	quiet := conflict
	quiet.kind = solver.KindPlace
	quiet.hiRow, quiet.hiCol = -1, -1
	if marked == quiet.renderCell(ray.Row, ray.Col, 5, 1) {
		t.Fatal("attack line square looks like an ordinary square")
	}
	if !strings.Contains(strip(cv), "Blocked on") {
		t.Fatal("conflict caption missing")
	}
}

func TestSolutionBlinkHoldsTheBoard(t *testing.T) {
	m := sized(mustModel(t, "step", 8), 100, 40)
	for i := 0; i < 200000 && !m.celebrating; i++ {
		m = frame(t, m)
	}
	if !m.celebrating || !m.blinkOn {
		t.Fatal("step mode never flashed a solution")
	}
	if m.solutions < 1 || m.queensOn() != 8 {
		t.Fatalf("celebration board is not a full solution: sols=%d queens=%d", m.solutions, m.queensOn())
	}
	attempts := m.attempts
	on := m.View()
	if !strings.Contains(on, truecolorBG(colFlashLight)) {
		t.Fatal("blink-on frame is missing the flash color")
	}
	if !strings.Contains(strip(on), "Solution 1 of 92") {
		t.Fatalf("banner missing on blink:\n%s", strip(on))
	}
	m = frame(t, m)
	if !m.celebrating || m.blinkOn {
		t.Fatalf("expected the blink to toggle off, celebrating=%v blink=%v", m.celebrating, m.blinkOn)
	}
	off := m.View()
	if on == off {
		t.Fatal("blink did not change the picture")
	}
	if strings.Contains(off, truecolorBG(colFlashLight)) {
		t.Fatal("blink-off frame still uses the flash color")
	}
	if strings.Contains(strip(off), "✦") || !strings.Contains(strip(off), "★") {
		t.Fatal("banner marks did not swap with the blink")
	}
	for guard := 0; m.celebrating && guard < 20; guard++ {
		m = frame(t, m)
		if m.attempts != attempts {
			t.Fatal("search advanced during the solution blink")
		}
	}
	if m.celebrating {
		t.Fatal("blink never ended")
	}
	m = frame(t, m)
	if m.kind != solver.KindBacktrack {
		t.Fatalf("after the blink, kind = %s, want backtrack", m.kind)
	}
}

func TestSolutionsModeShowsOnlyCompleteBoards(t *testing.T) {
	m := sized(mustModel(t, "solutions", 8), 100, 36)
	if m.solutions != 1 || m.queensOn() != 8 || m.celebrating {
		t.Fatalf("opening solutions frame: sols=%d queens=%d celebrating=%v", m.solutions, m.queensOn(), m.celebrating)
	}
	if !solver.Valid(m.board) {
		t.Fatalf("opening board is illegal: %v", m.board)
	}
	for i := 0; i < 10; i++ {
		m = frame(t, m)
		if m.celebrating {
			t.Fatal("solutions mode flashed the board")
		}
		if m.kind != solver.KindSolution || m.queensOn() != 8 || !solver.Valid(m.board) {
			t.Fatalf("frame %d is not a solution: kind=%s board=%v", i, m.kind, m.board)
		}
	}
	if m.solutions != 11 {
		t.Fatalf("solution counter = %d, want 11", m.solutions)
	}
	view := strip(m.View())
	if !strings.Contains(view, "solutions") || !strings.Contains(view, "11/92") {
		t.Fatalf("solutions view missing status:\n%s", view)
	}
}

func TestSolutionsModeLoops(t *testing.T) {
	m := mustModel(t, "solutions", 8)
	if m.solutions != 1 || m.cycles != 0 {
		t.Fatalf("start sols=%d cycles=%d", m.solutions, m.cycles)
	}
	for i := 0; i < 91; i++ {
		m = frame(t, m)
	}
	if m.solutions != 92 || m.cycles != 0 {
		t.Fatalf("at the end sols=%d cycles=%d", m.solutions, m.cycles)
	}
	m = frame(t, m)
	if m.solutions != 1 || m.cycles != 1 {
		t.Fatalf("after the loop sols=%d cycles=%d", m.solutions, m.cycles)
	}
	if !solver.Valid(m.board) {
		t.Fatal("looped board is illegal")
	}
}

func TestModeSwitchAndSpeedAndResize(t *testing.T) {
	m := sized(mustModel(t, "step", 4), 80, 24)
	m = press(t, m, "m")
	if m.mode != modeSolutions || m.kind != solver.KindSolution || m.queensOn() != 8 {
		t.Fatalf("switch to solutions: mode=%d kind=%s queens=%d", m.mode, m.kind, m.queensOn())
	}
	m = press(t, m, "+")
	m = press(t, m, "=")
	m = press(t, m, "up")
	if m.speed != 7 {
		t.Fatalf("speed = %d, want 7", m.speed)
	}
	m = press(t, m, "8")
	if m.speed != 8 {
		t.Fatalf("speed = %d, want 8", m.speed)
	}
	m = press(t, m, "-")
	if m.speed != 7 {
		t.Fatalf("speed = %d, want 7", m.speed)
	}
	for i := 0; i < 20; i++ {
		m = press(t, m, "+")
	}
	if m.speed != maxSpeed {
		t.Fatal("speed did not clamp at the top")
	}
	m = press(t, m, "1")
	for i := 0; i < 5; i++ {
		m = press(t, m, "left")
	}
	if m.speed != 1 {
		t.Fatal("speed did not clamp at 1")
	}
	attempts := m.attempts
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = next.(Model)
	if m.attempts != attempts || m.width != 120 || m.height != 40 {
		t.Fatal("resize disturbed the search")
	}
	m = press(t, m, "tab")
	if m.mode != modeStep {
		t.Fatal("tab did not return to step mode")
	}
}

func TestRestart(t *testing.T) {
	m := mustModel(t, "step", 8)
	for i := 0; i < 15; i++ {
		m = frame(t, m)
	}
	if m.attempts <= 1 {
		t.Fatal("precondition: search should have moved")
	}
	m = press(t, m, "r")
	if m.attempts != 1 || m.solutions != 0 || m.cycles != 0 {
		t.Fatalf("restart left attempts=%d sols=%d cycles=%d", m.attempts, m.solutions, m.cycles)
	}
	if m.square(m.hiRow, m.hiCol) != "a8" {
		t.Fatalf("restart square = %s", m.square(m.hiRow, m.hiCol))
	}
}

func TestLayoutsFitCommonTerminals(t *testing.T) {
	sizes := [][2]int{{28, 15}, {60, 20}, {80, 24}, {80, 40}, {100, 36}, {120, 50}, {160, 60}}
	for _, mode := range []string{"step", "solutions"} {
		for _, sz := range sizes {
			m := sized(mustModel(t, mode, 5), sz[0], sz[1])
			v := m.View()
			w, h := lipgloss.Width(v), lipgloss.Height(v)
			if w != sz[0] || h != sz[1] {
				t.Errorf("%s %dx%d rendered %dx%d", mode, sz[0], sz[1], w, h)
			}
			plain := strip(v)
			if !strings.Contains(plain, "Eight Queens") {
				t.Errorf("%s %dx%d view missing the title", mode, sz[0], sz[1])
			}
			if sz[0] >= 80 && !strings.Contains(plain, "space pause") {
				t.Errorf("%s %dx%d view missing the footer", mode, sz[0], sz[1])
			}
		}
	}
	tiny := sized(mustModel(t, "step", 5), 20, 8)
	if !strings.Contains(strip(tiny.View()), "at least") {
		t.Fatal("tiny terminal did not ask for a resize")
	}
}

func TestPausedLabel(t *testing.T) {
	m := sized(mustModel(t, "step", 5), 100, 36)
	m = press(t, m, " ")
	view := strip(m.View())
	if !strings.Contains(view, "paused") || !strings.Contains(view, "space resume") {
		t.Fatalf("paused view:\n%s", view)
	}
}

func mustModel(t *testing.T, mode string, speed int) Model {
	t.Helper()
	m, err := New(mode, speed)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func frame(t *testing.T, m Model) Model {
	t.Helper()
	next, cmd := m.Update(frameMsg{})
	if cmd == nil {
		t.Fatal("missing follow-up tick")
	}
	return next.(Model)
}

func press(t *testing.T, m Model, key string) Model {
	t.Helper()
	next, _ := m.Update(keyMsg(key))
	got, ok := next.(Model)
	if !ok {
		t.Fatalf("update returned %T", next)
	}
	return got
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func strip(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			continue
		}
		j := i + 1
		for j < len(s) && s[j] != 'm' && s[j] != 'A' && s[j] != 'B' && s[j] != 'C' && s[j] != 'D' && s[j] != 'H' && s[j] != 'K' && s[j] != 'J' {
			j++
		}
		if j < len(s) {
			i = j
		}
	}
	return b.String()
}

func truecolorBG(hex string) string {
	var r, g, b int
	fmt.Sscanf(strings.TrimPrefix(hex, "#"), "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
}
