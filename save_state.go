package main

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

const saveStateFilename = "./boxa-savestate.json"

type SaveState struct {
	ID          string        `json:"id"`
	Elapsed     time.Duration `json:"elapsed"`
	Sudoku      *sudoku       `json:"board_data"`
	UndoHistory *undoHistory  `json:"undo_history"`
}

func (s SaveState) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(s)
}

func (s SaveState) storeToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return s.encode(f)
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
