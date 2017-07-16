package numbers

import (
	"testing"
)

// test examples
var storeToSliceExamples = []struct {
	input    map[int]struct{}
	expected []int
}{
	{
		input:    map[int]struct{}{3: {}, 19: {}, 18: {}, 7: {}, 8: {}, 25: {}, 29: {}, 12: {}, 14: {}, 10: {}, 11: {}, 4: {}, 24: {}, 5: {}, 26: {}, 20: {}, 28: {}, 22: {}, 17: {}},
		expected: []int{3, 4, 5, 7, 8, 10, 11, 12, 14, 17, 18, 19, 20, 22, 24, 25, 26, 28, 29},
	},
	{
		input:    map[int]struct{}{520: {}, 528: {}, 512: {}, 514: {}, 503: {}, 504: {}, 517: {}, 508: {}, 524: {}, 526: {}, 510: {}, 522: {}, 525: {}, 529: {}, 505: {}, 518: {}, 507: {}, 519: {}, 511: {}},
		expected: []int{503, 504, 505, 507, 508, 510, 511, 512, 514, 517, 518, 519, 520, 522, 524, 525, 526, 528, 529},
	},
	{
		input:    map[int]struct{}{1: {}},
		expected: []int{1},
	},
	{
		input:    nil,
		expected: []int{},
	},
}

func compareSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestStoreToKeySortedSlice(t *testing.T) {
	for _, example := range storeToSliceExamples {
		slice := storeToKeySortedSlice(example.input)
		if !compareSlices(example.expected, slice) {
			t.Errorf("storeToKeySortedSlice returned unexpected data: got %v want %v", slice, example.expected)
		}
	}
}
