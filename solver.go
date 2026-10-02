package main

type solver struct {
	changes chan solverStep
	done    chan bool
	started bool
}

func newSolver() *solver {
	return &solver{
		changes: make(chan solverStep, 30),
		done:    make(chan bool, 1),
	}
}

type solverStep struct {
	pos      Cell
	value    int
	selected bool
}

func sudokuSolveBacktracking(sudoku sudoku, pos Cell, changes chan<- solverStep) bool {
	for !sudoku.at(pos.Col, pos.Row).isEmpty() {
		pos = pos.next()
	}

	if sudoku.isSolved() {
		return true
	}

	for n := 1; n <= cellCount; n++ {
		change := solverStep{pos: pos, value: n, selected: true}
		sudoku.set(pos, cellState{
			selected: change.selected,
			Value:    change.value,
		})
		changes <- change

		if sudoku.hasCellConflict(pos.Col, pos.Row) {
			change = solverStep{pos: pos}
			sudoku.set(pos, cellState{})
			changes <- change
			continue
		}

		next := pos.next()

		if sudokuSolveBacktracking(sudoku, next, changes) {
			return true
		}

		change = solverStep{pos: pos}
		sudoku.set(pos, cellState{})
		changes <- change
	}

	return false
}
