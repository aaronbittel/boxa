package main

import (
	"math/rand/v2"
)

const cellCount = 9

type sudoku [cellCount][cellCount]cellState

func (s *sudoku) isSolved() bool {
	for y := range cellCount {
		for x := range cellCount {
			if s.at(x, y).isEmpty() {
				return false
			}
			if s.hasCellConflict(x, y) {
				return false
			}
		}
	}
	return true
}

func (s *sudoku) reset() {
	for y := range cellCount {
		for x := range cellCount {
			if !s.at(x, y).Given {
				s[y][x] = cellState{}
			} else {
				s[y][x] = cellState{
					Value: s[y][x].Value,
					Given: s[y][x].Given,
				}
			}
		}
	}
}

func (s *sudoku) set(pos Cell, state cellState) {
	s[pos.Row][pos.Col] = state
}

func (s *sudoku) hasConflictFor(x, y, value int) bool {
	return s.hasNumberInCol(x, y, value) ||
		s.hasNumberInRow(x, y, value) ||
		s.hasNumberInBox(x, y, value)
}

func (s *sudoku) hasCellConflict(x, y int) bool {
	if s.at(x, y).isEmpty() {
		return false
	}
	return s.hasConflictFor(x, y, s.at(x, y).Value)
}

func (s *sudoku) hasNumberInRow(x, y, value int) bool {
	for xo := range cellCount {
		if x == xo {
			continue
		}
		if s.at(xo, y).Value == value {
			return true
		}
	}
	return false
}

func (s *sudoku) hasNumberInCol(x, y, value int) bool {
	for yo := range cellCount {
		if y == yo {
			continue
		}
		if s.at(x, yo).Value == value {
			return true
		}
	}
	return false
}

func (s *sudoku) hasNumberInBox(x, y, value int) bool {
	boxX := (x / 3) * 3
	boxY := (y / 3) * 3

	for dx := range 3 {
		for dy := range 3 {
			xo := boxX + dx
			yo := boxY + dy
			if xo == x && yo == y {
				continue
			}
			if s.at(xo, yo).Value == value {
				return true
			}
		}
	}
	return false
}

func (s *sudoku) selectIf(predicate func(cell cellState) bool) {
	for y := range cellCount {
		for x := range cellCount {
			if predicate(s[y][x]) {
				s[y][x].selected = true
			}
		}
	}
}

func (s *sudoku) at(x, y int) *cellState {
	if x < 0 || x >= cellCount || y < 0 || y >= cellCount {
		panic("invalid position")
	}
	return &s[y][x]
}

func (s *sudoku) forEachSelectedCell(fn func(pos Cell, cell *cellState)) {
	for y := range cellCount {
		for x := range cellCount {
			if !s[y][x].selected {
				continue
			}
			pos := Cell{
				Row: y,
				Col: x,
			}
			fn(pos, &s[y][x])
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

func (s *sudoku) unselectAllCells() {
	for y := range cellCount {
		for x := range cellCount {
			s[y][x].selected = false
		}
	}
}

func initFilledSudoku() *sudoku {
	var sudoku sudoku
	for y := range cellCount {
		for x := range cellCount {
			if rand.IntN(100) < 20 {
				sudoku[y][x].Value = rand.IntN(cellCount) + 1
				sudoku[y][x].Given = true
			}
		}
	}
	return &sudoku
}
