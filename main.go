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

//go:embed fonts/DejaVuSans.ttf
var fontBytes []byte

type gameMode int

const (
	modePlaying gameMode = iota
	modeSolving
	modeSolved
	modeUnsolveable
)

func (m gameMode) String() string {
	switch m {
	case modePlaying:
		return "Playing"
	case modeSolving:
		return "Solving"
	case modeSolved:
		return "Solved"
	case modeUnsolveable:
		return "Unsolveable"
	default:
		panic(fmt.Sprintf("unhandled game mode %d", m))
	}
}

type gameState struct {
	id            string
	sudoku        *sudoku
	undoHistory   *undoHistory
	elapsed       time.Duration
	currentCell   Cell
	selectionMode selectionMode
	mode          gameMode
	solver        *solver
}

func newGameState(id string, sudoku *sudoku) *gameState {
	return &gameState{
		id:          id,
		sudoku:      sudoku,
		undoHistory: new(undoHistory),
		solver:      newSolver(),
	}
}

func gameStateFromSaveState(save SaveState) *gameState {
	return &gameState{
		id:          save.ID,
		sudoku:      save.Sudoku,
		undoHistory: save.UndoHistory,
		elapsed:     save.Elapsed,
		solver:      newSolver(),
	}
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

func (g *gameState) updateSolver() {
	if !g.solver.started {
		g.solver.started = true
		go func() {
			g.solver.done <- sudokuSolveBacktracking(*g.sudoku, Cell{}, g.solver.changes)
		}()
	}
	select {
	case change, ok := <-g.solver.changes:
		if !ok {
			g.solver.changes = nil
			break
		}
		g.sudoku.deselectAllCells()
		g.sudoku.set(change.pos, cellState{
			selected: change.selected,
			Value:    change.value,
		})
	case solved := <-g.solver.done:
		if solved {
			g.mode = modeSolved
		} else {
			g.mode = modeUnsolveable
		}
		g.solver.done = nil
	default: // continue
	}
}

func loadGameState(id string) (*gameState, error) {
	if id == "" {
		return newGameState(id, initFilledSudoku()), nil
	}

	save, err := loadSaveState(id)
	switch {
	case err == nil:
		if save.ID == id {
			return gameStateFromSaveState(*save), nil
		}

	case !errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("reading save state: %w", err)
	}

	sudoku, err := fetchSudoku(id)
	if err != nil {
		return nil, err
	}

	return newGameState(id, sudoku), nil
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
