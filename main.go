package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aaronbittel/boxa/fetch"
	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/fonts/DejaVuSans.ttf
var fontBytes []byte

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

	rl.SetMouseCursor(rl.MouseCursorPointingHand)

	font := rl.LoadFontFromMemory(".ttf", fontBytes, int32(textFontSize), nil)
	defer rl.UnloadFont(font)

	var input Input

	start := time.Now()
	var duration time.Duration

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		events := input.poll()

		for _, event := range events {
			if event.Key == KeyV && event.Modifiers.Ctrl {
				newGameState, err := loadSudokuFromClipboard()
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: %s", err)
					continue
				}
				game = newGameState
				continue
			}

			if debug {
				fmt.Println("event", event, "selectionMode", game.selectionMode)
			}

			switch game.mode {
			case modePlaying:
				if err := game.handlePlayingEvent(event); err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: %s", err)
					continue
				}

				if game.sudoku.isSolved() {
					game.mode = modeSolved
					duration = game.elapsed + time.Since(start)
				}
			case modeSolving, modeSolved, modeUnsolveable: // do nothing
			default:
				panic(fmt.Sprintf("unhandled game mode %q", game.mode))
			}
		}

		if game.mode == modeSolving {
			game.updateSolver()
		}

		rl.BeginDrawing()
		rl.ClearBackground(backgroundColor)

		drawCellBackground(*game.sudoku)
		drawSelectedBorders(*game.sudoku)
		drawSudoku(*game.sudoku, font)

		drawGrid(borderColor(game.mode))
		if game.mode == modeSolved {
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
					return nil, fmt.Errorf("parse cell value %q: %w", data.Cells[y][x].Value, err)
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
