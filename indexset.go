package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	comparisonPositionUnknown comparisonPosition = iota
	comparisonPositionLeft
	comparisonPositionRight
	comparisonPositionOverlap

	comparisonOverlapUnknown comparisonOverlap = iota
	comparisonOverlapEqual
	comparisonOverlapInside
	comparisonOverlapRightInside
	comparisonOverlapRightOutside
)

var (
	logger = log.New(os.Stderr, "", 0)
)

type comparisonPosition int
type comparisonOverlap int

type indexRange struct {
	left   int64
	right  int64
	weight int64
}

func makeRange(left int64, right int64, weight int64) indexRange {
	if left > right {
		panic(fmt.Sprintf("invalid range: left (%v) is larger than right (%v)", left, right))
	}

	if left < 0 {
		panic(fmt.Sprintf("invalid range: left (%v) is less than 0", left))
	}

	if right < 0 {
		panic(fmt.Sprintf("invalid range: right (%v) is less than 0", right))
	}

	return indexRange{
		left:   left,
		right:  right,
		weight: weight,
	}
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
		"|%v_%v|",
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

	if primary.right > secondary.right {
		return comparisonOverlapInside
	} else if primary.right < secondary.right {
		return comparisonOverlapRightOutside
	}

	return comparisonOverlapUnknown
}

type indexRangeOverlap struct {
	primary   indexRange
	secondary indexRange
	overlap   comparisonOverlap
}

func (b indexRange) compare(a indexRange) indexRangeOverlap {
	primary, secondary := b.choosePrimary(a)
	overlap := primary.calculateOverlap(secondary)

	return indexRangeOverlap{
		primary:   primary,
		secondary: secondary,
		overlap:   overlap,
	}
}

type node struct {
	indexRange indexRange
	prev       *node
	next       *node
}

func (n *node) setPrev(prev *node) {
	if n.next == prev {
		panic("circular ref")
	}

	n.prev = prev
}

func (n *node) setNext(next *node) {
	if n.prev == next {
		panic("circular ref")
	}

	n.next = next
}

func (n *node) String() string {
	return n.indexRange.String()
}

func (n *node) iterate(f func(*node)) {
	currentNode := n

	for {
		if currentNode == nil {
			break
		}

		f(currentNode)

		currentNode = currentNode.next
	}
}

func (n *node) combine(b indexRange) (tail *node, newRange *indexRange) {
	logger.Println("starting combine")
	tail = n

	comparison := n.indexRange.compare(b)

	a := comparison.primary
	n.indexRange = a
	b = comparison.secondary

	logger.Println("a:", n.indexRange)
	logger.Println("b:", b)

	switch comparison.overlap {
	case comparisonOverlapEqual:
		n.indexRange.weight += b.weight

	case comparisonOverlapInside:
		//left
		n.indexRange.right = b.left - 1

		//middle
		middle := &node{
			indexRange: b,
			prev:       n,
		}
		n.setNext(middle)

		//right
		right := &node{
			indexRange: makeRange(b.right+1, a.right, a.weight+b.weight),
			prev:       middle,
		}
		middle.setNext(right)

		tail = right

	case comparisonOverlapRightInside:
		n.indexRange.right = b.left - 1

		//right
		right := &node{
			indexRange: makeRange(a.right, b.right, a.weight+b.weight),
			prev:       n,
		}
		n.setNext(right)

		tail = right

	case comparisonOverlapRightOutside:
		someRange := makeRange(a.right+1, b.right, b.weight)
		newRange = &someRange

		//left
		n.indexRange.right = b.left - 1

		//middle
		middle := &node{
			indexRange: makeRange(b.left, a.right, a.weight+b.weight),
			prev:       n,
		}
		n.setNext(middle)

		//right
		right := &node{
			indexRange: makeRange(a.right+1, b.right, b.weight),
			prev:       middle,
		}
		middle.setNext(right)

		tail = right

	case comparisonOverlapUnknown:
		panic(comparison)
	}

	logger.Println("ending combine")
	return tail, newRange
}

type indexSet struct {
	head *node
}

func (q *indexSet) add(newRange indexRange) {
	logger.Println("starting add")

	if q.head == nil {
		logger.Println("no nodes, adding new and returning")

		q.head = &node{
			indexRange: newRange,
		}

		logger.Printf("new: %v", q.head)

		return
	}

	currentNode := q.head
	previousNode := currentNode
	previousComparePosition := comparisonPositionRight

	//finding overlapping index ranges
	for {
		logger.Println("looping")
		logger.Println("current:", q)

		if currentNode == nil {
			logger.Println("end loop")
			break
		}

		position := currentNode.indexRange.comparePosition(newRange)

		switch position {
		case comparisonPositionLeft:
			logger.Println("current left")
			switch previousComparePosition {
			case comparisonPositionOverlap:
				logger.Println("prev overlap: noop")
				//noop because the previous node overlaps newRange, it'll take care of combining
				//with newRange

			case comparisonPositionRight:
				logger.Println("prev right: add node")
				//this node has no overlap with our current set, so we can add it
				//immediately and stop iterating.
				//todo confirm this works with the start of previous head
				newNode := &node{
					indexRange: newRange,
					next:       currentNode,
				}

				if previousNode == currentNode {
					logger.Println("current is head")
					q.head = newNode
				} else {
					logger.Println("inserting new node")
					previousNode.setNext(newNode)
					newNode.setPrev(previousNode)
					currentNode.setPrev(newNode)
				}

				return

			default:
				panic("impossible state")
			}

		case comparisonPositionOverlap:
			logger.Println("current overlap")
			switch previousComparePosition {
			case comparisonPositionOverlap:
				logger.Println("prev overlap: impossible?")
				//todo make this an impossible state
				//we're continuing an existing overlap chain
			case comparisonPositionRight:
				logger.Println("prev right: start new chain")
				//we're starting a brand new overlap chain
				nextNode := currentNode.next
				newTail, modifiedNewRange := currentNode.combine(newRange)

				if nextNode != nil {
					newTail.setNext(nextNode)
					nextNode.setPrev(newTail)

					if modifiedNewRange != nil {
						newRange = *modifiedNewRange
					}
				}

				panic("")

			default:
				panic("impossible state")
			}

		case comparisonPositionRight:
			logger.Println("current right")
			switch previousComparePosition {
			case comparisonPositionRight:
				logger.Println("prev right: peek and do next node's work")
				//peek at the next node to see if we're at the end. If we are, we
				//perform what the next node would have and stop
				nextNode := currentNode.next
				if nextNode == nil {
					newNode := &node{
						indexRange: newRange,
						prev:       currentNode,
					}
					currentNode.setNext(newNode)
					return
				}

			default:
				panic("impossible state")
			}

		default:
			panic("impossible state")
		}

		currentNode = currentNode.next
	}
}

func (i *indexSet) String() string {
	sb := strings.Builder{}

	i.head.iterate(
		func(n *node) {
			sb.WriteString(n.indexRange.String())
			sb.WriteString("\n")
		},
	)

	return sb.String()
}

func (q *indexSet) max() int64 {
	max := int64(0)

	currentNode := q.head

	q.head.iterate(
		func(n *node) {
			nodeWeight := currentNode.indexRange.weight

			if nodeWeight > max {
				max = nodeWeight
			}
		},
	)

	return max
}

// Complete the arrayManipulation function below.
func arrayManipulation(n int32, someindexSet [][]int32) int64 {
	q := &indexSet{}

	for _, a := range someindexSet {
		aIndexRange := makeRange(int64(a[0]), int64(a[1]), int64(a[2]))

		q.add(aIndexRange)
		fmt.Println("adding:\n", q)
	}

	fmt.Println("final:\n", q)

	return q.max()
}
