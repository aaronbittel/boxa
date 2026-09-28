package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aaronbittel/boxa/fetch"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	cellSize  = 85
	cellCount = 9
	width     = cellSize * cellCount
	height    = cellSize * cellCount

	borderThickness    = 4
	highlightThickness = 9
	selectionMargin    = 15

	cellNumberFontSize float32 = cellSize * 0.7
	textFontSize       float32 = 64.0
)

type SelectionMode int

const (
	SelectionUnset SelectionMode = iota
	SelectionSelect
	SelectionDeselect
)

var (
	defaultColor            = rl.Blue
	highlightColor          = rl.NewColor(0x4C, 0xA4, 0xFF, 0xFF)
	conflictCellColor       = rl.NewColor(0xD9, 0x9C, 0x9C, 0xFF)
	conflictPencilMarkColor = rl.NewColor(0xBB, 0x4E, 0x4E, 0xFF)
	conflictBorderColor     = rl.NewColor(0x80, 0x6F, 0x9C, 0xFF)
)

var debug = false

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
					sudoku[y][x].given = true
				}
			}
		}
	} else {
		sudoku = initFilledSudoku()
	}

	rl.InitWindow(width, height, "Boxa")
	defer rl.CloseWindow()

	font := rl.LoadFontEx("./fonts/DejaVuSans.ttf", int32(textFontSize), nil, 0)
	defer rl.UnloadFont(font)

	var input Input

	sudokuIsSolved := false
	start := time.Now()
	var duration time.Duration
	selectionMode := SelectionUnset

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if !sudokuIsSolved {
			events := input.poll()

			for _, event := range events {
				if debug {
					fmt.Println("event", event, "selectionMode", selectionMode)
				}
				handleEvent(event, &sudoku, &selectionMode)
				if sudoku.isSolved() {
					sudokuIsSolved = true
					duration = time.Since(start)
					break
				}
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawConflictingCells(sudoku)
		drawSelectedBorders(sudoku)
		drawSudoku(sudoku, font)
		drawGrid(sudokuIsSolved)
		if sudokuIsSolved {
			drawSudokuSolvedScreen(duration, font)
		}

		if debug {
			drawDebug()
		}

		rl.EndDrawing()
	}
}

func handleEvent(event Event, sudoku *sudoku, selectionMode *SelectionMode) {
	switch event.Type {
	case EventMousePressed:
		handleSingleClick(event, sudoku, selectionMode)
	case EventMouseCellEntered:
		handleDragging(event, sudoku, selectionMode)
	case EventMouseDoubleClick:
		handleDoubleClick(event, sudoku)
	case EventMouseReleased:
		*selectionMode = SelectionUnset
	case EventKeyPressed:
		switch event.Key {
		case KeyOne, KeyTwo, KeyThree, KeyFour, KeyFive, KeySix, KeySeven, KeyEight, KeyNine:
			handleNumberKey(event, sudoku)
		case KeyDelete:
			handleDeleteKey(event, sudoku)
		case KeyArrowUp, KeyArrowDown, KeyArrowRight, KeyArrowLeft:
			handleArrowKey(event, sudoku)
		case KeyD: // Debug
			if event.Modifiers.Ctrl {
				debug = !debug
			}
		}
	}
}

func handleDoubleClick(event Event, sudoku *sudoku) {
	if !event.Modifiers.Ctrl {
		sudoku.unselectAllCells()
	}

	clickedCell := sudoku.at(event.Cell.Col, event.Cell.Row)
	clickedCell.selected = true

	sudoku.selectIf(func(candidate cellState) bool {
		switch {
		case !clickedCell.isEmpty():
			return candidate.value == clickedCell.value
		case clickedCell.hasCenterMarks():
			return candidate.isEmpty() && candidate.containsCenterMarksOf(*clickedCell)
		case clickedCell.hasCornerMarks():
			return candidate.isEmpty() && candidate.containsCornerMarksOf(*clickedCell)
		}
		return false
	})
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
	if !event.Modifiers.Ctrl {
		sudoku.unselectAllCells()
		sudoku.toggleSelection(event.Cell.Col, event.Cell.Row)
		*selectionMode = SelectionSelect
		return
	}

	if sudoku.isSelected(event.Cell.Col, event.Cell.Row) {
		*selectionMode = SelectionDeselect
	} else {
		*selectionMode = SelectionSelect
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

func handleArrowKey(event Event, sudoku *sudoku) {
	var (
		selectedCount = 0
		selectedCell  Cell
	)
	for y := range cellCount {
		for x := range cellCount {
			if sudoku.at(x, y).selected {
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

	sudoku.deselectCell(selectedCell.Col, selectedCell.Row)

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

	sudoku.selectCell(selectedCell.Col, selectedCell.Row)
}

func handleNumberKey(event Event, sudoku *sudoku) {
	value := event.Key.Value()
	index := value - 1
	switch {
	case event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			if sudoku.at(event.Cell.Col, event.Cell.Row).isEmpty() {
				cell.toggleCornerMark(index)
			}
		})
	case event.Modifiers.Ctrl:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			if sudoku.at(event.Cell.Col, event.Cell.Row).isEmpty() {
				cell.toggleCenterMark(index)
			}
		})
	default:
		sudoku.forEachSelectedCell(func(cell *cellState) {
			cell.value = value
		})
	}
}

func drawGrid(solved bool) {
	border := rl.Rectangle{Width: width, Height: height}
	borderColor := rl.Black
	if solved {
		borderColor = rl.Green
	}
	rl.DrawRectangleLinesEx(border, borderThickness, borderColor)

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

func drawSudokuSolvedScreen(duration time.Duration, font rl.Font) {
	const (
		spacing  = 2.0
		bgWidth  = width * 0.5
		bgHeight = height * 0.4
	)

	bg := rl.Rectangle{
		X:      (width - bgWidth) / 2,
		Y:      (height - bgHeight) / 2,
		Width:  bgWidth,
		Height: bgHeight,
	}

	congratsText := "Congrats!"
	congratsTextSize := rl.MeasureTextEx(font, congratsText, textFontSize, spacing)
	congratsTextPos := rl.Vector2{
		X: bg.X + (bg.Width-congratsTextSize.X)/2,
		Y: bg.Y + bg.Height*0.2,
	}

	timeText := fmt.Sprintf("Time: %s", formatDuration(duration))
	timeTextSize := rl.MeasureTextEx(font, timeText, textFontSize, spacing)
	timeTextPos := rl.Vector2{
		X: bg.X + (bg.Width-timeTextSize.X)/2,
		Y: bg.Y + bg.Height*0.5,
	}

	rl.DrawRectangleRounded(bg, 0.4, 32, rl.NewColor(0xE8, 0xE6, 0xE1, 0xFF))
	rl.DrawTextEx(font, congratsText, congratsTextPos, textFontSize, spacing, rl.Green)
	rl.DrawTextEx(font, timeText, timeTextPos, textFontSize, spacing, rl.Green)
}

func drawSudoku(s sudoku, font rl.Font) {
	for y := range cellCount {
		for x := range cellCount {
			if !s.at(x, y).isEmpty() {
				drawCellNumber(s, x, y, font)
			} else {
				drawCornerMarks(s, x, y, font)
				drawCenterMarks(s, x, y, font)
			}
		}
	}
}
func drawCellNumber(s sudoku, x, y int, font rl.Font) {
	cellX := float32(x * cellSize)
	cellY := float32(y * cellSize)

	text := strconv.Itoa(s.at(x, y).value)
	textWidth := rl.MeasureTextEx(font, text, cellNumberFontSize, 0.0)
	pos := rl.Vector2{
		X: cellX + (cellSize-textWidth.X)/2,
		Y: cellY + 14.0,
	}

	color := defaultColor
	if s.at(x, y).given {
		color = rl.Black
	}
	rl.DrawTextEx(font, text, pos, cellNumberFontSize, 0.0, color)
}

func drawCornerMarks(s sudoku, x, y int, font rl.Font) {
	if !s.at(x, y).hasCornerMarks() {
		return
	}

	cellX := float32(x * cellSize)
	cellY := float32(y * cellSize)
	cornerSize := float32((cellSize - 2*highlightThickness)) / 3
	i := 0
	for j, marked := range s.at(x, y).cornerMarks {
		num := j + 1
		if marked {
			text := strconv.Itoa(num)
			markSize := rl.MeasureTextEx(font, text, cornerSize, 0.0)

			offset := pencilMarkCornerOffsets[i]
			markX := cellX + highlightThickness + offset.X*cornerSize + (cornerSize-markSize.X)/2
			markY := cellY + highlightThickness + offset.Y*cornerSize + 4

			pos := rl.Vector2{
				X: markX,
				Y: markY,
			}

			color := defaultColor
			if s.hasConflictFor(x, y, num) {
				color = conflictPencilMarkColor
			}
			rl.DrawTextEx(font, text, pos, cornerSize, 0.0, color)
			i++
		}
	}
}

func drawCenterMarks(s sudoku, x, y int, font rl.Font) {
	if !s.at(x, y).hasCenterMarks() {
		return
	}

	var sb strings.Builder

	for i, marked := range s.at(x, y).centerMarks {
		if marked {
			fmt.Fprintf(&sb, "%d", i+1)
		}
	}

	text := sb.String()
	var padding float32 = 4.0

	fontSize := cellNumberFontSize / 2
	textSize := rl.MeasureTextEx(font, text, fontSize, 0.0)
	for textSize.X+padding > cellSize {
		fontSize--
		textSize = rl.MeasureTextEx(font, text, fontSize, 0.0)
	}

	pos := rl.Vector2{
		X: float32(x*cellSize) + float32((cellSize)-textSize.X)/2,
		Y: float32(y*cellSize) + float32((cellSize-fontSize))/2 + 2,
	}

	for i, marked := range s.at(x, y).centerMarks {
		if !marked {
			continue
		}

		num := i + 1
		color := defaultColor
		if s.hasConflictFor(x, y, num) {
			color = conflictPencilMarkColor
		}
		rl.DrawTextEx(font, strconv.Itoa(num), pos, fontSize, 0.0, color)
		pos.X += textSize.X / float32(len(text))
	}
}

func drawSelectedBorders(s sudoku) {
	for y := range cellCount {
		for x := range cellCount {
			if !s.at(x, y).selected {
				continue
			}

			color := highlightColor
			if s.hasCellConflict(x, y) {
				color = conflictBorderColor
			}

			edgesCount := 0

			cellX := int32(x * cellSize)
			cellY := int32(y * cellSize)

			if y == 0 || !s.at(x, y-1).selected {
				rl.DrawRectangle(cellX, cellY, cellSize, highlightThickness, color)
				edgesCount++
			}
			if y == 8 || !s.at(x, y+1).selected {
				rl.DrawRectangle(cellX, cellY+cellSize-highlightThickness, cellSize, highlightThickness, color)
				edgesCount++
			}
			if x == 0 || !s.at(x-1, y).selected {
				rl.DrawRectangle(cellX, cellY, highlightThickness, cellSize, color)
				edgesCount++
			}
			if x == 8 || !s.at(x+1, y).selected {
				rl.DrawRectangle(cellX+cellSize-highlightThickness, cellY, highlightThickness, cellSize, color)
				edgesCount++
			}

			if y > 0 && x < cellCount-1 && s.at(x, y-1).selected && s.at(x+1, y).selected && !s.at(x+1, y-1).selected {
				center := rl.Vector2{X: float32(cellX + cellSize), Y: float32(cellY)}
				rl.DrawCircleSector(center, highlightThickness, 180.0, 90.0, 16, color)
			}
			if y < cellCount-1 && x < cellCount-1 && s.at(x+1, y).selected && s.at(x, y+1).selected && !s.at(x+1, y+1).selected {
				center := rl.Vector2{X: float32(cellX + cellSize), Y: float32(cellY + cellSize)}
				rl.DrawCircleSector(center, highlightThickness, 180.0, 270.0, 16, color)
			}
			if y < cellCount-1 && x > 0 && s.at(x, y+1).selected && s.at(x-1, y).selected && !s.at(x-1, y+1).selected {
				center := rl.Vector2{X: float32(cellX), Y: float32(cellY + cellSize)}
				rl.DrawCircleSector(center, highlightThickness, 270.0, 360.0, 16, color)
			}
			if y > 0 && x > 0 && s.at(x-1, y).selected && s.at(x, y-1).selected && !s.at(x-1, y-1).selected {
				center := rl.Vector2{X: float32(cellX), Y: float32(cellY)}
				rl.DrawCircleSector(center, highlightThickness, 0.0, 90.0, 16, color)
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
	rl.DrawRectangleLinesEx(rec, borderThickness, defaultColor)
}

func drawConflictingCells(s sudoku) {
	for y := range cellCount {
		for x := range cellCount {
			if s.hasCellConflict(x, y) {
				cellX := int32(x * cellSize)
				cellY := int32(y * cellSize)
				rl.DrawRectangle(cellX, cellY, cellSize, cellSize, conflictCellColor)
			}
		}
	}
}

func drawDebug() {
	rl.DrawText("debug", 10, 10, 32, rl.LightGray)
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

func (s SelectionMode) String() string {
	switch s {
	case SelectionUnset:
		return "SelectionUnset"
	case SelectionSelect:
		return "SelectionSelect"
	case SelectionDeselect:
		return "SelectionDeselect"
	default:
		panic("new SelectionMode variant was added")
	}
}

func formatDuration(d time.Duration) string {
	totalSeconds := int(d / time.Second)
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
