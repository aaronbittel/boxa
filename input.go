package main

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EventType int

const (
	EventMousePressed EventType = iota
	EventMouseCellEntered
	EventMouseReleased
	EventMouseDoubleClick
	EventKeyPressed
)

func (et EventType) String() string {
	switch et {
	case EventMousePressed:
		return "MousePressed"
	case EventMouseCellEntered:
		return "MouseCellEntered"
	case EventMouseReleased:
		return "MouseReleased"
	case EventMouseDoubleClick:
		return "MouseDoubleClick"
	case EventKeyPressed:
		return "KeyPressed"
	default:
		panic("update EventType.String()")
	}
}

type Key int

const (
	KeyOne Key = iota + 1
	KeyTwo
	KeyThree
	KeyFour
	KeyFive
	KeySix
	KeySeven
	KeyEight
	KeyNine

	KeyDelete

	KeyArrowUp
	KeyArrowDown
	KeyArrowRight
	KeyArrowLeft

	KeyR
	KeyY
	KeyZ

	KeyD // DebugKey
)

func (k Key) Value() int {
	if k < KeyOne || k > KeyNine {
		panic("key has no Sudoku value")
	}
	return int(k)
}

func (k Key) String() string {
	switch k {
	case KeyOne:
		return "KeyOne"
	case KeyTwo:
		return "KeyTwo"
	case KeyThree:
		return "KeyThree"
	case KeyFour:
		return "KeyFour"
	case KeyFive:
		return "KeyFive"
	case KeySix:
		return "KeySix"
	case KeySeven:
		return "KeySeven"
	case KeyEight:
		return "KeyEight"
	case KeyNine:
		return "KeyNine"
	case KeyDelete:
		return "KeyDelete"
	case KeyArrowUp:
		return "KeyArrowUp"
	case KeyArrowDown:
		return "KeyArrowDown"
	case KeyArrowRight:
		return "KeyArrowRight"
	case KeyArrowLeft:
		return "KeyArrowLeft"
	case KeyD:
		return "KeyD"
	case KeyR:
		return "KeyR"
	case KeyY:
		return "KeyY"
	case KeyZ:
		return "KeyZ"
	default:
		panic("new key was added")
	}
}

const doubleClickTimeThreshold = 500 * time.Millisecond

type Modifiers struct {
	Ctrl  bool
	Shift bool
}

type Cell struct {
	Row int
	Col int
}

func (c Cell) String() string {
	return fmt.Sprintf("x=%d, y=%d", c.Col, c.Row)
}

type Event struct {
	Type      EventType
	Cell      Cell
	Key       Key
	Modifiers Modifiers
}

type Input struct {
	mouse MouseState
}

type MouseState struct {
	Active       bool
	VisitedCells map[Cell]struct{}
	LastClick    struct {
		Active bool
		Cell   Cell
		When   time.Time
	}
}

func (ms MouseState) isDoubleClick(currentCell Cell) bool {
	return ms.LastClick.Active &&
		currentCell == ms.LastClick.Cell &&
		time.Since(ms.LastClick.When) < doubleClickTimeThreshold
}

func (e Event) String() string {
	switch e.Type {
	case EventMousePressed, EventMouseCellEntered, EventMouseReleased, EventMouseDoubleClick:
		return fmt.Sprintf("%s(%s)", e.Type, e.Cell)
	case EventKeyPressed:
		return fmt.Sprintf("%s(%s)", e.Type, e.Key)
	default:
		panic("update EventType.String()")
	}
}

func (i *Input) poll() []Event {
	events := []Event{}

	if mouseEvent, ok := i.mouseEvent(); ok {
		events = append(events, mouseEvent)
	}

	if keyEvent, ok := i.keyEvent(); ok {
		events = append(events, keyEvent)
	}

	return events
}

func (i *Input) mouseEvent() (Event, bool) {
	mousePos := rl.GetMousePosition()
	x, y := positionToCellIdx(mousePos)
	currentCell := Cell{Row: y, Col: x}

	modifiers := Modifiers{
		Ctrl:  rl.IsKeyDown(rl.KeyLeftControl),
		Shift: rl.IsKeyDown(rl.KeyLeftShift),
	}

	now := time.Now()

	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		i.mouse.Active = true
		i.mouse.VisitedCells = map[Cell]struct{}{
			currentCell: {},
		}
		if i.mouse.isDoubleClick(currentCell) {
			i.mouse.LastClick.Active = false
			return Event{
				Type:      EventMouseDoubleClick,
				Cell:      currentCell,
				Modifiers: modifiers,
			}, true
		} else {
			i.mouse.LastClick.Active = true
			i.mouse.LastClick.Cell = currentCell
			i.mouse.LastClick.When = now
			return Event{
				Type:      EventMousePressed,
				Cell:      currentCell,
				Modifiers: modifiers,
			}, true
		}
	case rl.IsMouseButtonDown(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		if _, visited := i.mouse.VisitedCells[currentCell]; visited {
			break
		}
		if !isInsideSelectionArea(rl.GetMousePosition(), x, y) {
			break
		}
		i.mouse.LastClick.Active = false
		i.mouse.VisitedCells[currentCell] = struct{}{}
		return Event{
			Type:      EventMouseCellEntered,
			Cell:      currentCell,
			Modifiers: modifiers,
		}, true
	case rl.IsMouseButtonReleased(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		i.mouse.Active = false
		return Event{
			Type:      EventMouseReleased,
			Cell:      currentCell,
			Modifiers: modifiers,
		}, true
	}

	return Event{}, false
}

func (*Input) keyEvent() (Event, bool) {
	mousePos := rl.GetMousePosition()
	x, y := positionToCellIdx(mousePos)
	currentCell := Cell{Row: y, Col: x}

	ok := true
	keyEvent := Event{
		Type: EventKeyPressed,
		Cell: currentCell,
		Modifiers: Modifiers{
			Ctrl:  rl.IsKeyDown(rl.KeyLeftControl),
			Shift: rl.IsKeyDown(rl.KeyLeftShift),
		},
	}

	switch {
	case rl.IsKeyPressed(rl.KeyOne), rl.IsKeyPressed(rl.KeyKp1):
		keyEvent.Key = KeyOne
	case rl.IsKeyPressed(rl.KeyTwo), rl.IsKeyPressed(rl.KeyKp2):
		keyEvent.Key = KeyTwo
	case rl.IsKeyPressed(rl.KeyThree), rl.IsKeyPressed(rl.KeyKp3):
		keyEvent.Key = KeyThree
	case rl.IsKeyPressed(rl.KeyFour), rl.IsKeyPressed(rl.KeyKp4):
		keyEvent.Key = KeyFour
	case rl.IsKeyPressed(rl.KeyFive), rl.IsKeyPressed(rl.KeyKp5):
		keyEvent.Key = KeyFive
	case rl.IsKeyPressed(rl.KeySix), rl.IsKeyPressed(rl.KeyKp6):
		keyEvent.Key = KeySix
	case rl.IsKeyPressed(rl.KeySeven), rl.IsKeyPressed(rl.KeyKp7):
		keyEvent.Key = KeySeven
	case rl.IsKeyPressed(rl.KeyEight), rl.IsKeyPressed(rl.KeyKp8):
		keyEvent.Key = KeyEight
	case rl.IsKeyPressed(rl.KeyNine), rl.IsKeyPressed(rl.KeyKp9):
		keyEvent.Key = KeyNine
	case rl.IsKeyPressed(rl.KeyBackspace), rl.IsKeyPressed(rl.KeyDelete):
		keyEvent.Key = KeyDelete
	case rl.IsKeyPressed(rl.KeyUp):
		keyEvent.Key = KeyArrowUp
	case rl.IsKeyPressed(rl.KeyDown):
		keyEvent.Key = KeyArrowDown
	case rl.IsKeyPressed(rl.KeyRight):
		keyEvent.Key = KeyArrowRight
	case rl.IsKeyPressed(rl.KeyLeft):
		keyEvent.Key = KeyArrowLeft
	case rl.IsKeyPressed(rl.KeyD):
		keyEvent.Key = KeyD
	case rl.IsKeyPressed(rl.KeyR):
		keyEvent.Key = KeyR
	case rl.IsKeyPressed(rl.KeyZ): // German keyboard: physical Z key is reported for Y
		keyEvent.Key = KeyY
	case rl.IsKeyPressed(rl.KeyY): // German keyboard: physical Y key is reported for Z
		keyEvent.Key = KeyZ
	default:
		ok = false
	}

	return keyEvent, ok
}

type selectionMode int

const (
	selectionUnset selectionMode = iota
	selectionSelect
	selectionDeselect
)

func (s selectionMode) String() string {
	switch s {
	case selectionUnset:
		return "SelectionUnset"
	case selectionSelect:
		return "SelectionSelect"
	case selectionDeselect:
		return "SelectionDeselect"
	default:
		panic("new SelectionMode variant was added")
	}
}

func handleEvent(event Event, gameState *gameState) {
	switch event.Type {
	case EventMousePressed:
		handleSingleClick(event, gameState)
	case EventMouseCellEntered:
		handleDragging(event, gameState)
	case EventMouseDoubleClick:
		handleDoubleClick(event, gameState)
	case EventMouseReleased:
		gameState.selectionMode = selectionUnset
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
				gameState.undoHistory = undoHistory{}
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
	switch gameState.selectionMode {
	case selectionSelect:
		gameState.sudoku.selectCell(event.Cell.Col, event.Cell.Row)
	case selectionDeselect:
		gameState.sudoku.deselectCell(event.Cell.Col, event.Cell.Row)
	}
}

func handleSingleClick(event Event, gameState *gameState) {
	if !event.Modifiers.Ctrl {
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
	var (
		selectedCount = 0
		selectedCell  Cell
	)
	for y := range cellCount {
		for x := range cellCount {
			if gameState.sudoku.at(x, y).selected {
				selectedCount++
				if selectedCount > 1 {
					return
				}
				selectedCell = Cell{
					Row: y,
					Col: x,
				}
			}
		}
	}
	if selectedCount == 0 {
		return
	}

	gameState.sudoku.deselectCell(selectedCell.Col, selectedCell.Row)

	switch event.Key {
	case KeyArrowUp:
		if selectedCell.Row == 0 {
			selectedCell.Row = cellCount - 1
		} else {
			selectedCell.Row -= 1
		}
	case KeyArrowDown:
		if selectedCell.Row == cellCount-1 {
			selectedCell.Row = 0
		} else {
			selectedCell.Row += 1
		}
	case KeyArrowLeft:
		if selectedCell.Col == 0 {
			selectedCell.Col = cellCount - 1
		} else {
			selectedCell.Col -= 1
		}
	case KeyArrowRight:
		if selectedCell.Col == cellCount-1 {
			selectedCell.Col = 0
		} else {
			selectedCell.Col += 1
		}
	default:
		panic("illegal key, expected arrow key")
	}

	gameState.sudoku.selectCell(selectedCell.Col, selectedCell.Row)
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

func positionToCellIdx(pos rl.Vector2) (x, y int) {
	return min(cellCount-1, int(pos.X/cellSize)), min(cellCount-1, int(pos.Y/cellSize))
}

func isInsideSelectionArea(pos rl.Vector2, x, y int) bool {
	r1 := rl.Rectangle{
		X:      float32(x) * cellSize,
		Y:      float32(y)*cellSize + selectionMargin,
		Width:  cellSize,
		Height: cellSize - 2*selectionMargin,
	}
	r2 := rl.Rectangle{
		X:      float32(x)*cellSize + selectionMargin,
		Y:      float32(y) * cellSize,
		Width:  cellSize - 2*selectionMargin,
		Height: cellSize,
	}

	return rl.CheckCollisionPointRec(pos, r1) || rl.CheckCollisionPointRec(pos, r2)
}
