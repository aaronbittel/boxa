package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aaronbittel/boxa/fetch"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var debug = false

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

	rl.InitWindow(windowWidth, windowHeight, "Boxa")
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
