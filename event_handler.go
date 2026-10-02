package main

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

func (g *gameState) handlePlayingEvent(event Event) error {
	switch event.Type {
	case EventMousePressed:
		g.handleSingleClick(event)
	case EventMouseCellEntered:
		g.handleDragging(event)
	case EventMouseDoubleClick:
		g.handleDoubleClick(event)
	case EventKeyPressed:
		return g.handleKeyEvent(event)
	}

	return nil
}

func (g *gameState) handleKeyEvent(event Event) error {
	switch event.Key {
	case KeyOne, KeyTwo, KeyThree, KeyFour, KeyFive, KeySix, KeySeven, KeyEight, KeyNine:
		g.handleNumberKey(event)
	case KeyDelete:
		g.handleDeleteKey(event)
	case KeyArrowUp, KeyArrowDown, KeyArrowRight, KeyArrowLeft:
		g.handleArrowKey(event)
	case KeyD: // Debug
		if event.Modifiers.Ctrl {
			debug = !debug
		}
	case KeyS:
		if event.Modifiers.Ctrl {
			g.mode = modeSolving
		}
	case KeyR:
		if event.Modifiers.Ctrl {
			g.sudoku.reset()
			g.undoHistory = &undoHistory{}
		}
	case KeyY:
		if event.Modifiers.Ctrl {
			g.undoHistory.redo(g.sudoku)
		}
	case KeyZ:
		if event.Modifiers.Ctrl {
			g.undoHistory.undo(g.sudoku)
		}
	}

	return nil
}

// Selection precedence:
//  1. Modifier-specific selection is attempted first:
//     - Ctrl + Shift: coloring
//     - Ctrl:         center marks
//     - Shift:        corner marks
//  2. If the requested modifier-specific property is not present,
//     selection falls back to the normal priority:
//     - number
//     - coloring
//     - center marks
//     - corner marks
//
// The first matching rule wins.
func (g *gameState) handleDoubleClick(event Event) {
	g.currentCell = event.Cell

	if !event.Modifiers.Ctrl {
		g.sudoku.deselectAllCells()
	}

	clickedCell := g.sudoku.at(event.Cell.Col, event.Cell.Row)
	clickedCell.selected = true

	g.sudoku.selectIf(func(candidate cellState) bool {
		switch {
		case event.Modifiers.Ctrl && event.Modifiers.Shift && clickedCell.isColored():
			return candidate.hasAllColors(*clickedCell)
		case event.Modifiers.Ctrl && clickedCell.hasCenterMarks():
			return candidate.isEmpty() && candidate.containsCenterMarksOf(*clickedCell)
		case event.Modifiers.Shift && clickedCell.hasCornerMarks():
			return candidate.isEmpty() && candidate.containsCornerMarksOf(*clickedCell)
		default:
			switch {
			case !clickedCell.isEmpty():
				return candidate.Value == clickedCell.Value
			case clickedCell.isColored():
				return candidate.hasAllColors(*clickedCell)
			case clickedCell.hasCenterMarks():
				return candidate.isEmpty() && candidate.containsCenterMarksOf(*clickedCell)
			case clickedCell.hasCornerMarks():
				return candidate.isEmpty() && candidate.containsCornerMarksOf(*clickedCell)
			}
		}
		return false
	})
}

func (g *gameState) handleDragging(event Event) {
	g.currentCell = event.Cell

	switch g.selectionMode {
	case selectionSelect:
		g.sudoku.selectCell(event.Cell.Col, event.Cell.Row)
	case selectionDeselect:
		g.sudoku.deselectCell(event.Cell.Col, event.Cell.Row)
	}
}

func (g *gameState) handleSingleClick(event Event) {
	g.currentCell = event.Cell

	if !event.Modifiers.Ctrl {
		g.selectionMode = selectionSelect
		g.sudoku.deselectAllCells()
		g.sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
		g.selectionMode = selectionSelect
		return
	}

	if g.sudoku.at(event.Cell.Col, event.Cell.Row).selected {
		g.selectionMode = selectionDeselect
	} else {
		g.selectionMode = selectionSelect
	}
	g.sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
}

func (g *gameState) handleDeleteKey(event Event) {
	g.undoHistory.begin()
	defer g.undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			cell.Colors.clear()
		})
	case event.Modifiers.Shift:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCornerMarks()
			g.undoHistory.record(pos, before, *cell)
		})
	case event.Modifiers.Ctrl:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCenterMarks()
			g.undoHistory.record(pos, before, *cell)
		})
	default:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearNumber()
			g.undoHistory.record(pos, before, *cell)
		})
	}
}

func (g *gameState) handleArrowKey(event Event) {
	moveDir, ok := arrowKeyDirection(event.Key)
	if !ok {
		return
	}

	g.currentCell.move(moveDir)

	if !event.Modifiers.Ctrl && !event.Modifiers.Shift {
		g.sudoku.deselectAllCells()
		g.sudoku.selectCell(g.currentCell.Col, g.currentCell.Row)
		return
	}

	switch g.selectionMode {
	case selectionSelect:
		g.sudoku.selectCell(g.currentCell.Col, g.currentCell.Row)
	case selectionDeselect:
		g.sudoku.deselectCell(g.currentCell.Col, g.currentCell.Row)
	}
}

func (g *gameState) handleNumberKey(event Event) {
	value := event.Key.Value()
	index := value - 1

	g.undoHistory.begin()
	defer g.undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		if index < len(cellColors) {
			g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
				before := *cell
				cell.toggleColor(index)
				g.undoHistory.record(pos, before, *cell)
			})
		}
	case event.Modifiers.Shift:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCornerMark(index)
				g.undoHistory.record(pos, before, *cell)
			}
		})
	case event.Modifiers.Ctrl:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCenterMark(index)
				g.undoHistory.record(pos, before, *cell)
			}
		})
	default:
		g.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if !cell.Given {
				before := *cell
				cell.Value = value
				g.undoHistory.record(pos, before, *cell)
			}
		})
	}
}

func loadSudokuFromClipboard() (*gameState, error) {
	id, err := clipboard.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pasting of clipboard content failed: %w", err)
	}

	id = strings.TrimSpace(id)

	newGameState, err := loadGameState(id)
	if err != nil {
		return nil, fmt.Errorf("loading of sudoku with id %q failed: %w", id, err)
	}

	return newGameState, nil
}
