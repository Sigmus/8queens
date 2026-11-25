// Package ui draws the eight-queens search in the terminal.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"eightqueens/solver"
)

const (
	boardN      = 8
	maxSpeed    = 8
	blinkFrames = 8
	blinkEvery  = 120 * time.Millisecond
)

type mode int

const (
	modeStep mode = iota
	modeSolutions
)

// Model is the Bubble Tea model. The solver is a pull iterator; only a frame
// tick advances it, and only while the search is running.
type Model struct {
	step   *solver.Stepper
	mode   mode
	speed  int
	paused bool
	cycles int
	total  int

	width  int
	height int

	kind      solver.Kind
	board     []int
	hiRow     int
	hiCol     int
	attacks   []solver.Pos
	rays      []solver.Pos
	attempts  int
	solutions int

	celebrating bool
	blinkOn     bool
	blinkPhase  int
}

type frameMsg time.Time

// New builds a model and applies the first step immediately, so the first
// paint is already a live board. modeFlag is "step" or "solutions". speed is
// 1 (slowest) through 8 (fastest).
func New(modeFlag string, speed int) (Model, error) {
	md, err := parseMode(modeFlag)
	if err != nil {
		return Model{}, err
	}
	if speed < 1 || speed > maxSpeed {
		return Model{}, fmt.Errorf("speed %d is outside 1–%d", speed, maxSpeed)
	}
	step := solver.NewStepper(boardN)
	m := Model{
		step:  step,
		mode:  md,
		speed: speed,
		total: step.Total(),
		board: emptyBoard(),
		hiRow: -1,
		hiCol: -1,
	}
	m.advance()
	return m, nil
}

func parseMode(flag string) (mode, error) {
	switch strings.ToLower(strings.TrimSpace(flag)) {
	case "step", "steps":
		return modeStep, nil
	case "solutions", "solution", "sols":
		return modeSolutions, nil
	default:
		return 0, fmt.Errorf("unknown mode %q (use step or solutions)", flag)
	}
}

func (m Model) Init() tea.Cmd {
	return frameTick(m.interval())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case frameMsg:
		return m.onFrame()
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "Q", "ctrl+c":
		return m, tea.Quit
	case " ":
		m.paused = !m.paused
		return m, nil
	case "m", "M", "tab":
		m.toggleMode()
		return m, nil
	case "+", "=", "up", "right":
		if m.speed < maxSpeed {
			m.speed++
		}
		return m, nil
	case "-", "_", "down", "left":
		if m.speed > 1 {
			m.speed--
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8":
		m.speed = int(msg.String()[0] - '0')
		return m, nil
	case "r", "R":
		m.restart()
		return m, nil
	}
	return m, nil
}

func (m Model) onFrame() (tea.Model, tea.Cmd) {
	if m.paused {
		return m, frameTick(m.interval())
	}
	if m.celebrating {
		m.blinkPhase++
		if m.blinkPhase >= blinkFrames {
			m.celebrating = false
			m.blinkOn = false
			m.blinkPhase = 0
			return m, frameTick(m.interval())
		}
		m.blinkOn = m.blinkPhase%2 == 0
		return m, frameTick(blinkEvery)
	}
	m.advance()
	if m.celebrating {
		return m, frameTick(blinkEvery)
	}
	return m, frameTick(m.interval())
}

func (m *Model) advance() {
	if m.mode == modeSolutions {
		m.forwardToSolution()
		return
	}
	ev := m.nextEvent()
	m.apply(ev)
	if ev.Kind == solver.KindSolution {
		m.celebrating = true
		m.blinkOn = true
		m.blinkPhase = 0
	}
}

func (m *Model) nextEvent() solver.Event {
	ev := m.step.Next()
	if ev.Kind == solver.KindDone {
		m.cycles++
		m.step.Reset()
		ev = m.step.Next()
	}
	return ev
}

// forwardToSolution consumes events until the next complete solution. One
// wrap past the end of the trace is enough for n = 8, which has solutions.
func (m *Model) forwardToSolution() {
	for wrap := 0; wrap < 2; wrap++ {
		for {
			ev := m.step.Next()
			if ev.Kind == solver.KindDone {
				m.cycles++
				m.step.Reset()
				break
			}
			if ev.Kind == solver.KindSolution {
				m.apply(ev)
				return
			}
		}
	}
}

func (m *Model) apply(ev solver.Event) {
	m.kind = ev.Kind
	if ev.Board != nil {
		m.board = ev.Board
	}
	m.hiRow = ev.Row
	m.hiCol = ev.Col
	m.attacks = ev.Attacks
	m.rays = ev.Rays
	m.attempts = ev.Attempt
	m.solutions = ev.Solution
}

func (m *Model) toggleMode() {
	m.celebrating = false
	m.blinkOn = false
	m.blinkPhase = 0
	if m.mode == modeStep {
		m.mode = modeSolutions
		if m.kind != solver.KindSolution {
			m.forwardToSolution()
		}
		return
	}
	m.mode = modeStep
}

func (m *Model) restart() {
	m.step.Reset()
	m.cycles = 0
	m.celebrating = false
	m.blinkOn = false
	m.blinkPhase = 0
	m.kind = solver.KindNone
	m.board = emptyBoard()
	m.hiRow, m.hiCol = -1, -1
	m.attacks = nil
	m.rays = nil
	m.attempts = 0
	m.solutions = 0
	if !m.paused {
		m.advance()
	}
}

func (m Model) interval() time.Duration {
	if m.mode == modeSolutions {
		return [...]time.Duration{
			0,
			1600 * time.Millisecond,
			1100 * time.Millisecond,
			800 * time.Millisecond,
			550 * time.Millisecond,
			400 * time.Millisecond,
			280 * time.Millisecond,
			180 * time.Millisecond,
			110 * time.Millisecond,
		}[m.speed]
	}
	return [...]time.Duration{
		0,
		460 * time.Millisecond,
		280 * time.Millisecond,
		180 * time.Millisecond,
		110 * time.Millisecond,
		70 * time.Millisecond,
		42 * time.Millisecond,
		24 * time.Millisecond,
		12 * time.Millisecond,
	}[m.speed]
}

func frameTick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}

func emptyBoard() []int {
	board := make([]int, boardN)
	for i := range board {
		board[i] = -1
	}
	return board
}

func (m Model) square(row, col int) string {
	if row < 0 || col < 0 || col >= boardN {
		return "—"
	}
	return fmt.Sprintf("%c%d", 'a'+col, boardN-row)
}

func (m Model) queensOn() int {
	n := 0
	for _, c := range m.board {
		if c >= 0 {
			n++
		}
	}
	return n
}

func (m Model) modeName() string {
	if m.mode == modeSolutions {
		return "solutions"
	}
	return "step"
}
