package main

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type SaveState struct {
	ID          string        `json:"id"`
	Elapsed     time.Duration `json:"elapsed"`
	Sudoku      *sudoku       `json:"board_data"`
	UndoHistory undoHistory   `json:"undo_history"`
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
