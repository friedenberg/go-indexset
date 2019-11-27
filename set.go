package indexset

import (
	"fmt"
	"strings"
)

type Member interface {
	IndexRange() indexRange
}

type Implementation interface {
	FindOverlapping(overlap indexRange) []Member
	Replace(original Member, replacements ...indexRange) error
	AddOrFindOverlapping(indexRange) ([]Member, error)
	Do(func(Member) (stop bool))
}

type Set struct {
	Implementation
}

func (s *Set) Nth(n int) Member {
	i := 0
	var found Member

	s.Implementation.Do(
		func(m Member) bool {
			if i == n {
				found = m
				return true
			}

			i++
			return false
		},
	)

	return found
}

func (s *Set) Add(newRange indexRange) error {
	overlapping, err := s.AddOrFindOverlapping(newRange)

	if err != nil {
		return err
	}

	if len(overlapping) == 0 {
		return nil
	}

	carryover := newRange

	for _, currentNode := range overlapping {
		currentRange := currentNode.IndexRange()

		var replacements []indexRange
		replacements, carryover, err = currentRange.SplitWith(carryover)

		if err != nil {
			return err
		}

		if err = s.Replace(currentNode, replacements...); err != nil {
			return err
		}
	}

	return nil
}

func (i *Set) String() string {
	if stringer, ok := i.Implementation.(fmt.Stringer); ok {
		return stringer.String()
	}

	sb := strings.Builder{}
	sb.WriteString("\n")

	i.Do(
		func(m Member) bool {
			sb.WriteString(m.IndexRange().String())
			return false
		},
	)

	return sb.String()
}

func (s *Set) Max() int64 {
	max := int64(0)

	s.Do(
		func(m Member) bool {
			weight := m.IndexRange().weight

			if weight > max {
				max = weight
			}

			return false
		},
	)

	return max
}
