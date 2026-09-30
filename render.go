package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	cellSize     = 85
	windowWidth  = cellSize * cellCount
	windowHeight = cellSize * cellCount

	borderThickness    = 5
	highlightThickness = 9
	selectionMargin    = 15

	cellNumberFontSize float32 = cellSize * 0.7
	textFontSize       float32 = 64.0
)

var (
	defaultColor            = rl.Blue
	highlightColor          = rl.NewColor(0x4C, 0xA4, 0xFF, 0xFF)
	conflictCellColor       = rl.NewColor(0xD9, 0x9C, 0x9C, 0xFF)
	conflictPencilMarkColor = rl.NewColor(0xBB, 0x4E, 0x4E, 0xFF)
	conflictBorderColor     = rl.NewColor(0x80, 0x6F, 0x9C, 0xFF)
)

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

func drawGrid(borderColor color.RGBA) {
	border := rl.Rectangle{Width: windowWidth, Height: windowHeight}
	rl.DrawRectangleLinesEx(border, borderThickness, borderColor)

	for y := 1; y < cellCount; y++ {
		start := rl.Vector2{X: 0, Y: float32(y * cellSize)}
		end := rl.Vector2{X: windowWidth, Y: float32(y * cellSize)}

		t := float32(borderThickness / 2)
		if y%3 == 0 {
			t = borderThickness
		}
		rl.DrawLineEx(start, end, t, rl.Black)
	}

	for x := 1; x < cellCount; x++ {
		start := rl.Vector2{X: float32(x * cellSize), Y: 0}
		end := rl.Vector2{X: float32(x * cellSize), Y: windowHeight}

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
		bgWidth  = windowWidth * 0.5
		bgHeight = windowHeight * 0.4
	)

	bg := rl.Rectangle{
		X:      (windowWidth - bgWidth) / 2,
		Y:      (windowHeight - bgHeight) / 2,
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
	for _, num := range s.at(x, y).cornerMarks() {
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

func drawCenterMarks(s sudoku, x, y int, font rl.Font) {
	if !s.at(x, y).hasCenterMarks() {
		return
	}

	var sb strings.Builder

	for _, num := range s.at(x, y).centerMarks() {
		fmt.Fprintf(&sb, "%d", num)
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

	for _, num := range s.at(x, y).centerMarks() {
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
