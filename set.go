//go:generate stringer -type=comparisonPosition
//go:generate stringer -type=comparisonOverlap

package indexset

import (
	"fmt"
	"strings"
)

type Implementation interface {
	Add(indexRange) error
	Do(func(indexRange))
}

type Set struct {
	Implementation
}

func (i *Set) String() string {
	if stringer, ok := i.Implementation.(fmt.Stringer); ok {
		return stringer.String()
	}

	sb := strings.Builder{}
	sb.WriteString("\n")

	i.Do(
		func(v indexRange) {
			sb.WriteString(v.String())
		},
	)

	return sb.String()
}

func (s *Set) Max() int64 {
	max := int64(0)

	s.Do(
		func(v indexRange) {
			weight := v.weight

			if weight > max {
				max = weight
			}
		},
	)

	return max
}
