// Code generated by "stringer -type=comparisonOverlap"; DO NOT EDIT.

package indexset

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[comparisonOverlapUnknown-4]
	_ = x[comparisonOverlapEqual-5]
	_ = x[comparisonOverlapInside-6]
	_ = x[comparisonOverlapLeftInside-7]
	_ = x[comparisonOverlapRightInside-8]
	_ = x[comparisonOverlapRightOutside-9]
}

const _comparisonOverlap_name = "comparisonOverlapUnknowncomparisonOverlapEqualcomparisonOverlapInsidecomparisonOverlapLeftInsidecomparisonOverlapRightInsidecomparisonOverlapRightOutside"

var _comparisonOverlap_index = [...]uint8{0, 24, 46, 69, 96, 124, 153}

func (i comparisonOverlap) String() string {
	i -= 4
	if i < 0 || i >= comparisonOverlap(len(_comparisonOverlap_index)-1) {
		return "comparisonOverlap(" + strconv.FormatInt(int64(i+4), 10) + ")"
	}
	return _comparisonOverlap_name[_comparisonOverlap_index[i]:_comparisonOverlap_index[i+1]]
}