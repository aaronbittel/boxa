package main

import "math/rand/v2"

const emptyCell = 0

type sudoku [cellCount][cellCount]cellState

type cellState struct {
	value       int
	selected    bool
	cornerMarks [cellCount]bool
	centerMarks [cellCount]bool
}

func (s *sudoku) toggleCornerMark(num int) {
	s.forEachSelectedCell(func(cell *cellState) {
		cell.cornerMarks[num] = !cell.cornerMarks[num]
	})
}

func (s *sudoku) forEachSelectedCell(fn func(cell *cellState)) {
	for y := range cellCount {
		for x := range cellCount {
			if !s[y][x].selected {
				continue
			}
			fn(&s[y][x])
		}
	}
}

func (s *sudoku) isSelected(x, y int) bool {
	return s[y][x].selected
}

func (s *sudoku) selectCell(x, y int) {
	s[y][x].selected = true
}

func (s *sudoku) deselectCell(x, y int) {
	s[y][x].selected = false
}

func (s *sudoku) toggleSelection(x, y int) {
	s[y][x].selected = !s[y][x].selected
}

func (s sudoku) isEmpty(x, y int) bool {
	return s[y][x].value == emptyCell
}

func (s *sudoku) unselectAllCells() {
	for y := range cellCount {
		for x := range cellCount {
			s[y][x].selected = false
		}
	}
}

func initFilledSudoku() sudoku {
	s := sudoku{}
	for y := range cellCount {
		for x := range cellCount {
			if rand.IntN(100) < 20 {
				s[y][x].value = rand.IntN(cellCount) + 1
			}
		}
	}
	return s
}
