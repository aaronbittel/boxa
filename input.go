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
	default:
		ok = false
	}
	return keyEvent, ok
}
