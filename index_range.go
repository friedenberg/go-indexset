//go:generate stringer -type=comparisonPosition
//go:generate stringer -type=comparisonOverlap

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

	comparisonOverlapUnknown = comparisonOverlap(iota)
	comparisonOverlapEqual
	comparisonOverlapInside
	comparisonOverlapLeftInside
	comparisonOverlapRightInside
	comparisonOverlapRightOutside
)

var (
	indexRangeZero = indexRange{0, 0, 0}
)

type comparisonPosition int
type comparisonOverlap int

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

func (b indexRange) choosePrimary(a indexRange) (indexRange, indexRange) {
	if a.left < b.left {
		return a, b
	} else if a.left > b.left {
		return b, a
	} else if a.right > b.right {
		return a, b
	} else {
		return b, a
	}
}

/*
   ___ ____

   -------
   -------

   _______
        __

   _______
        _______

   ________
       __
*/
func (primary indexRange) calculateOverlap(secondary indexRange) comparisonOverlap {
	if primary.right < secondary.left {
		return comparisonOverlapUnknown
	}

	if primary.right == secondary.right {
		if primary.left < secondary.left {
			return comparisonOverlapRightInside
		} else if primary.left == secondary.left {
			return comparisonOverlapEqual
		}
	}

	if primary.left == secondary.left {
		if primary.right > secondary.left {
			return comparisonOverlapLeftInside
		} else {
			return comparisonOverlapUnknown
		}
	}

	if primary.right > secondary.right {
		return comparisonOverlapInside
	} else if primary.right < secondary.right {
		return comparisonOverlapRightOutside
	}

	return comparisonOverlapUnknown
}

type indexRangeOverlap struct {
	primary         indexRange
	secondary       indexRange
	overlap         comparisonOverlap
	rightIsNewRange bool
}

func (a indexRange) compare(b indexRange) indexRangeOverlap {
	primary, secondary := a.choosePrimary(b)
	overlap := primary.calculateOverlap(secondary)

	rightIsNewRange := b.right > a.right

	return indexRangeOverlap{
		primary:         primary,
		secondary:       secondary,
		overlap:         overlap,
		rightIsNewRange: rightIsNewRange,
	}
}

func (a indexRange) combine(b indexRange) (replacement []indexRange, carryover indexRange, err error) {
	comparison := a.compare(b)

	a = comparison.primary
	b = comparison.secondary

	switch comparison.overlap {
	case comparisonOverlapEqual:
		replacement, err = MakeRanges([3]int64{a.left, a.right, a.weight + b.weight})

	case comparisonOverlapLeftInside:
		replacement, err = MakeRanges(
			[3]int64{a.left, b.right, a.weight + b.weight},
			[3]int64{b.right + 1, a.right, a.weight},
		)

		if comparison.rightIsNewRange {
			carryover = replacement[1]
			replacement = replacement[0:1]
		}

	case comparisonOverlapInside:
		replacement, err = MakeRanges(
			[3]int64{a.left, b.left - 1, a.weight},
			[3]int64{b.left, b.right, a.weight + b.weight},
			[3]int64{b.right + 1, a.right, a.weight},
		)

		if comparison.rightIsNewRange {
			carryover = replacement[2]
			replacement = replacement[0:2]
		}

	case comparisonOverlapRightInside:
		replacement, err = MakeRanges(
			[3]int64{a.left, b.left - 1, a.weight},
			[3]int64{b.left, b.right, a.weight + b.weight},
		)

	case comparisonOverlapRightOutside:
		rightRange := [3]int64{a.right, b.right, b.weight}

		if comparison.rightIsNewRange {
			replacement, err = MakeRanges(
				[3]int64{a.left, b.left - 1, a.weight},
				[3]int64{b.left, a.right, a.weight + b.weight},
				rightRange,
			)

			carryover = replacement[2]
			replacement = replacement[0:2]
		} else {
			replacement, err = MakeRanges(
				[3]int64{a.left, b.left - 1, a.weight},
				[3]int64{b.left, a.right, a.weight + b.weight},
				rightRange,
			)
		}

	default:
		err = fmt.Errorf("failed to combine nodes unknown overlap")
	}

	if err != nil {
		return nil, indexRangeZero, err
	}

	return replacement, carryover, nil
}
