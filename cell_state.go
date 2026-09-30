package main

import (
	"encoding/json"
	"slices"
)

const emptyCell = 0

type cellState struct {
	Value       int `json:"value"`
	selected    bool
	Given       bool            `json:"given"`
	CornerMarks [cellCount]bool `json:"corner_marks"`
	CenterMarks [cellCount]bool `json:"center_marks"`
	ColorMask   bitMask         `json:"color_mask,omitempty"`
}

func (c *cellState) toggleCornerMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid corner index")
	}
	c.CornerMarks[index] = !c.CornerMarks[index]
}

func (c *cellState) toggleCenterMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid center index")
	}
	c.CenterMarks[index] = !c.CenterMarks[index]
}

func (c *cellState) toggleColor(index int) {
	c.ColorMask.toggle(index)
}

func (c *cellState) clearCornerMarks() {
	c.CornerMarks = [cellCount]bool{}
}

func (c *cellState) clearCenterMarks() {
	c.CenterMarks = [cellCount]bool{}
}

func (c *cellState) clearNumber() {
	if !c.Given {
		c.Value = emptyCell
	}
}

func (c *cellState) isEmpty() bool {
	return c.Value == emptyCell
}

func (c *cellState) hasCenterMarks() bool {
	return slices.Contains(c.CenterMarks[:], true)
}

func (c *cellState) hasCornerMarks() bool {
	return slices.Contains(c.CornerMarks[:], true)
}

func (c *cellState) containsCenterMarksOf(other cellState) bool {
	for i, marked := range other.CenterMarks {
		if marked && !c.CenterMarks[i] {
			return false
		}
	}
	return true
}

func (c *cellState) containsCornerMarksOf(other cellState) bool {
	for i, marked := range other.CornerMarks {
		if marked && !c.CornerMarks[i] {
			return false
		}
	}
	return true
}

func (c *cellState) isColored() bool {
	return c.ColorMask.count() > 0
}

func (c *cellState) colorCount() int {
	return c.ColorMask.count()
}

func (c *cellState) hasAllColors(other bitMask) bool {
	for i := range len(cellColors) {
		if other.has(i) && !c.ColorMask.has(i) {
			return false
		}
	}
	return true
}

func (c *cellState) colorIndexes() []int {
	return c.ColorMask.indexes()
}

type savedCellState struct {
	Value      int     `json:"value"`
	Given      bool    `json:"given,omitempty"`
	CornerMask uint16  `json:"corner_marks,omitempty"`
	CenterMask uint16  `json:"center_marks,omitempty"`
	ColorMask  bitMask `json:"color_mask,omitempty"`
}

func (c cellState) MarshalJSON() ([]byte, error) {
	return json.Marshal(savedCellState{
		Value:      c.Value,
		Given:      c.Given,
		CornerMask: encodeMarks(c.CornerMarks),
		CenterMask: encodeMarks(c.CenterMarks),
		ColorMask:  c.ColorMask,
	})
}

func (c *cellState) UnmarshalJSON(data []byte) error {
	var saved savedCellState

	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	c.Value = saved.Value
	c.Given = saved.Given
	c.CornerMarks = decodeMarks(saved.CornerMask)
	c.CenterMarks = decodeMarks(saved.CenterMask)
	c.ColorMask = saved.ColorMask

	return nil
}

func encodeMarks(marks [cellCount]bool) uint16 {
	var mask uint16

	for i, marked := range marks {
		if marked {
			mask |= (1 << i)
		}
	}

	return mask
}

func decodeMarks(mask uint16) [cellCount]bool {
	var marks [cellCount]bool

	for i := range cellCount {
		if (mask & (1 << i)) > 0 {
			marks[i] = true
		}
	}

	return marks
}
