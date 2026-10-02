package main

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

func handleEvent(event Event, game *gameState) (*gameState, error) {
	switch event.Type {
	case EventMousePressed:
		handleSingleClick(event, game)
	case EventMouseCellEntered:
		handleDragging(event, game)
	case EventMouseDoubleClick:
		handleDoubleClick(event, game)
	case EventKeyPressed:
		return handleKeyEvent(event, game)
	}

	return game, nil
}

func handleKeyEvent(event Event, game *gameState) (*gameState, error) {
	switch event.Key {
	case KeyOne, KeyTwo, KeyThree, KeyFour, KeyFive, KeySix, KeySeven, KeyEight, KeyNine:
		handleNumberKey(event, game)
	case KeyDelete:
		handleDeleteKey(event, game)
	case KeyArrowUp, KeyArrowDown, KeyArrowRight, KeyArrowLeft:
		handleArrowKey(event, game)
	case KeyD: // Debug
		if event.Modifiers.Ctrl {
			debug = !debug
		}
	case KeyV:
		if event.Modifiers.Ctrl {
			newGameState, err := pasteSudoku()
			if err != nil {
				return nil, err
			}
			return newGameState, nil
		}
	case KeyR:
		if event.Modifiers.Ctrl {
			game.sudoku.reset()
			game.undoHistory = &undoHistory{}
		}
	case KeyY:
		if event.Modifiers.Ctrl {
			game.undoHistory.redo(game.sudoku)
		}
	case KeyZ:
		if event.Modifiers.Ctrl {
			game.undoHistory.undo(game.sudoku)
		}
	}

	return game, nil
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
func handleDoubleClick(event Event, gameState *gameState) {
	gameState.currentCell = event.Cell

	if !event.Modifiers.Ctrl {
		gameState.sudoku.unselectAllCells()
	}

	clickedCell := gameState.sudoku.at(event.Cell.Col, event.Cell.Row)
	clickedCell.selected = true

	gameState.sudoku.selectIf(func(candidate cellState) bool {
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

func handleDragging(event Event, gameState *gameState) {
	gameState.currentCell = event.Cell

	switch gameState.selectionMode {
	case selectionSelect:
		gameState.sudoku.selectCell(event.Cell.Col, event.Cell.Row)
	case selectionDeselect:
		gameState.sudoku.deselectCell(event.Cell.Col, event.Cell.Row)
	}
}

func handleSingleClick(event Event, gameState *gameState) {
	gameState.currentCell = event.Cell

	if !event.Modifiers.Ctrl {
		gameState.selectionMode = selectionSelect
		gameState.sudoku.unselectAllCells()
		gameState.sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
		gameState.selectionMode = selectionSelect
		return
	}

	if gameState.sudoku.isSelected(event.Cell.Col, event.Cell.Row) {
		gameState.selectionMode = selectionDeselect
	} else {
		gameState.selectionMode = selectionSelect
	}
	gameState.sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
}

func handleDeleteKey(event Event, gameState *gameState) {
	gameState.undoHistory.begin()
	defer gameState.undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			cell.Colors.clear()
		})
	case event.Modifiers.Shift:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCornerMarks()
			gameState.undoHistory.record(pos, before, *cell)
		})
	case event.Modifiers.Ctrl:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCenterMarks()
			gameState.undoHistory.record(pos, before, *cell)
		})
	default:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearNumber()
			gameState.undoHistory.record(pos, before, *cell)
		})
	}
}

func handleArrowKey(event Event, gameState *gameState) {
	moveDir, ok := arrowKeyDirection(event.Key)
	if !ok {
		return
	}

	gameState.currentCell.move(moveDir)

	if !event.Modifiers.Ctrl && !event.Modifiers.Shift {
		gameState.sudoku.unselectAllCells()
		gameState.sudoku.selectCell(gameState.currentCell.Col, gameState.currentCell.Row)
		return
	}

	switch gameState.selectionMode {
	case selectionSelect:
		gameState.sudoku.selectCell(gameState.currentCell.Col, gameState.currentCell.Row)
	case selectionDeselect:
		gameState.sudoku.deselectCell(gameState.currentCell.Col, gameState.currentCell.Row)
	}
}

func handleNumberKey(event Event, gameState *gameState) {
	value := event.Key.Value()
	index := value - 1

	gameState.undoHistory.begin()
	defer gameState.undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		if index < len(cellColors) {
			gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
				before := *cell
				cell.toggleColor(index)
				gameState.undoHistory.record(pos, before, *cell)
			})
		}
	case event.Modifiers.Shift:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCornerMark(index)
				gameState.undoHistory.record(pos, before, *cell)
			}
		})
	case event.Modifiers.Ctrl:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCenterMark(index)
				gameState.undoHistory.record(pos, before, *cell)
			}
		})
	default:
		gameState.sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if !cell.Given {
				before := *cell
				cell.Value = value
				gameState.undoHistory.record(pos, before, *cell)
			}
		})
	}
}

func pasteSudoku() (*gameState, error) {
	id, err := clipboard.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("WARNING: pasting of clipboard content failed: %w", err)
	}

	id = strings.TrimSpace(id)

	newGameState, err := loadGameState(id)
	if err != nil {
		return nil, fmt.Errorf("WARNING: loading of sudoku with id %q failed: %w", id, err)
	}

	return newGameState, nil
}
