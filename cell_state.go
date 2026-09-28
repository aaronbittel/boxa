package main

import (
	"cmp"
	"image/color"
	"slices"
)

const emptyCell = 0

type cellState struct {
	value       int
	selected    bool
	given       bool
	cornerMarks [cellCount]bool
	centerMarks [cellCount]bool
	colors      []color.RGBA
}

func (c *cellState) toggleCornerMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid corner index")
	}
	c.cornerMarks[index] = !c.cornerMarks[index]
}

func (c *cellState) toggleCenterMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid center index")
	}
	c.centerMarks[index] = !c.centerMarks[index]
}

func (c *cellState) toggleColor(color color.RGBA) {
	i := slices.Index(c.colors, color)
	if i == -1 {
		c.colors = append(c.colors, color)
		c.sortColors()
	} else {
		c.colors = slices.Delete(c.colors, i, i+1)
	}
}

func (c *cellState) sortColors() {
	slices.SortFunc(c.colors, func(a, b color.RGBA) int {
		return cmp.Compare(
			slices.Index(regionColors[:], a),
			slices.Index(regionColors[:], b),
		)
	})
}

func (c *cellState) clearCornerMarks() {
	c.cornerMarks = [cellCount]bool{}
}

func (c *cellState) clearCenterMarks() {
	c.centerMarks = [cellCount]bool{}
}

func (c *cellState) clearNumber() {
	if !c.given {
		c.value = emptyCell
	}
}

func (c *cellState) isEmpty() bool {
	return c.value == emptyCell
}

func (c *cellState) hasCenterMarks() bool {
	return slices.Contains(c.centerMarks[:], true)
}

func (c *cellState) hasCornerMarks() bool {
	return slices.Contains(c.cornerMarks[:], true)
}

func (c *cellState) containsCenterMarksOf(other cellState) bool {
	for i, marked := range other.centerMarks {
		if marked && !c.centerMarks[i] {
			return false
		}
	}
	return true
}

func (c *cellState) containsCornerMarksOf(other cellState) bool {
	for i, marked := range other.cornerMarks {
		if marked && !c.cornerMarks[i] {
			return false
		}
	}
	return true
}
