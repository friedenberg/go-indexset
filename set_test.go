package indexset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testcase struct {
	description string
	indexRanges []indexRange
	expectedMax int64
}

func basicTestCases() []testcase {
	return []testcase{
		{
			description: "one range",
			indexRanges: []indexRange{
				{1, 5, 1},
			},
			expectedMax: 1,
		},
		{
			description: "overlap same",
			indexRanges: []indexRange{
				{1, 5, 1},
				{1, 5, 1},
				{1, 5, 1},
			},
			expectedMax: 3,
		},
		{
			description: "overlap with other",
			indexRanges: []indexRange{
				{1, 5, 1},
				{5, 10, 1},
				{1, 5, 1},
			},
			expectedMax: 3,
		},
		{
			description: "overlap with new",
			indexRanges: []indexRange{
				{1, 5, 1},
				{2, 6, 1},
				{3, 7, 1},
				{4, 8, 1},
				{5, 9, 1},
			},
			expectedMax: 5,
		},
		{
			description: "overlap with various sizes",
			indexRanges: []indexRange{
				{1, 5, 1},
				{1, 6, 1},
				{1, 10, 1},
				{2, 8, 1},
				{5, 9, 1},
			},
			expectedMax: 5,
		},
	}
}

func TestAdd(t *testing.T) {
	for _, test := range basicTestCases() {
		t.Run(
			test.description,
			func(t *testing.T) {
				set := &Set{
					Implementation: &linkedList{},
				}

				for _, indexRange := range test.indexRanges {
					set.Add(indexRange)
				}

				max := set.Max()

				assert.Equal(t, test.expectedMax, max)
			},
		)
	}
}
