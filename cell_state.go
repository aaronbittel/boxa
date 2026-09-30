package main

import (
	"cmp"
	"encoding/json"
	"image/color"
	"slices"
)

const emptyCell = 0

type cellState struct {
	Value       int `json:"value"`
	selected    bool
	Given       bool            `json:"given"`
	CornerMarks [cellCount]bool `json:"corner_marks"`
	CenterMarks [cellCount]bool `json:"center_marks"`
	Colors      []color.RGBA    `json:"colors,omitempty"`
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

func (c *cellState) toggleColor(color color.RGBA) {
	i := slices.Index(c.Colors, color)
	if i == -1 {
		c.Colors = append(c.Colors, color)
		c.sortColors()
	} else {
		c.Colors = slices.Delete(c.Colors, i, i+1)
	}
}

func (c *cellState) sortColors() {
	slices.SortFunc(c.Colors, func(a, b color.RGBA) int {
		return cmp.Compare(
			slices.Index(cellColors[:], a),
			slices.Index(cellColors[:], b),
		)
	})
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
	return len(c.Colors) > 0
}

func (c *cellState) hasAllColors(colors []color.RGBA) bool {
	if len(c.Colors) < len(colors) {
		return false
	}

	for _, color := range colors {
		if !slices.Contains(c.Colors, color) {
			return false
		}
	}

	return true
}

type savedCellState struct {
	Value      int          `json:"value"`
	Given      bool         `json:"given,omitempty"`
	CornerMask uint16       `json:"corner_marks,omitempty"`
	CenterMask uint16       `json:"center_marks,omitempty"`
	Colors     []color.RGBA `json:"colors,omitempty"`
}

func (c cellState) MarshalJSON() ([]byte, error) {
	return json.Marshal(savedCellState{
		Value:      c.Value,
		Given:      c.Given,
		CornerMask: encodeMarks(c.CornerMarks),
		CenterMask: encodeMarks(c.CenterMarks),
		Colors:     c.Colors,
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
	c.Colors = saved.Colors

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
