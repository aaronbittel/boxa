package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/aaronbittel/boxa/fetch"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	cellSize  = 85
	cellCount = 9
	width     = cellSize * cellCount
	height    = cellSize * cellCount

	borderThickness    = 4.0
	highlightThickness = 9
	selectionMargin    = 15

	cellNumberFontSize = cellSize * 0.7
)

type SelectionMode int

const (
	SelectionUnset SelectionMode = iota
	SelectionSelect
	SelectionDeselect
)

var highlightColor = rl.NewColor(0x4C, 0xA4, 0xFF, 0xFF)

var pencilMarkCornerOffsets = [cellCount]rl.Vector2{
	{X: 0, Y: 0},
	{X: 2, Y: 0},
	{X: 0, Y: 2},
	{X: 2, Y: 2},
	{X: 1, Y: 0},
	{X: 1, Y: 2},
	{X: 0, Y: 1},
	{X: 2, Y: 1},
	{X: 1, Y: 1},
}

func main() {
	var sudoku sudoku

	if len(os.Args) >= 2 {
		data := fetch.FetchSudoku(os.Args[1])
		if len(data.Cells) != cellCount {
			panic("illegal cell count")
		}
		for y := range cellCount {
			for x := range cellCount {
				if data.Cells[y][x].Value != "" {
					n, err := strconv.Atoi(data.Cells[y][x].Value)
					if err != nil {
						log.Fatal(err)
					}
					sudoku[y][x].value = n
				}
			}
		}
	} else {
		sudoku = initFilledSudoku()
	}

	rl.InitWindow(width, height, "Boxa")
	defer rl.CloseWindow()

	var input Input
	selectionMode := SelectionUnset

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		events := input.handle()

		for _, event := range events {
			handleEvent(event, &sudoku, &selectionMode)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawSelectedBorders(sudoku)
		drawSudoku(sudoku)
		drawGrid()

		rl.EndDrawing()
	}
}

func handleEvent(event Event, sudoku *sudoku, selectionMode *SelectionMode) {
	switch event.Type {
	case EventMousePressed:
		handleSingleClick(event, sudoku, selectionMode)
	case EventMouseCellEntered:
		handleDragging(event, sudoku, selectionMode)
	case EventMouseReleased:
		*selectionMode = SelectionUnset
	case EventKeyPressed:
		switch event.Key {
		case KeyOne, KeyTwo, KeyThree, KeyFour, KeyFive, KeySix, KeySeven, KeyEight, KeyNine:
			handleNumberKey(event, sudoku)
		case KeyDelete:
			handleDeleteKey(event, sudoku)
		}
	}
}

func handleDragging(event Event, sudoku *sudoku, selectionMode *SelectionMode) {
	switch *selectionMode {
	case SelectionSelect:
		sudoku.selectCell(event.Cell.Col, event.Cell.Row)
	case SelectionDeselect:
		sudoku.deselectCell(event.Cell.Col, event.Cell.Row)
	}
}

func handleSingleClick(event Event, sudoku *sudoku, selectionMode *SelectionMode) {
	if sudoku.isSelected(event.Cell.Col, event.Cell.Row) {
		*selectionMode = SelectionDeselect
	} else {
		*selectionMode = SelectionSelect
	}
	if !event.Modifiers.Ctrl {
		sudoku.unselectAllCells()
	}
	sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
}

func handleDeleteKey(event Event, sudoku *sudoku) {
	switch {
	case event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			cell.clearCornerMarks()
		})
	case event.Modifiers.Ctrl:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			cell.clearCenterMarks()
		})
	default:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			cell.clearNumber()
		})
	}
}

func handleNumberKey(event Event, sudoku *sudoku) {
	value := event.Key.Value()
	index := value - 1
	switch {
	case event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			if sudoku.isEmpty(event.Cell.Col, event.Cell.Row) {
				cell.toggleCornerMark(index)
			}
		})
	case event.Modifiers.Ctrl:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			if sudoku.isEmpty(event.Cell.Col, event.Cell.Row) {
				cell.toggleCenterMark(index)
			}
		})
	default:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			cell.value = value
		})
	}
}

func drawGrid() {
	border := rl.Rectangle{Width: width, Height: height}
	rl.DrawRectangleLinesEx(border, borderThickness, rl.Black)

	for y := 1; y < cellCount; y++ {
		start := rl.Vector2{X: 0, Y: float32(y * cellSize)}
		end := rl.Vector2{X: width, Y: float32(y * cellSize)}

		t := float32(borderThickness / 2)
		if y%3 == 0 {
			t = borderThickness
		}
		rl.DrawLineEx(start, end, t, rl.Black)
	}

	for x := 1; x < cellCount; x++ {
		start := rl.Vector2{X: float32(x * cellSize), Y: 0}
		end := rl.Vector2{X: float32(x * cellSize), Y: height}

		t := float32(borderThickness / 2)
		if x%3 == 0 {
			t = borderThickness
		}
		rl.DrawLineEx(start, end, t, rl.Black)
	}

}

func drawSudoku(s sudoku) {
	for y := range cellCount {
		for x := range cellCount {
			if !s.isEmpty(x, y) {
				drawCellNumber(s, x, y)
			} else {
				drawCornerMarks(s, x, y)
				drawCenterMarks(s, x, y)
			}
		}
	}
}
func drawCellNumber(s sudoku, x, y int) {
	cellX := int32(x * cellSize)
	cellY := int32(y * cellSize)

	text := strconv.Itoa(s[y][x].value)
	fontSize := int32(math.Round(cellNumberFontSize))
	textWidth := rl.MeasureText(text, fontSize)
	rl.DrawText(text, cellX+(cellSize-textWidth)/2, cellY+14, fontSize, rl.Blue)
}

func drawCornerMarks(s sudoku, x, y int) {
	cellX := int32(x * cellSize)
	cellY := int32(y * cellSize)
	cornerSize := int32((cellSize - 2*highlightThickness) / 3)
	i := 0
	for j, c := range s[y][x].cornerMarks {
		num := j + 1
		if c {
			text := strconv.Itoa(num)
			markWidth := rl.MeasureText(text, cornerSize)

			offset := pencilMarkCornerOffsets[i]
			markX := cellX + highlightThickness + int32(offset.X)*cornerSize + (cornerSize-markWidth)/2
			markY := cellY + highlightThickness + int32(offset.Y)*cornerSize + 4

			rl.DrawText(text, int32(markX), int32(markY), cornerSize, rl.Blue)
			i++
		}
	}
}

func drawCenterMarks(s sudoku, x, y int) {
	var sb strings.Builder

	for i, marked := range s[y][x].centerMarks {
		if marked {
			fmt.Fprintf(&sb, "%d", i+1)
		}
	}

	if sb.Len() == 0 {
		return
	}

	text := sb.String()
	var padding int32 = 4

	fontSize := int32(math.Round(cellNumberFontSize / 2))
	textWidth := rl.MeasureText(text, fontSize)
	for textWidth+padding > cellSize {
		fontSize--
		textWidth = rl.MeasureText(text, fontSize)
	}

	cellX := int32(x * cellSize)
	cellY := int32(y * cellSize)

	rl.DrawText(text, cellX+(cellSize-textWidth)/2, cellY+(cellSize-fontSize)/2+2, fontSize, rl.Blue)
}

func drawSelectedBorders(s sudoku) {
	for y := range cellCount {
		for x := range cellCount {
			if !s[y][x].selected {
				continue
			}

			edgesCount := 0

			cellX := int32(x * cellSize)
			cellY := int32(y * cellSize)

			if y == 0 || !s[y-1][x].selected {
				rl.DrawRectangle(cellX, cellY, cellSize, highlightThickness, highlightColor)
				edgesCount++
			}
			if y == 8 || !s[y+1][x].selected {
				rl.DrawRectangle(cellX, cellY+cellSize-highlightThickness, cellSize, highlightThickness, highlightColor)
				edgesCount++
			}
			if x == 0 || !s[y][x-1].selected {
				rl.DrawRectangle(cellX, cellY, highlightThickness, cellSize, highlightColor)
				edgesCount++
			}
			if x == 8 || !s[y][x+1].selected {
				rl.DrawRectangle(cellX+cellSize-highlightThickness, cellY, highlightThickness, cellSize, highlightColor)
				edgesCount++
			}

			if y > 0 && x < cellCount-1 && s[y-1][x].selected && s[y][x+1].selected && !s[y-1][x+1].selected {
				center := rl.Vector2{X: float32(cellX + cellSize), Y: float32(cellY)}
				rl.DrawCircleSector(center, highlightThickness, 180.0, 90.0, 16, highlightColor)
			}
			if y < cellCount-1 && x < cellCount-1 && s[y][x+1].selected && s[y+1][x].selected && !s[y+1][x+1].selected {
				center := rl.Vector2{X: float32(cellX + cellSize), Y: float32(cellY + cellSize)}
				rl.DrawCircleSector(center, highlightThickness, 180.0, 270.0, 16, highlightColor)
			}
			if y < cellCount-1 && x > 0 && s[y+1][x].selected && s[y][x-1].selected && !s[y+1][x-1].selected {
				center := rl.Vector2{X: float32(cellX), Y: float32(cellY + cellSize)}
				rl.DrawCircleSector(center, highlightThickness, 270.0, 360.0, 16, highlightColor)
			}
			if y > 0 && x > 0 && s[y][x-1].selected && s[y-1][x].selected && !s[y-1][x-1].selected {
				center := rl.Vector2{X: float32(cellX), Y: float32(cellY)}
				rl.DrawCircleSector(center, highlightThickness, 0.0, 90.0, 16, highlightColor)
			}
		}
	}
}

func drawSelectedCell(y, x int) {
	rec := rl.Rectangle{
		X:      float32(x) * cellSize,
		Y:      float32(y) * cellSize,
		Width:  cellSize,
		Height: cellSize,
	}
	rl.DrawRectangleLinesEx(rec, borderThickness, rl.Blue)
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
