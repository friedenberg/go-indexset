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
