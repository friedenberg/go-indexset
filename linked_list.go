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
