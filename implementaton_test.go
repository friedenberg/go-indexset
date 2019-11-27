package indexset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type iterationTestCase struct {
	description         string
	indexRanges         []indexRange
	expectedIndexRanges []indexRange
}

func iterationTestCases() []iterationTestCase {
	return []iterationTestCase{
		{
			description: "two ranges, overlap",
			indexRanges: []indexRange{
				{1, 5, 1},
				{1, 5, 1},
			},
			expectedIndexRanges: []indexRange{
				{1, 5, 2},
			},
		},
		{
			description: "one range",
			indexRanges: []indexRange{
				{1, 5, 1},
			},
			expectedIndexRanges: []indexRange{
				{1, 5, 1},
			},
		},
		{
			description: "overlap same",
			indexRanges: []indexRange{
				{1, 5, 1},
				{1, 5, 1},
				{1, 5, 1},
			},
			expectedIndexRanges: []indexRange{
				{1, 5, 3},
			},
		},
		{
			description: "overlap with other",
			indexRanges: []indexRange{
				{1, 5, 1},
				{5, 10, 1},
				{1, 5, 1},
			},
			expectedIndexRanges: []indexRange{
				{1, 4, 2},
				{5, 5, 3},
				{6, 10, 1},
			},
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
			expectedIndexRanges: []indexRange{
				{1, 1, 1},
				{2, 2, 2},
				{3, 3, 3},
				{4, 4, 4},
				{5, 5, 5},
				{6, 6, 5},
			},
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
			expectedIndexRanges: []indexRange{
				{1, 1, 3},
				{2, 4, 4},
				{5, 5, 5},
				{6, 8, 2},
				{8, 9, 2},
				{9, 10, 1},
			},
		},
	}
}

func implementationsToTest(t *testing.T) map[string]func() Implementation {
	return map[string]func() Implementation{
		"linked_list": func() Implementation { return &linkedList{} },
	}
}

func TestIteration(t *testing.T) {
	for implementationName, implementation := range implementationsToTest(t) {
		for _, test := range iterationTestCases() {
			t.Run(
				fmt.Sprintf("%s: %s", implementationName, test.description),
				func(t *testing.T) {
					set := &Set{Implementation: implementation()}

					for _, indexRange := range test.indexRanges {
						err := set.Add(indexRange)
						assert.Nil(t, err)
					}

					idx := 0

					set.Implementation.Do(
						func(m Member) bool {
							assert.Equal(t, test.expectedIndexRanges[idx], m.IndexRange())
							idx++
							return false
						},
					)
				},
			)
		}
	}
}

type overlappingPair struct {
	overlap     indexRange
	overlapping []indexRange
}

type overlappingTestCase struct {
	description string
	indexRanges []indexRange
	pairs       []overlappingPair
}

func overlappingTestCases() []overlappingTestCase {
	return []overlappingTestCase{
		{
			"basic",
			[]indexRange{
				{1, 5, 1},
				{6, 8, 1},
			},
			[]overlappingPair{
				overlappingPair{
					overlap: indexRange{2, 3, 1},
					overlapping: []indexRange{
						indexRange{1, 5, 1},
					},
				},
				overlappingPair{
					overlap: indexRange{6, 7, 1},
					overlapping: []indexRange{
						indexRange{6, 8, 1},
					},
				},
				overlappingPair{
					overlap: indexRange{4, 7, 1},
					overlapping: []indexRange{
						indexRange{1, 5, 1},
						indexRange{6, 8, 1},
					},
				},
			},
		},
	}
}

func TestFindOverlapping(t *testing.T) {
	for implementationName, implementation := range implementationsToTest(t) {
		for _, test := range overlappingTestCases() {
			t.Run(
				fmt.Sprintf("%s: %s", implementationName, test.description),
				func(t *testing.T) {
					set := &Set{Implementation: implementation()}

					for _, indexRange := range test.indexRanges {
						err := set.Add(indexRange)
						assert.Nil(t, err)
					}

					for _, pair := range test.pairs {
						for i, overlappingRange := range set.FindOverlapping(pair.overlap) {
							assert.Equal(t, pair.overlapping[i], overlappingRange.IndexRange())
						}
					}
				},
			)
		}
	}
}

type replaceOperation struct {
	selectionIndex int
	replacement    []indexRange
}

type replaceTestCase struct {
	description string
	indexRanges []indexRange
	operations  []replaceOperation
	endState    []indexRange
}

func replaceTestCases() []replaceTestCase {
	return []replaceTestCase{
		{
			"basic",
			[]indexRange{
				{1, 5, 1},
				{6, 8, 1},
			},
			[]replaceOperation{
				replaceOperation{
					0,
					[]indexRange{
						indexRange{2, 5, 1},
					},
				},
			},
			[]indexRange{
				{2, 5, 1},
				{6, 8, 1},
			},
		},
		{
			"basic",
			[]indexRange{
				{1, 5, 1},
				{6, 8, 1},
			},
			[]replaceOperation{
				replaceOperation{
					0,
					[]indexRange{
						indexRange{1, 1, 2},
						indexRange{2, 5, 1},
					},
				},
			},
			[]indexRange{
				{1, 1, 2},
				{2, 5, 1},
				{6, 8, 1},
			},
		},
	}
}

func TestReplace(t *testing.T) {
	for implementationName, implementation := range implementationsToTest(t) {
		for _, test := range replaceTestCases() {
			t.Run(
				fmt.Sprintf("%s: %s", implementationName, test.description),
				func(t *testing.T) {
					set := &Set{Implementation: implementation()}

					for _, indexRange := range test.indexRanges {
						err := set.Add(indexRange)
						assert.Nil(t, err)
					}

					for _, operation := range test.operations {
						nth := set.Nth(operation.selectionIndex)
						set.Replace(nth, operation.replacement...)
					}

					idx := 0

					set.Implementation.Do(
						func(m Member) bool {
							assert.Equal(t, test.endState[idx], m.IndexRange())
							idx++
							return false
						},
					)
				},
			)
		}
	}
}
