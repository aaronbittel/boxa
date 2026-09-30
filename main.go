package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
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

	borderThickness    = 5
	highlightThickness = 9
	selectionMargin    = 15

	cellNumberFontSize float32 = cellSize * 0.7
	textFontSize       float32 = 64.0

	saveStateFilename = "./boxa-savestate.json"
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

var cellColors = [6]color.RGBA{
	rl.NewColor(0xD0, 0xE0, 0xB7, 0xFF),
	rl.NewColor(0xF1, 0xB0, 0xF7, 0xFF),
	rl.NewColor(0xEF, 0xC0, 0x84, 0xFF),
	rl.NewColor(0xF9, 0x89, 0x87, 0xFF),
	rl.RayWhite,
	// rl.NewColor(0xFD, 0xF3, 0x8C, 0xFF),
	// rl.NewColor(0x8B, 0xC2, 0xF9, 0xFF),
}

func main() {
	var id string
	switch len(os.Args) {
	case 1: // generate sudoku
	case 2:
		id = os.Args[1]
	default:
		fmt.Fprintf(os.Stderr, "ERROR: too many arguments\nUsage: %s [id]\n", os.Args[0])
		os.Exit(1)
	}

	var (
		sudoku      = &sudoku{}
		undoHistory = &undoHistory{}
		elapsed     time.Duration
	)

	if id == "" {
		sudoku = initFilledSudoku()
	} else {
		saveState, err := loadSaveState(id)
		switch {
		case err == nil:
			sudoku = saveState.Sudoku
			undoHistory = &saveState.UndoHistory
			elapsed = saveState.Elapsed
		case errors.Is(err, os.ErrNotExist):
			sudoku, err = fetchSudoku(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: fetching sudoku failed: %s", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "ERROR: loading sudoku save state failed: %s\n", err)
			os.Exit(1)
		}
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
				handleEvent(event, sudoku, &selectionMode, undoHistory)
				if sudoku.isSolved() {
					sudokuIsSolved = true
					duration = elapsed + time.Since(start)
					break
				}
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawCellBackground(*sudoku)
		drawConflictingCells(*sudoku)
		drawSelectedBorders(*sudoku)
		drawSudoku(*sudoku, font)

		borderColor := rl.Black
		if sudokuIsSolved {
			borderColor = rl.Green
		}

		drawGrid(borderColor)
		if sudokuIsSolved {
			drawSudokuSolvedScreen(duration, font)
		}

		if debug {
			drawDebug()
		}

		rl.EndDrawing()
	}

	if id != "" {
		if !sudoku.isSolved() {
			saveState := &SaveState{
				ID:          id,
				Elapsed:     elapsed + time.Since(start),
				Sudoku:      sudoku,
				UndoHistory: *undoHistory,
			}
			if err := saveState.storeToFile(saveStateFilename); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s", err)
				os.Exit(1)
			}
			fmt.Printf("INFO: saved state %q\n", saveStateFilename)
		} else {
			if err := os.Remove(saveStateFilename); err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "ERROR: could not delete save state %q: %s\n", saveStateFilename, err)
				os.Exit(1)
			}

			fmt.Printf("INFO: deleted save state %q after successfully solving\n", saveStateFilename)
		}
	}
}

func fetchSudoku(id string) (*sudoku, error) {
	data, err := fetch.FetchSudoku(id)
	if err != nil {
		return nil, err
	}

	if len(data.Cells) != cellCount {
		return nil, fmt.Errorf("illegal cell count: %d", len(data.Cells))
	}

	var sudoku sudoku
	for y := range cellCount {
		for x := range cellCount {
			if data.Cells[y][x].Value != "" {
				n, err := strconv.Atoi(data.Cells[y][x].Value)
				if err != nil {
					return nil, err
				}
				sudoku[y][x].Value = n
				sudoku[y][x].Given = true
			}
		}
	}
	return &sudoku, nil
}

func handleEvent(event Event, sudoku *sudoku, selectionMode *SelectionMode, undoHistory *undoHistory) {
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
			handleNumberKey(event, sudoku, undoHistory)
		case KeyDelete:
			handleDeleteKey(event, sudoku, undoHistory)
		case KeyArrowUp, KeyArrowDown, KeyArrowRight, KeyArrowLeft:
			handleArrowKey(event, sudoku)
		case KeyD: // Debug
			if event.Modifiers.Ctrl {
				debug = !debug
			}
		case KeyY:
			if event.Modifiers.Ctrl {
				undoHistory.redo(sudoku)
			}
		case KeyZ:
			if event.Modifiers.Ctrl {
				undoHistory.undo(sudoku)
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
		case clickedCell.isColored():
			return candidate.hasAllColors(clickedCell.ColorMask)
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

func handleDeleteKey(event Event, sudoku *sudoku, undoHistory *undoHistory) {
	undoHistory.begin()
	defer undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			cell.ColorMask.clear()
		})
	case event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCornerMarks()
			undoHistory.record(pos, before, *cell)
		})
	case event.Modifiers.Ctrl:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearCenterMarks()
			undoHistory.record(pos, before, *cell)
		})
	default:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			before := *cell
			cell.clearNumber()
			undoHistory.record(pos, before, *cell)
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

func handleNumberKey(event Event, sudoku *sudoku, undoHistory *undoHistory) {
	value := event.Key.Value()
	index := value - 1

	undoHistory.begin()
	defer undoHistory.commit()

	switch {
	case event.Modifiers.Ctrl && event.Modifiers.Shift:
		switch value {
		case 1, 2, 3, 4, 5:
			selectedColorIndex := value - 1
			sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
				before := *cell
				cell.toggleColor(selectedColorIndex)
				undoHistory.record(pos, before, *cell)
			})
		}
	case event.Modifiers.Shift:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCornerMark(index)
				undoHistory.record(pos, before, *cell)
			}
		})
	case event.Modifiers.Ctrl:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if cell.isEmpty() {
				before := *cell
				cell.toggleCenterMark(index)
				undoHistory.record(pos, before, *cell)
			}
		})
	default:
		sudoku.forEachSelectedCell(func(pos Cell, cell *cellState) {
			if !cell.Given {
				before := *cell
				cell.Value = value
				undoHistory.record(pos, before, *cell)
			}
		})
	}
}

func drawGrid(borderColor color.RGBA) {
	border := rl.Rectangle{Width: width, Height: height}
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

type triangle [3]rl.Vector2
type colorRegion []triangle
type cellLayout []colorRegion

var cellRegionsOffsets = map[int]cellLayout{
	1: {
		{
			{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 0, Y: 1},
			},
			{
				{X: 1, Y: 0},
				{X: 0, Y: 1},
				{X: 1, Y: 1},
			},
		},
	},
	2: {
		{
			{
				{X: 0, Y: 0},
				{X: 0.7, Y: 0},
				{X: 0, Y: 1},
			},
			{
				{X: 0.7, Y: 0},
				{X: 0, Y: 1},
				{X: 0.3, Y: 1},
			},
		},
		{
			{
				{X: 1, Y: 0},
				{X: 0.3, Y: 1},
				{X: 1, Y: 1},
			},
			{
				{X: 0.3, Y: 1},
				{X: 0.7, Y: 0},
				{X: 1, Y: 0},
			},
		},
	},
	3: {
		{
			{
				{X: 0.0, Y: 0},
				{X: 0.75, Y: 0},
				{X: 0, Y: 0.55},
			},
			{
				{X: 0.75, Y: 0},
				{X: 0, Y: 0.55},
				{X: 0.5, Y: 0.5},
			},
		},
		{
			{
				{X: 0, Y: 0.55},
				{X: 0, Y: 1},
				{X: 0.5, Y: 0.5},
			},
			{
				{X: 0.5, Y: 0.5},
				{X: 0, Y: 1},
				{X: 0.85, Y: 1},
			},
		},
		{
			{
				{X: 0.75, Y: 0},
				{X: 1, Y: 0},
				{X: 0.5, Y: 0.5},
			},
			{
				{X: 0.5, Y: 0.5},
				{X: 0.85, Y: 1},
				{X: 1, Y: 1},
			},
			{
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0.5, Y: 0.5},
			},
		},
	},
	4: {
		{
			{
				{X: 0, Y: 0},
				{X: 0, Y: 0.25},
				{X: 0.5, Y: 0.5},
			},
			{
				{X: 0, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 0.75, Y: 0},
			},
		},
		{
			{
				{X: 0, Y: 0.25},
				{X: 0.5, Y: 0.5},
				{X: 0, Y: 1},
			},
			{
				{X: 0.5, Y: 0.5},
				{X: 0, Y: 1},
				{X: 0.3, Y: 1},
			},
		},
		{
			{
				{X: 0.3, Y: 1},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 1},
			},
			{
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 1},
				{X: 1, Y: 0.75},
			},
		},
		{
			{
				{X: 0.75, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 0},
			},
			{
				{X: 1, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 0.75},
			},
		},
	},
	5: {
		{
			{
				{X: 0, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 0, Y: 0.75},
			},
		},
		{
			{
				{X: 0, Y: 0.75},
				{X: 0.5, Y: 0.5},
				{X: 0, Y: 1},
			},
			{
				{X: 0, Y: 1},
				{X: 0.5, Y: 0.5},
				{X: 0.65, Y: 1},
			},
		},
		{
			{
				{X: 0.65, Y: 1},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 1},
			},
			{
				{X: 1, Y: 1},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 0.55},
			},
		},
		{
			{
				{X: 1, Y: 0.55},
				{X: 0.5, Y: 0.5},
				{X: 1, Y: 0},
			},
			{
				{X: 1, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 0.7, Y: 0},
			},
		},
		{
			{
				{X: 0, Y: 0},
				{X: 0.5, Y: 0.5},
				{X: 0.7, Y: 0},
			},
		},
	},
}

func drawCellBackground(s sudoku) {
	for y := range cellCount {
		for x := range cellCount {
			cell := s.at(x, y)

			if !cell.isColored() {
				continue
			}

			layout, ok := cellRegionsOffsets[cell.colorCount()]
			if !ok {
				continue
			}

			cellPos := rl.Vector2{
				X: float32(x * cellSize),
				Y: float32(y * cellSize),
			}

			indexes := cell.colorIndexes()

			if len(layout) != len(indexes) {
				panic("cell color layout and color index count mismatch")
			}

			for i, region := range layout {
				color := cellColors[indexes[i]]
				for _, triangle := range region {
					v1 := rl.Vector2Add(cellPos, rl.Vector2Scale(triangle[0], float32(cellSize)))
					v2 := rl.Vector2Add(cellPos, rl.Vector2Scale(triangle[1], float32(cellSize)))
					v3 := rl.Vector2Add(cellPos, rl.Vector2Scale(triangle[2], float32(cellSize)))
					drawTriangleCCW(v1, v2, v3, color)
				}
			}
		}
	}
}

func drawCellNumber(s sudoku, x, y int, font rl.Font) {
	cellX := float32(x * cellSize)
	cellY := float32(y * cellSize)

	text := strconv.Itoa(s.at(x, y).Value)
	textWidth := rl.MeasureTextEx(font, text, cellNumberFontSize, 0.0)
	pos := rl.Vector2{
		X: cellX + (cellSize-textWidth.X)/2,
		Y: cellY + 14.0,
	}

	color := defaultColor
	if s.at(x, y).Given {
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
	for j, marked := range s.at(x, y).CornerMarks {
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

	for i, marked := range s.at(x, y).CenterMarks {
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

	for i, marked := range s.at(x, y).CenterMarks {
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

func drawTriangleCCW(v1, v2, v3 rl.Vector2, color rl.Color) {
	cross := (v2.X-v1.X)*(v3.Y-v1.Y) - (v2.Y-v1.Y)*(v3.X-v1.X)

	if cross > 0 {
		rl.DrawTriangle(v1, v3, v2, color)
	} else {
		rl.DrawTriangle(v1, v2, v3, color)
	}
}

func loadSaveState(id string) (*SaveState, error) {
	f, err := os.Open(saveStateFilename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var saveState SaveState
	if err := json.NewDecoder(f).Decode(&saveState); err != nil {
		return nil, err
	}

	if saveState.ID != id {
		return nil, os.ErrNotExist
	}

	return &saveState, nil
}
