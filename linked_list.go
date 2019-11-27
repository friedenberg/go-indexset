package indexset

import (
	"errors"
	"fmt"
)

type node struct {
	indexRange indexRange
	prev       *node
	next       *node
}

func (n *node) IndexRange() indexRange {
	return n.indexRange
}

func (n *node) setPrev(prev *node) error {
	if n.next == prev {
		return errors.New("failed to set prev because it's a circular reference")
	}

	n.prev = prev
	return nil
}

func (n *node) setNext(next *node) error {
	if n.prev == next {
		return errors.New("failed to set next because it's a circular reference")
	}

	n.next = next
	return nil
}

func (n *node) String() string {
	return n.indexRange.String()
}

func (n *node) combine(b indexRange) (tail *node, newRange *indexRange, err error) {
	tail = n

	comparison := n.indexRange.compare(b)

	a := comparison.primary
	n.indexRange = a
	b = comparison.secondary

	switch comparison.overlap {
	case comparisonOverlapEqual:
		n.indexRange.weight += b.weight

	case comparisonOverlapLeftInside:
		n.indexRange.right = b.right
		n.indexRange.weight += b.weight

		rightRange, err := MakeRange(b.right+1, a.right, a.weight)

		if err != nil {
			return nil, nil, err
		}

		if comparison.rightIsNewRange {
			newRange = rightRange
		} else {
			two := &node{
				indexRange: *rightRange,
				prev:       n,
			}

			n.setNext(two)
			tail = two
		}

	case comparisonOverlapInside:
		//left
		n.indexRange.right = b.left - 1

		//middle
		middle := &node{
			indexRange: b,
			prev:       n,
		}
		n.setNext(middle)

		rightRange, err := MakeRange(b.right+1, a.right, a.weight+b.weight)

		if err != nil {
			return nil, nil, err
		}

		if comparison.rightIsNewRange {
			newRange = rightRange
			tail = middle
		} else {
			//right
			right := &node{
				indexRange: *rightRange,
				prev:       middle,
			}

			middle.setNext(right)
			tail = right
		}

	case comparisonOverlapRightInside:
		n.indexRange.right = b.left - 1

		rightRange, err := MakeRange(a.right, b.right, a.weight+b.weight)

		if err != nil {
			return nil, nil, err
		}

		//right
		right := &node{
			indexRange: *rightRange,
			prev:       n,
		}
		n.setNext(right)

		tail = right

	case comparisonOverlapRightOutside:
		rightRange, err := MakeRange(a.right+1, b.right, b.weight)

		if err != nil {
			return nil, nil, err
		}

		var right *node

		if comparison.rightIsNewRange {
			newRange = rightRange

			//left
			n.indexRange.right = b.left - 1

			rightRange, err := MakeRange(b.left, a.right, a.weight+b.weight)

			if err != nil {
				return nil, nil, err
			}

			//right
			right = &node{
				indexRange: *rightRange,
				prev:       n,
			}

			n.setNext(right)

		} else {
			//left
			n.indexRange.right = b.left - 1

			middleRange, err := MakeRange(b.left, a.right, a.weight+b.weight)

			if err != nil {
				return nil, nil, err
			}

			//middle
			middle := &node{
				indexRange: *middleRange,
				prev:       n,
			}

			n.setNext(middle)

			//right
			right = &node{
				indexRange: *rightRange,
				prev:       middle,
			}

			middle.setNext(right)

		}

		tail = right

	case comparisonOverlapUnknown:
		return nil, nil, fmt.Errorf("failed to combine nodes unknown overlap")
	}

	return tail, newRange, nil
}

type linkedList struct {
	head *node
}

func (l *linkedList) FindOverlapping(overlap indexRange) []Member {
	overlapping := make([]Member, 0)

	l.Do(
		func(m Member) bool {
			switch m.IndexRange().comparePosition(overlap) {
			case comparisonPositionOverlap:
				overlapping = append(overlapping, m)
			}

			return false
		},
	)

	return overlapping
}

func (l *linkedList) Replace(original Member, replacements ...indexRange) error {
	n, ok := original.(*node)

	if !ok {
		return errors.New("member is not an instance of node")
	}

	switch len(replacements) {
	case 0:
		if n == l.head {
			l.head = n.next
			l.head.prev = nil
		} else if n.next != nil {
			prev := n.prev
			next := n.next
			prev.next = next
			next.prev = prev
			n.next = nil
			n.prev = nil
		} else {
			n.prev.next = nil
			n.prev = nil
		}

	case 1:
		n.indexRange = replacements[0]

	default:
		n.indexRange = replacements[0]
		next := n.next

		current := n

		for _, v := range replacements[1:] {
			newNode := &node{
				indexRange: v,
				prev:       current,
			}

			current.next = newNode
			current = newNode
		}

		current.next = next

		if next != nil {
			next.prev = current
		}
	}

	return nil
}

func (q *linkedList) AddOrFindOverlapping(newRange indexRange) (overlapping []Member, err error) {
	if q.head == nil {
		q.head = &node{
			indexRange: newRange,
		}

		return nil, nil
	}

	currentNode := q.head

	//finding overlapping index ranges
	for {
		if currentNode == nil {
			break
		}

		position := currentNode.indexRange.comparePosition(newRange)

		switch position {
		case comparisonPositionLeft:
			if len(overlapping) == 0 {
				next := currentNode.next

				newNode := &node{
					indexRange: newRange,
					prev:       currentNode,
					next:       next,
				}

				if next != nil {
					next.prev = newNode
				}

				currentNode.next = newNode
			}

			break

		case comparisonPositionOverlap:
			overlapping = append(overlapping, currentNode)

		case comparisonPositionRight:
			//noop

		default:
			//todo more detail
			err = errors.New("impossible state")
			break
		}

		currentNode = currentNode.next
	}

	return overlapping, err
}

func (i *linkedList) Do(f func(Member) (stop bool)) {
	currentNode := i.head
	var prevNode *node

	for {
		if currentNode == nil {
			break
		}

		if prevNode != nil && (currentNode.indexRange.left < prevNode.indexRange.right || currentNode.indexRange.right < prevNode.indexRange.right) {
			panic(fmt.Errorf("ranges out of order: prev (%s), current(%s)", prevNode.indexRange, currentNode.indexRange))
		}

		stop := f(currentNode)

		if stop {
			break
		}

		prevNode = currentNode
		currentNode = currentNode.next
	}
}
