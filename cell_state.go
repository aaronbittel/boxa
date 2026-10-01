package main

const emptyCell = 0

type cellState struct {
	Value       int `json:"value"`
	selected    bool
	Given       bool    `json:"given,omitempty"`
	CornerMarks bitMask `json:"corner_marks,omitempty"`
	CenterMarks bitMask `json:"center_marks,omitempty"`
	Colors      bitMask `json:"colors,omitempty"`
}

func (c *cellState) toggleCornerMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid corner index")
	}
	c.CornerMarks.toggle(index)
}

func (c *cellState) toggleCenterMark(index int) {
	if index < 0 || index >= cellCount {
		panic("invalid center index")
	}
	c.CenterMarks.toggle(index)
}

func (c *cellState) toggleColor(index int) {
	if c.Colors.has(index) {
		c.Colors.unset(index)
		return
	}

	if c.colorCount() >= len(cellRegionsOffsets) {
		return
	}

	c.Colors.set(index)
}

func (c *cellState) clearCornerMarks() {
	c.CornerMarks.clear()
}

func (c *cellState) clearCenterMarks() {
	c.CenterMarks.clear()
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
	return !c.CenterMarks.isEmpty()
}

func (c *cellState) hasCornerMarks() bool {
	return !c.CornerMarks.isEmpty()
}

func (c *cellState) containsCenterMarksOf(other cellState) bool {
	return c.CenterMarks.contains(other.CenterMarks)
}

func (c *cellState) containsCornerMarksOf(other cellState) bool {
	return c.CornerMarks.contains(other.CornerMarks)
}

func (c *cellState) isColored() bool {
	return !c.Colors.isEmpty()
}

func (c *cellState) colorCount() int {
	return c.Colors.count()
}

func (c *cellState) hasAllColors(other cellState) bool {
	return c.Colors.contains(other.Colors)
}

func (c *cellState) colorIndexes() []int {
	return c.Colors.indexes()
}

func (c *cellState) centerMarks() []int {
	indexes := c.CenterMarks.indexes()
	marks := make([]int, len(indexes))

	for i, index := range indexes {
		marks[i] = index + 1
	}

	return marks
}

func (c *cellState) cornerMarks() []int {
	indexes := c.CornerMarks.indexes()
	marks := make([]int, len(indexes))

	for i, index := range indexes {
		marks[i] = index + 1
	}

	return marks
}
