package indexset

import (
	"errors"
)

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
type indexRangeOverlap struct {
	a, b            indexRange
	rightIsNewRange bool
}

func makeIndexRangeOverlap(a, b indexRange) indexRangeOverlap {
	relation := indexRangeOverlap{
		a: a,
		b: b,
	}

	if a.left > b.left || (a.left == b.left && a.right < b.right) {
		relation.a = b
		relation.b = a
	}

	relation.rightIsNewRange = b.right > a.right

	return relation
}

func (r indexRangeOverlap) splitFunc() indexRangeSplitFunc {
	if r.a.right < r.b.left {
		return indexRangeSplitFuncUnknown
	}

	if r.a.right == r.b.right {
		if r.a.left < r.b.left {
			return indexRangeSplitFuncRightInside
		} else if r.a.left == r.b.left {
			return indexRangeSplitFuncEqual
		}
	}

	if r.a.left == r.b.left {
		if r.a.right > r.b.left {
			return indexRangeSplitFuncLeftInside
		} else {
			return indexRangeSplitFuncUnknown
		}
	}

	if r.a.right > r.b.right {
		return indexRangeSplitFuncInside
	} else if r.a.right < r.b.right {
		return indexRangeSplitFuncRightOutside
	}

	return indexRangeSplitFuncUnknown
}

type indexRangeSplitFunc func(indexRangeOverlap) (replacement []indexRange, carryover indexRange, err error)

var (
	indexRangeSplitFuncUnknown = func(_ indexRangeOverlap) ([]indexRange, indexRange, error) {
		return nil, indexRangeZero, errors.New("failed to combine nodes: unknown overlap")
	}

	indexRangeSplitFuncEqual = func(r indexRangeOverlap) ([]indexRange, indexRange, error) {
		replacement, err := MakeRanges([3]int64{r.a.left, r.a.right, r.a.weight + r.b.weight})
		return replacement, indexRangeZero, err
	}

	indexRangeSplitFuncInside = func(r indexRangeOverlap) ([]indexRange, indexRange, error) {
		replacement, err := MakeRanges(
			[3]int64{r.a.left, r.b.left - 1, r.a.weight},
			[3]int64{r.b.left, r.b.right, r.a.weight + r.b.weight},
			[3]int64{r.b.right + 1, r.a.right, r.a.weight},
		)

		if r.rightIsNewRange {
			return replacement[0:2], replacement[2], err
		} else {
			return replacement, indexRangeZero, err
		}
	}

	indexRangeSplitFuncLeftInside = func(r indexRangeOverlap) ([]indexRange, indexRange, error) {
		replacement, err := MakeRanges(
			[3]int64{r.a.left, r.b.right, r.a.weight + r.b.weight},
			[3]int64{r.b.right + 1, r.a.right, r.a.weight},
		)

		if r.rightIsNewRange {
			return replacement[0:1], replacement[1], err
		} else {
			return replacement, indexRangeZero, err
		}
	}

	indexRangeSplitFuncRightInside = func(r indexRangeOverlap) ([]indexRange, indexRange, error) {
		replacement, err := MakeRanges(
			[3]int64{r.a.left, r.b.left - 1, r.a.weight},
			[3]int64{r.b.left, r.b.right, r.a.weight + r.b.weight},
		)

		return replacement, indexRangeZero, err
	}

	indexRangeSplitFuncRightOutside = func(r indexRangeOverlap) ([]indexRange, indexRange, error) {
		rightRange := [3]int64{r.a.right, r.b.right, r.b.weight}

		if r.rightIsNewRange {
			replacement, err := MakeRanges(
				[3]int64{r.a.left, r.b.left - 1, r.a.weight},
				[3]int64{r.b.left, r.a.right, r.a.weight + r.b.weight},
				rightRange,
			)

			return replacement[0:2], replacement[2], err
		} else {
			replacement, err := MakeRanges(
				[3]int64{r.a.left, r.b.left - 1, r.a.weight},
				[3]int64{r.b.left, r.a.right, r.a.weight + r.b.weight},
				rightRange,
			)

			return replacement, indexRangeZero, err
		}
	}
)
