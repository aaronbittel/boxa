package main

type solver struct {
	algorithm solverAlgorithm
	steps     chan solverStep
	solved    chan bool
	started   bool
}

func newSolver() *solver {
	return &solver{
		steps: make(chan solverStep, 30),
		// Buffered so the solver can report the result before the game loop drains all
		// steps.
		solved: make(chan bool, 1),
	}
}

type solverStep struct {
	pos      Cell
	value    int
	selected bool
}

type solverAlgorithm func(s sudoku, steps chan<- solverStep) bool

// sudokuSolveBacktrackingMRV solves the Sudoku using backtracking, choosing the empty
// cell with the fewest candidates at each step.
func sudokuSolveBacktrackingMRV(s sudoku, steps chan<- solverStep) bool {
	if s.isSolved() {
		return true
	}

	var (
		mostConstrainedCell Cell
		possibilities       []int
		found               bool
	)

outer:
	for y := range cellCount {
		for x := range cellCount {
			if !s.at(x, y).isEmpty() {
				continue
			}

			p := s.candidates(x, y)
			if len(p) == 0 { // unsolveable
				return false
			}
			if !found || len(p) < len(possibilities) {
				possibilities = p
				mostConstrainedCell = Cell{Row: y, Col: x}
				found = true

				if len(p) == 1 {
					break outer
				}
			}
		}
	}

	for _, num := range possibilities {
		step := solverStep{
			pos:      mostConstrainedCell,
			value:    num,
			selected: true,
		}
		s.set(mostConstrainedCell, cellState{
			selected: step.selected,
			Value:    step.value,
		})
		steps <- step

		if sudokuSolveBacktrackingMRV(s, steps) {
			return true
		}

		step = solverStep{pos: mostConstrainedCell}
		s.set(mostConstrainedCell, cellState{})
		steps <- step
	}

	return false
}

func sudokuSolveBacktracking(sudoku sudoku, steps chan<- solverStep) bool {
	return sudokuSolveBacktrackingImpl(sudoku, Cell{}, steps)
}

func sudokuSolveBacktrackingImpl(sudoku sudoku, pos Cell, steps chan<- solverStep) bool {
	for !sudoku.at(pos.Col, pos.Row).isEmpty() {
		pos = pos.next()
	}

	if sudoku.isSolved() {
		return true
	}

	for n := 1; n <= cellCount; n++ {
		step := solverStep{pos: pos, value: n, selected: true}
		sudoku.set(pos, cellState{
			selected: step.selected,
			Value:    step.value,
		})
		steps <- step

		if sudoku.hasCellConflict(pos.Col, pos.Row) {
			step = solverStep{pos: pos}
			sudoku.set(pos, cellState{})
			steps <- step
			continue
		}

		next := pos.next()

		if sudokuSolveBacktrackingImpl(sudoku, next, steps) {
			return true
		}

		step = solverStep{pos: pos}
		sudoku.set(pos, cellState{})
		steps <- step
	}

	return false
}
