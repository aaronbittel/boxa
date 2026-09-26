package main

import (
	"math/rand/v2"
	"slices"
)

const emptyCell = 0

type sudoku [cellCount][cellCount]cellState

type cellState struct {
	value       int
	selected    bool
	cornerMarks [cellCount]bool
	centerMarks [cellCount]bool
}

func (c *cellState) toggleCornerMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid corner index")
	}
	c.cornerMarks[index] = !c.cornerMarks[index]
}

func (c *cellState) toggleCenterMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid center index")
	}
	c.centerMarks[index] = !c.centerMarks[index]
}

func (c *cellState) clearCornerMarks() {
	c.cornerMarks = [cellCount]bool{}
}

func (c *cellState) clearCenterMarks() {
	c.centerMarks = [cellCount]bool{}
}

func (c *cellState) clearNumber() {
	c.value = emptyCell
}

func (c *cellState) isEmpty() bool {
	return c.value == emptyCell
}

func (c *cellState) hasCenterMarks() bool {
	return slices.Contains(c.centerMarks[:], true)
}

func (c *cellState) hasCornerMarks() bool {
	return slices.Contains(c.cornerMarks[:], true)
}

func (c *cellState) containsCenterMarksOf(other cellState) bool {
	for i, marked := range other.centerMarks {
		if marked && !c.centerMarks[i] {
			return false
		}
	}
	return true
}

func (c *cellState) containsCornerMarksOf(other cellState) bool {
	for i, marked := range other.cornerMarks {
		if marked && !c.cornerMarks[i] {
			return false
		}
	}
	return true
}

func (s *sudoku) toggleCornerMark(num int) {
	s.forEachSelectedCell(func(cell *cellState) {
		cell.cornerMarks[num] = !cell.cornerMarks[num]
	})
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

func (s *sudoku) at(x, y int) cellState {
	if x < 0 || x >= cellCount || y < 0 || y >= cellCount {
		panic("invalid position")
	}
	return s[y][x]
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
	return s[y][x].isEmpty()
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
