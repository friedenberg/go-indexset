package indexset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type combineSuccessTestCase struct {
	description  string
	a            [3]int64
	b            [3]int64
	replacements [][3]int64
	carryover    indexRange
}

func combineSuccessTestCases() []combineSuccessTestCase {
	return []combineSuccessTestCase{
		{
			"overlap same",
			[3]int64{1, 2, 1},
			[3]int64{1, 2, 1},
			[][3]int64{
				[3]int64{1, 2, 2},
			},
			indexRangeZero,
		},
		{
			"inside, secondary",
			[3]int64{1, 5, 1},
			[3]int64{3, 4, 1},
			[][3]int64{
				[3]int64{1, 2, 1},
				[3]int64{3, 4, 2},
				[3]int64{5, 5, 1},
			},
			indexRangeZero,
		},
		{
			"inside, primary",
			[3]int64{3, 4, 1},
			[3]int64{1, 5, 1},
			[][3]int64{
				[3]int64{1, 2, 1},
				[3]int64{3, 4, 2},
			},
			indexRange{5, 5, 1},
		},
		{
			"inside right, secondary",
			[3]int64{1, 5, 1},
			[3]int64{3, 5, 1},
			[][3]int64{
				[3]int64{1, 2, 1},
				[3]int64{3, 5, 2},
			},
			indexRangeZero,
		},
		{
			"inside right, primary",
			[3]int64{3, 5, 1},
			[3]int64{1, 5, 1},
			[][3]int64{
				[3]int64{1, 2, 1},
				[3]int64{3, 5, 2},
			},
			indexRangeZero,
		},
		{
			"inside left, secondary",
			[3]int64{1, 5, 1},
			[3]int64{1, 3, 1},
			[][3]int64{
				[3]int64{1, 3, 2},
				[3]int64{4, 5, 1},
			},
			indexRangeZero,
		},
		{
			"inside left, primary",
			[3]int64{1, 3, 1},
			[3]int64{1, 5, 1},
			[][3]int64{
				[3]int64{1, 3, 2},
			},
			indexRange{4, 5, 1},
		},
		{
			"outside right, secondary",
			[3]int64{1, 5, 1},
			[3]int64{4, 6, 1},
			[][3]int64{
				[3]int64{1, 3, 1},
				[3]int64{4, 5, 2},
			},
			indexRange{5, 6, 1},
		},
		{
			"outside right, primary",
			[3]int64{4, 6, 1},
			[3]int64{1, 5, 1},
			[][3]int64{
				[3]int64{1, 3, 1},
				[3]int64{4, 5, 2},
				[3]int64{5, 6, 1},
			},
			indexRangeZero,
		},
	}
}

func TestCombine(t *testing.T) {
	for _, testcase := range combineSuccessTestCases() {
		t.Run(
			testcase.description,
			func(t *testing.T) {
				a, err := MakeRange(testcase.a[0], testcase.a[1], testcase.a[2])
				assert.Nil(t, err)

				b, err := MakeRange(testcase.b[0], testcase.b[1], testcase.b[2])
				assert.Nil(t, err)

				var replacements []indexRange

				if len(testcase.replacements) > 0 {
					replacements, err = MakeRanges(testcase.replacements...)
					assert.Nil(t, err)
				}

				actualReplacements, actualCarryover, err := a.combine(*b)
				assert.Nil(t, err)
				assert.Equal(t, replacements, actualReplacements)
				assert.Equal(t, testcase.carryover, actualCarryover)
			},
		)
	}
}
