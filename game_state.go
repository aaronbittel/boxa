package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

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
