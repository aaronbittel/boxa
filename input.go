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
)

func (k Key) Value() int {
	if k < KeyOne || k > KeyNine {
		panic("key has no Sudoku value")
	}
	return int(k)
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
	Active        bool
	VisitedCells  map[Cell]struct{}
	lastCell      Cell
	lastClickTime time.Time
}

func (ms MouseState) isDoubleClick(currentCell Cell) bool {
	return currentCell == ms.lastCell && time.Since(ms.lastClickTime) < doubleClickTimeThreshold
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

func (i *Input) handle() []Event {
	events := []Event{}

	mousePos := rl.GetMousePosition()
	modifiers := Modifiers{
		Ctrl:  rl.IsKeyDown(rl.KeyLeftControl),
		Shift: rl.IsKeyDown(rl.KeyLeftShift),
	}

	x, y := positionToCellIdx(mousePos)
	currentCell := Cell{Row: y, Col: x}
	now := time.Now()
	defer func() {
		i.mouse.lastCell = currentCell
	}()

	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		i.mouse.Active = true
		i.mouse.VisitedCells = map[Cell]struct{}{
			currentCell: {},
		}
		if i.mouse.isDoubleClick(currentCell) {
			i.mouse.lastClickTime = time.Time{}
			events = append(events, Event{
				Type:      EventMouseDoubleClick,
				Cell:      currentCell,
				Modifiers: modifiers,
			})
		} else {
			i.mouse.lastClickTime = now
			events = append(events, Event{
				Type:      EventMousePressed,
				Cell:      currentCell,
				Modifiers: modifiers,
			})
		}
	case rl.IsMouseButtonDown(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		if _, visited := i.mouse.VisitedCells[currentCell]; visited {
			break
		}
		if !isInsideSelectionArea(mousePos, x, y) {
			break
		}
		i.mouse.VisitedCells[currentCell] = struct{}{}
		events = append(events, Event{
			Type:      EventMouseCellEntered,
			Cell:      currentCell,
			Modifiers: modifiers,
		})
	case rl.IsMouseButtonReleased(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		i.mouse.Active = false
		events = append(events, Event{
			Type:      EventMouseReleased,
			Cell:      currentCell,
			Modifiers: modifiers,
		})
	}

	pressed := true
	keyEvent := Event{
		Type:      EventKeyPressed,
		Cell:      currentCell,
		Modifiers: modifiers,
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
	default:
		pressed = false
	}

	if pressed {
		events = append(events, keyEvent)
	}

	return events
}
