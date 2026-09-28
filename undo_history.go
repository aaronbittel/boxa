package main

type cellChange struct {
	before cellState
	after  cellState
}

type historyEntry map[Cell]cellChange

type undoHistory struct {
	entries []historyEntry
	cursor  int

	current historyEntry
}

func (u *undoHistory) undo(s *sudoku) {
	if !u.canUndo() {
		return
	}

	u.cursor--
	s.unselectAllCells()

	for cell, change := range u.entries[u.cursor] {
		change.before.selected = true
		s.set(cell, change.before)
	}
}

func (u *undoHistory) redo(s *sudoku) {
	if !u.canRedo() {
		return
	}

	s.unselectAllCells()

	for cell, change := range u.entries[u.cursor] {
		change.before.selected = true
		s.set(cell, change.after)
	}

	u.cursor++
}

func (u *undoHistory) add(entry historyEntry) {
	if len(entry) == 0 {
		return
	}

	u.entries = u.entries[:u.cursor]

	u.entries = append(u.entries, entry)
	u.cursor++
}

func (u *undoHistory) begin() {
	u.current = historyEntry{}
}

func (u *undoHistory) commit() {
	if len(u.current) == 0 {
		u.current = nil
		return
	}

	u.entries = u.entries[:u.cursor]
	u.entries = append(u.entries, u.current)
	u.cursor++

	u.current = nil
}

func (u *undoHistory) record(pos Cell, before, after cellState) {
	u.current[pos] = cellChange{
		before: before,
		after:  after,
	}
}

func (u *undoHistory) canUndo() bool {
	return u.cursor > 0
}

func (u *undoHistory) canRedo() bool {
	return u.cursor < len(u.entries)
}
