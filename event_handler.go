package main

func handleEvent(event Event, gameState *gameState) {
	switch event.Type {
	case EventMousePressed:
		handleSingleClick(event, gameState)
	case EventMouseCellEntered:
		handleDragging(event, gameState)
	case EventMouseDoubleClick:
		handleDoubleClick(event, gameState)
	case EventKeyPressed:
		switch event.Key {
		case KeyOne, KeyTwo, KeyThree, KeyFour, KeyFive, KeySix, KeySeven, KeyEight, KeyNine:
			handleNumberKey(event, gameState)
		case KeyDelete:
			handleDeleteKey(event, gameState)
		case KeyArrowUp, KeyArrowDown, KeyArrowRight, KeyArrowLeft:
			handleArrowKey(event, gameState)
		case KeyD: // Debug
			if event.Modifiers.Ctrl {
				debug = !debug
			}
		case KeyR:
			if event.Modifiers.Ctrl {
				gameState.sudoku.reset()
				gameState.undoHistory = &undoHistory{}
			}
		case KeyY:
			if event.Modifiers.Ctrl {
				gameState.undoHistory.redo(gameState.sudoku)
			}
		case KeyZ:
			if event.Modifiers.Ctrl {
				gameState.undoHistory.undo(gameState.sudoku)
			}
		}
	}
}

func handleDoubleClick(event Event, gameState *gameState) {
	gameState.currentCell = event.Cell

	if !event.Modifiers.Ctrl {
		gameState.sudoku.unselectAllCells()
	}

	clickedCell := gameState.sudoku.at(event.Cell.Col, event.Cell.Row)
	clickedCell.selected = true

	gameState.sudoku.selectIf(func(candidate cellState) bool {
		switch {
		case clickedCell.isColored():
			return candidate.hasAllColors(*clickedCell)
		case !clickedCell.isEmpty():
			return candidate.Value == clickedCell.Value
		case clickedCell.hasCenterMarks():
			return candidate.isEmpty() && candidate.containsCenterMarksOf(*clickedCell)
		case clickedCell.hasCornerMarks():
			return candidate.isEmpty() && candidate.containsCornerMarksOf(*clickedCell)
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
