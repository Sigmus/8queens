package solver

import (
	"reflect"
	"testing"
)

func TestSolutionCounts(t *testing.T) {
	// Classic sequence: n = 1..8.
	want := []int{0, 1, 0, 0, 2, 10, 4, 40, 92}
	for n := 1; n <= 8; n++ {
		got := len(Solutions(n))
		if got != want[n] {
			t.Errorf("n=%d: got %d solutions, want %d", n, got, want[n])
		}
	}
}

func TestEightQueensFinds92(t *testing.T) {
	sols := Solutions(8)
	if len(sols) != 92 {
		t.Fatalf("got %d solutions, want 92", len(sols))
	}
}

func TestSolutionsAreValidAndUnique(t *testing.T) {
	for n := 1; n <= 8; n++ {
		seen := make(map[string]struct{})
		for i, board := range Solutions(n) {
			if len(board) != n {
				t.Fatalf("n=%d solution %d has length %d", n, i, len(board))
			}
			if !Valid(board) {
				t.Fatalf("n=%d solution %d is illegal: %v", n, i, board)
			}
			key := sprint(board)
			if _, ok := seen[key]; ok {
				t.Fatalf("n=%d duplicate solution %v", n, board)
			}
			seen[key] = struct{}{}
		}
	}
}

func TestInvalidBoards(t *testing.T) {
	cases := [][]int{
		nil,
		{},
		{0, 0},
		{0, 1},
		{0, 2, 1},
		{0, 4, 7, 5, 2, 6, 1, 2},
	}
	for _, board := range cases {
		if Valid(board) {
			t.Errorf("Valid(%v) = true, want false", board)
		}
	}
	if !Valid([]int{0, 4, 7, 5, 2, 6, 1, 3}) {
		t.Fatal("known first 8-queens solution reported invalid")
	}
}

func TestFirstSolutions(t *testing.T) {
	if got, want := Solutions(4), [][]int{{1, 3, 0, 2}, {2, 0, 3, 1}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("n=4 solutions = %v, want %v", got, want)
	}
	got := Solutions(8)
	want := []int{0, 4, 7, 5, 2, 6, 1, 3}
	if !reflect.DeepEqual(got[0], want) {
		t.Fatalf("first 8-queens solution = %v, want %v", got[0], want)
	}
}

func TestStepperMatchesSolutions(t *testing.T) {
	for n := 1; n <= 8; n++ {
		want := Solutions(n)
		st := NewStepper(n)
		if st.Total() != len(want) {
			t.Fatalf("n=%d total = %d, want %d", n, st.Total(), len(want))
		}
		var got [][]int
		var places, conflicts, backs, sols, dones int
		var prev Event
		for {
			ev := st.Next()
			switch ev.Kind {
			case KindPlace:
				places++
				if ev.Board[ev.Row] != ev.Col {
					t.Fatalf("n=%d place not recorded on board: %+v", n, ev)
				}
			case KindConflict:
				conflicts++
				if len(ev.Attacks) == 0 {
					t.Fatalf("n=%d conflict without an attacker: %+v", n, ev)
				}
				if ev.Row < len(ev.Board) && ev.Board[ev.Row] != -1 {
					t.Fatalf("n=%d conflict left a queen on the rank: %+v", n, ev)
				}
			case KindSolution:
				sols++
				if !Valid(ev.Board) {
					t.Fatalf("n=%d stepper solution invalid: %v", n, ev.Board)
				}
				if prev.Kind != KindPlace || prev.Row != ev.Row || prev.Col != ev.Col {
					t.Fatalf("n=%d solution not preceded by the placing step", n)
				}
				got = append(got, append([]int(nil), ev.Board...))
			case KindBacktrack:
				backs++
				if ev.Row < len(ev.Board) && ev.Board[ev.Row] != -1 {
					t.Fatalf("n=%d backtrack did not clear the rank: %+v", n, ev)
				}
				if prev.Kind == KindSolution && (prev.Row != ev.Row || prev.Col != ev.Col) {
					t.Fatalf("n=%d backtrack after solution left a different square", n)
				}
			case KindDone:
				dones++
				if dones == 1 && (ev.Attempt != places+conflicts || ev.Solution != len(want)) {
					t.Fatalf("n=%d done attempt=%d place+conflict=%d sols=%d", n, ev.Attempt, places+conflicts, ev.Solution)
				}
			default:
				t.Fatalf("n=%d unexpected kind %s", n, ev.Kind)
			}
			if ev.Kind == KindDone {
				again := st.Next()
				if again.Kind != KindDone {
					t.Fatal("Done was not sticky")
				}
				break
			}
			prev = ev
		}
		if sols != len(want) || !reflect.DeepEqual(got, want) {
			t.Fatalf("n=%d stepper solutions = %d, want %d (match %v)", n, sols, len(want), reflect.DeepEqual(got, want))
		}
		if places == 0 || (n >= 4 && conflicts == 0) || backs != places {
			t.Fatalf("n=%d unexpected shape places=%d conflicts=%d backs=%d", n, places, conflicts, backs)
		}
	}
}

func TestEventBoardsAreIndependent(t *testing.T) {
	st := NewStepper(4)
	first := st.Next()
	first.Board[0] = 99
	second := st.Next()
	if second.Board[0] == 99 {
		t.Fatal("event boards alias each other")
	}
}

func TestResetReplays(t *testing.T) {
	st := NewStepper(5)
	var first []Kind
	for i := 0; i < 6; i++ {
		first = append(first, st.Next().Kind)
	}
	st.Reset()
	for i, kind := range first {
		if got := st.Next().Kind; got != kind {
			t.Fatalf("event %d after reset = %s, want %s", i, got, kind)
		}
	}
}

func TestAttackRays(t *testing.T) {
	board := []int{0, -1, -1, -1, -1, -1, -1, -1}
	queens, rays := Attacks(board, 3, 0)
	if !reflect.DeepEqual(queens, []Pos{{0, 0}}) {
		t.Fatalf("file queens = %v", queens)
	}
	if !reflect.DeepEqual(rays, []Pos{{1, 0}, {2, 0}}) {
		t.Fatalf("file rays = %v", rays)
	}
	queens, rays = Attacks(board, 3, 3)
	if !reflect.DeepEqual(queens, []Pos{{0, 0}}) {
		t.Fatalf("diagonal queens = %v", queens)
	}
	if !reflect.DeepEqual(rays, []Pos{{1, 1}, {2, 2}}) {
		t.Fatalf("diagonal rays = %v", rays)
	}
	queens, _ = Attacks(board, 2, 1)
	if len(queens) != 0 {
		t.Fatalf("expected a safe square, attacked by %v", queens)
	}
}

func TestN2HasNoSolution(t *testing.T) {
	if got := Solutions(2); len(got) != 0 {
		t.Fatalf("n=2 solutions = %v", got)
	}
	st := NewStepper(2)
	var kinds []Kind
	for {
		ev := st.Next()
		kinds = append(kinds, ev.Kind)
		if ev.Kind == KindDone {
			break
		}
	}
	want := []Kind{
		KindPlace,     // (0,0)
		KindConflict,  // (1,0) file
		KindConflict,  // (1,1) diagonal
		KindBacktrack, // remove (0,0)
		KindPlace,     // (0,1)
		KindConflict,  // (1,0) diagonal
		KindConflict,  // (1,1) file
		KindBacktrack,
		KindDone,
	}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatalf("n=2 trace = %v, want %v", kinds, want)
	}
}

func sprint(board []int) string {
	b := make([]byte, len(board))
	for i, c := range board {
		b[i] = byte(c + 1)
	}
	return string(b)
}
