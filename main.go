package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aaronbittel/boxa/fetch"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type gameState struct {
	id             string
	sudoku         *sudoku
	undoHistory    undoHistory
	elapsed        time.Duration
	selectionMode  selectionMode
	sudokuIsSolved bool
}

var debug = false

func main() {
	id := flag.String("id", "", "id of sudoku")
	flag.Parse()

	game, err := loadGameState(*id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: loading game state: %s\n", err)
		os.Exit(1)
	}

	rl.InitWindow(windowWidth, windowHeight, "Boxa")
	defer rl.CloseWindow()

	font := rl.LoadFontEx("./fonts/DejaVuSans.ttf", int32(textFontSize), nil, 0)
	defer rl.UnloadFont(font)

	var input Input

	start := time.Now()
	var duration time.Duration

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if !game.sudokuIsSolved {
			events := input.poll()

			for _, event := range events {
				if debug {
					fmt.Println("event", event, "selectionMode", game.selectionMode)
				}
				handleEvent(event, game)
				if game.sudoku.isSolved() {
					game.sudokuIsSolved = true
					duration = game.elapsed + time.Since(start)
					break
				}
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawCellBackground(*game.sudoku)
		drawConflictingCells(*game.sudoku)
		drawSelectedBorders(*game.sudoku)
		drawSudoku(*game.sudoku, font)

		borderColor := rl.Black
		if game.sudokuIsSolved {
			borderColor = rl.Green
		}

		drawGrid(borderColor)
		if game.sudokuIsSolved {
			drawSudokuSolvedScreen(duration, font)
		}

		if debug {
			drawDebug()
		}

		rl.EndDrawing()
	}

	if game.id == "" {
		os.Exit(0)
	}

	if game.sudoku.isSolved() {
		if err := deleteSaveState(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not delete save state %q: %s", saveStateFilename, err)
			os.Exit(1)
		}

		fmt.Printf("INFO: deleted save state %q after successfully solving\n", saveStateFilename)
		os.Exit(0)
	}

	saveState := &SaveState{
		ID:          game.id,
		Elapsed:     game.elapsed + time.Since(start),
		Sudoku:      game.sudoku,
		UndoHistory: game.undoHistory,
	}
	if err := saveState.storeToFile(saveStateFilename); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s", err)
		os.Exit(1)
	}
	fmt.Printf("INFO: saved state %q\n", saveStateFilename)
}

func loadGameState(id string) (*gameState, error) {
	if id == "" {
		return &gameState{
			sudoku: initFilledSudoku(),
		}, nil
	}

	saveState, err := loadSaveState(id)
	switch {
	case err == nil:
		if saveState.ID == id {
			return &gameState{
				id:          id,
				sudoku:      saveState.Sudoku,
				undoHistory: saveState.UndoHistory,
				elapsed:     saveState.Elapsed,
			}, nil
		}

	case !errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("reading save state: %w", err)
	}

	sudoku, err := fetchSudoku(id)
	if err != nil {
		return nil, fmt.Errorf("fetching sudoku with id %s: %w", id, err)
	}

	return &gameState{
		id:     id,
		sudoku: sudoku,
	}, nil
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

func deleteSaveState() error {
	if err := os.Remove(saveStateFilename); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil

}
