package main

import (
	"encoding/json"
	"fmt"
)

type cellChange struct {
	Before cellState `json:"before"`
	After  cellState `json:"after"`
}

type HistoryEntry map[Cell]cellChange

func (h HistoryEntry) MarshalJSON() ([]byte, error) {
	m := make(map[string]cellChange, len(h))

	for cell, change := range h {
		key := fmt.Sprintf("%d,%d", cell.Col, cell.Row)
		m[key] = change
	}

	return json.Marshal(m)
}

func (h *HistoryEntry) UnmarshalJSON(data []byte) error {
	m := make(map[string]cellChange)

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	result := make(HistoryEntry, len(m))

	for cellStr, change := range m {
		cellPos, err := parseCellFromStr(cellStr)
		if err != nil {
			return err
		}

		result[cellPos] = change
	}

	*h = result

	return nil
}

func parseCellFromStr(s string) (Cell, error) {
	var col, row int

	n, err := fmt.Sscanf(s, "%d,%d", &col, &row)
	if err != nil {
		return Cell{}, err
	}
	if n != 2 {
		return Cell{}, fmt.Errorf("invalid cell: %q", s)
	}

	return Cell{
		Col: col,
		Row: row,
	}, nil
}

type undoHistory struct {
	Entries []HistoryEntry `json:"entries,omitempty"`
	Cursor  int            `json:"cursor"`

	current HistoryEntry
}

func (u *undoHistory) undo(s *sudoku) {
	s.unselectAllCells()

	if !u.canUndo() {
		return
	}

	u.Cursor--

	for cell, change := range u.Entries[u.Cursor] {
		change.Before.selected = true
		s.set(cell, change.Before)
	}
}

func (u *undoHistory) redo(s *sudoku) {
	if !u.canRedo() {
		return
	}

	s.unselectAllCells()

	for cell, change := range u.Entries[u.Cursor] {
		change.Before.selected = true
		s.set(cell, change.After)
	}

	u.Cursor++
}

func (u *undoHistory) add(entry HistoryEntry) {
	if len(entry) == 0 {
		return
	}

	u.Entries = u.Entries[:u.Cursor]

	u.Entries = append(u.Entries, entry)
	u.Cursor++
}

func (u *undoHistory) begin() {
	u.current = HistoryEntry{}
}

func (u *undoHistory) commit() {
	if len(u.current) == 0 {
		u.current = nil
		return
	}

	u.Entries = u.Entries[:u.Cursor]
	u.Entries = append(u.Entries, u.current)
	u.Cursor++

	u.current = nil
}

func (u *undoHistory) record(pos Cell, before, after cellState) {
	u.current[pos] = cellChange{
		Before: before,
		After:  after,
	}
}

func (u *undoHistory) canUndo() bool {
	return u.Cursor > 0
}

func (u *undoHistory) canRedo() bool {
	return u.Cursor < len(u.Entries)
}
