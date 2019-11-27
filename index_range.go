//go:generate stringer -type=comparisonPosition

package indexset

import (
	"fmt"
	"strings"
)

const (
	comparisonPositionUnknown = comparisonPosition(iota)
	comparisonPositionLeft
	comparisonPositionRight
	comparisonPositionOverlap
)

var (
	indexRangeZero = indexRange{0, 0, 0}
)

type comparisonPosition int

type indexRange struct {
	left   int64
	right  int64
	weight int64
}

func MakeRange(left int64, right int64, weight int64) (*indexRange, error) {
	if left > right {
		return nil, fmt.Errorf("invalid range: left (%v) is larger than right (%v)", left, right)
	}

	if left < 0 {
		return nil, fmt.Errorf("invalid range: left (%v) is less than 0", left)
	}

	if right < 0 {
		return nil, fmt.Errorf("invalid range: right (%v) is less than 0", right)
	}

	val := indexRange{
		left:   left,
		right:  right,
		weight: weight,
	}

	return &val, nil
}

func MakeRanges(ranges ...[3]int64) ([]indexRange, error) {
	output := make([]indexRange, len(ranges))

	for i, someRange := range ranges {
		validatedRange, err := MakeRange(someRange[0], someRange[1], someRange[2])

		if err != nil {
			//always return a valid output range
			return output, err
		}

		output[i] = *validatedRange
	}

	return output, nil
}

func (a indexRange) comparePosition(b indexRange) comparisonPosition {
	if a.right < b.left {
		return comparisonPositionRight
	} else if b.right < a.left {
		return comparisonPositionLeft
	} else {
		return comparisonPositionOverlap
	}
}

func (q indexRange) String() string {
	repeat := func(str string, count int64) string {
		if count > 0 {
			return strings.Repeat(str, int(count))
		}

		return ""
	}

	return fmt.Sprintf(
		"%v:|%v_%v|",
		q.weight,
		q.left,
		q.right,
	)

	return fmt.Sprintf(
		"%v|%v%v%v|",
		repeat(" ", q.left-1),
		q.left,
		repeat("_", q.right-q.left-1),
		q.right,
	)
}

func (a indexRange) SplitWith(b indexRange) (replacements []indexRange, carryover indexRange, err error) {
	relation := makeIndexRangeOverlap(a, b)
	splitFunc := relation.splitFunc()
	return splitFunc(relation)
}
