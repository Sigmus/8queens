// Package solver searches for solutions to the n-queens problem.
//
// The walk is ordinary backtracking: one queen per rank, trying files from
// left to right, giving up a rank when every file on it is attacked. Search
// state is exposed as a pull iterator of discrete events so a UI can draw one
// attempt per frame. The solver never sleeps and never reads input.
package solver

// Pos is a square on the board. Row 0 is the first rank filled (drawn as
// rank n at the top of a chessboard). Column 0 is the a-file.
type Pos struct {
	Row int
	Col int
}

// Kind is one step a UI can draw.
type Kind int

const (
	KindNone Kind = iota
	KindPlace
	KindConflict
	KindSolution
	KindBacktrack
	KindDone
)

func (k Kind) String() string {
	switch k {
	case KindPlace:
		return "place"
	case KindConflict:
		return "conflict"
	case KindSolution:
		return "solution"
	case KindBacktrack:
		return "backtrack"
	case KindDone:
		return "done"
	default:
		return "none"
	}
}

// Event is a single visible step of the search.
//
// Board holds the queens still on the board after the step, one file per
// rank, or -1 when that rank is empty. The slice is owned by the caller.
// Attacks and Rays are set for conflicts: Attacks are queens that can capture
// the square being tried, and Rays are the squares strictly between them.
type Event struct {
	Kind     Kind
	Row      int
	Col      int
	Board    []int
	Attacks  []Pos
	Rays     []Pos
	Attempt  int
	Solution int
}

// Solutions returns every solution for an n-queens board, in the order the
// backtracking search finds them. Each solution is a slice of length n whose
// value at index r is the file of the queen on rank r.
func Solutions(n int) [][]int {
	if n < 1 {
		return nil
	}
	board := emptyBoard(n)
	var out [][]int
	var walk func(int)
	walk = func(row int) {
		if row == n {
			out = append(out, clone(board))
			return
		}
		for col := 0; col < n; col++ {
			if !safe(board, row, col) {
				continue
			}
			board[row] = col
			walk(row + 1)
			board[row] = -1
		}
	}
	walk(0)
	return out
}

// Valid reports whether board is a complete n-queens solution.
func Valid(board []int) bool {
	n := len(board)
	if n == 0 {
		return false
	}
	used := make([]bool, n)
	for r, c := range board {
		if c < 0 || c >= n || used[c] {
			return false
		}
		used[c] = true
		for prev := 0; prev < r; prev++ {
			d := c - board[prev]
			if d < 0 {
				d = -d
			}
			if d == r-prev {
				return false
			}
		}
	}
	return true
}

// Attacks returns the queens already placed above row that can capture
// (row, col), and the open squares lying strictly between each of them and
// that square.
func Attacks(board []int, row, col int) (queens, rays []Pos) {
	for r := 0; r < row && r < len(board); r++ {
		c := board[r]
		if c < 0 {
			continue
		}
		dc := col - c
		if dc < 0 {
			dc = -dc
		}
		if c != col && dc != row-r {
			continue
		}
		queens = append(queens, Pos{Row: r, Col: c})
		stepC := 0
		if col > c {
			stepC = 1
		} else if col < c {
			stepC = -1
		}
		for rr, cc := r+1, c+stepC; rr != row || cc != col; rr, cc = rr+1, cc+stepC {
			rays = append(rays, Pos{Row: rr, Col: cc})
		}
	}
	return queens, rays
}

// Stepper yields every placement, conflict, solution, and backtrack of one
// complete search, then a final KindDone event. Next after Done keeps
// returning that same Done event until Reset.
type Stepper struct {
	n      int
	events []Event
	i      int
}

// NewStepper prepares a full trace for an n-queens search.
func NewStepper(n int) *Stepper {
	if n < 1 {
		return &Stepper{events: []Event{{Kind: KindDone, Board: []int{}}}}
	}
	return &Stepper{n: n, events: trace(n)}
}

// N is the board size this trace was built for.
func (s *Stepper) N() int { return s.n }

// Total is the number of solutions in the trace.
func (s *Stepper) Total() int {
	if len(s.events) == 0 {
		return 0
	}
	return s.events[len(s.events)-1].Solution
}

// Reset rewinds the trace to the first event.
func (s *Stepper) Reset() { s.i = 0 }

// Next returns the next event. The returned slices must not be mutated if the
// caller wants later events to stay intact; they are not shared with later
// events, but they alias the trace's own storage.
func (s *Stepper) Next() Event {
	if len(s.events) == 0 {
		return Event{Kind: KindDone}
	}
	if s.i >= len(s.events) {
		return s.events[len(s.events)-1]
	}
	ev := s.events[s.i]
	s.i++
	return ev
}

func trace(n int) []Event {
	board := emptyBoard(n)
	events := make([]Event, 0, 1024)
	var attempts, sols int

	var walk func(int)
	walk = func(row int) {
		if row == n {
			events = append(events, Event{
				Kind:     KindSolution,
				Row:      n - 1,
				Col:      board[n-1],
				Board:    clone(board),
				Attempt:  attempts,
				Solution: sols,
			})
			return
		}
		for col := 0; col < n; col++ {
			attempts++
			queens, rays := Attacks(board, row, col)
			if len(queens) > 0 {
				events = append(events, Event{
					Kind:     KindConflict,
					Row:      row,
					Col:      col,
					Board:    clone(board),
					Attacks:  queens,
					Rays:     rays,
					Attempt:  attempts,
					Solution: sols,
				})
				continue
			}
			board[row] = col
			if row == n-1 {
				sols++
			}
			events = append(events, Event{
				Kind:     KindPlace,
				Row:      row,
				Col:      col,
				Board:    clone(board),
				Attempt:  attempts,
				Solution: sols,
			})
			walk(row + 1)
			board[row] = -1
			events = append(events, Event{
				Kind:     KindBacktrack,
				Row:      row,
				Col:      col,
				Board:    clone(board),
				Attempt:  attempts,
				Solution: sols,
			})
		}
	}
	walk(0)
	events = append(events, Event{
		Kind:     KindDone,
		Board:    clone(board),
		Attempt:  attempts,
		Solution: sols,
	})
	return events
}

func safe(board []int, row, col int) bool {
	queens, _ := Attacks(board, row, col)
	return len(queens) == 0
}

func emptyBoard(n int) []int {
	board := make([]int, n)
	for i := range board {
		board[i] = -1
	}
	return board
}

func clone(board []int) []int {
	out := make([]int, len(board))
	copy(out, board)
	return out
}
