package main

import "testing"

func TestMarksEncoding(t *testing.T) {
	tests := []struct {
		name  string
		marks [cellCount]bool
		want  uint16
	}{
		{
			name:  "zero",
			marks: [9]bool{},
			want:  0,
		},
		{
			name:  "all",
			marks: [9]bool{true, true, true, true, true, true, true, true, true},
			want:  0b111111111,
		},
		{
			name:  "1,3,5,7,9",
			marks: [9]bool{true, false, true, false, true, false, true, false, true},
			want:  0b101010101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mask := encodeMarks(tt.marks)
			if mask != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, mask)
			}
			got := decodeMarks(mask)
			if tt.marks != got {
				t.Errorf("expected %v, got %v", tt.marks, got)
			}
		})
	}
}
