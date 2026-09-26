package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EventType int

const (
	EventMousePressed EventType = iota
	EventMouseCellEntered
	EventMouseReleased
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
}

func (e Event) String() string {
	switch e.Type {
	case EventMousePressed, EventMouseCellEntered, EventMouseReleased:
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
	cell := Cell{Row: y, Col: x}

	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		i.mouse.Active = true
		i.mouse.VisitedCells = map[Cell]struct{}{
			cell: {},
		}
		events = append(events, Event{
			Type:      EventMousePressed,
			Cell:      cell,
			Modifiers: modifiers,
		})
	case rl.IsMouseButtonDown(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		if _, visited := i.mouse.VisitedCells[cell]; visited {
			break
		}
		if !isInsideSelectionArea(mousePos, x, y) {
			break
		}
		i.mouse.VisitedCells[cell] = struct{}{}
		events = append(events, Event{
			Type:      EventMouseCellEntered,
			Cell:      cell,
			Modifiers: modifiers,
		})
	case rl.IsMouseButtonReleased(rl.MouseButtonLeft):
		if !i.mouse.Active {
			panic("invalid input state, must be active")
		}
		i.mouse.Active = false
		events = append(events, Event{
			Type:      EventMouseReleased,
			Cell:      cell,
			Modifiers: modifiers,
		})
	}

	pressed := true
	keyEvent := Event{
		Type:      EventKeyPressed,
		Cell:      cell,
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
