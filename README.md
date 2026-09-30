# Boxa

A small Sudoku solving pad for the desktop, inspired by
[SudokuPad](https://sudokupad.app).

**[Download the latest release](https://github.com/aaronbittel/boxa/releases/latest)**

Boxa takes a SudokuPad puzzle ID, downloads the puzzle, and lets you solve it locally.

## Usage

Start Boxa with a SudokuPad puzzle ID:

```bash
go run . -id T8qj67f8Mg
```

For example, the ID from:

```text
https://sudokupad.app/T8qj67f8Mg
```

Boxa downloads the initial puzzle state and populates the grid.

## Saving

Boxa automatically saves the current puzzle state when you close the application if the
Sudoku is not solved. Your progress is restored the next time you start Boxa with the
same puzzle ID.

## Controls

| Key                                  | Action                    |
| ------------------------------------ | ------------------------- |
| `1`–`9`                              | Enter a number            |
| `Backspace` / `Delete`               | Clear a cell              |
| `Arrow keys`                         | Move the selection        |
| `Shift` / `Ctrl` + `Arrow keys`      | Select cells while moving |
| `Shift` + `1`–`9`                    | Toggle corner marks       |
| `Ctrl` + `1`–`9`                     | Toggle center marks       |
| `Ctrl` + `Shift` + `1`–`5`           | Toggle cell colors        |
| `Ctrl` + `Delete`                    | Clear center marks        |
| `Shift` + `Delete`                   | Clear corner marks        |
| `Ctrl` + `Shift` + `Delete`          | Clear cell colors         |
| `Ctrl` + `Z`                         | Undo                      |
| `Ctrl` + `Y`                         | Redo                      |
| `Ctrl` + `R`                         | Reset the Sudoku          |

### Mouse

* Click a cell to select it.
* `Ctrl` + click to select multiple cells.
* Drag to select multiple cells.
* Double-click a cell to select other cells with the same number, marks, or colors.

## Build

Boxa is written in Go and uses [Raylib](https://www.raylib.com/).

```bash
go build -o boxa .
```

Then:

```bash
./boxa -id T8qj67f8Mg
```
