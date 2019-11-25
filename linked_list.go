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

type linkedList struct {
	head *node
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

func (q *linkedList) Add(newRange indexRange) error {
	if q.head == nil {
		q.head = &node{
			indexRange: newRange,
		}

		return nil
	}

	currentNode := q.head
	previousNode := currentNode
	previousComparePosition := comparisonPositionRight

	count := 0

	//finding overlapping index ranges
	for {
		count += 1

		if currentNode == nil {
			break
		}

		position := currentNode.indexRange.comparePosition(newRange)

		switch position {
		case comparisonPositionLeft:
			switch previousComparePosition {
			case comparisonPositionOverlap:
				//noop because the previous node overlaps newRange, it'll take care of combining
				//with newRange

			case comparisonPositionRight:
				//this node has no overlap with our current set, so we can add it
				//immediately and stop iterating.
				//todo confirm this works with the start of previous head
				newNode := &node{
					indexRange: newRange,
					next:       currentNode,
				}

				if previousNode == currentNode {
					q.head = newNode
				} else {
					previousNode.setNext(newNode)
					newNode.setPrev(previousNode)
					currentNode.setPrev(newNode)
				}

				return nil

			default:
				//todo more detail
				return errors.New("impossible state")
			}

		case comparisonPositionOverlap:
			switch previousComparePosition {
			case comparisonPositionOverlap:
				//todo make this an impossible state
				//we're continuing an existing overlap chain

			case comparisonPositionRight:
				//we're starting a brand new overlap chain
				nextNode := currentNode.next
				newTail, modifiedNewRange, err := currentNode.combine(newRange)

				if err != nil {
					return nil
				}

				currentNode = newTail

				if nextNode != nil {
					newTail.setNext(nextNode)
					nextNode.setPrev(newTail)
				}

				if modifiedNewRange != nil {
					newRange = *modifiedNewRange
				} else {
					return nil
				}

			default:
				//todo more detail
				return errors.New("impossible state")
			}

		case comparisonPositionRight:
			switch previousComparePosition {
			case comparisonPositionRight:
				//peek at the next node to see if we're at the end. If we are, we
				//perform what the next node would have and stop
				nextNode := currentNode.next
				if nextNode == nil {
					newNode := &node{
						indexRange: newRange,
						prev:       currentNode,
					}
					currentNode.setNext(newNode)
					return nil
				}

			default:
				//todo more detail
				return errors.New("impossible state")
			}

		default:
			//todo more detail
			return errors.New("impossible state")
		}

		currentNode = currentNode.next
	}

	return nil
}

func (i *linkedList) Do(f func(indexRange)) {
	currentNode := i.head
	var prevNode *node

	for {
		if currentNode == nil {
			break
		}

		if prevNode != nil && (currentNode.indexRange.left < prevNode.indexRange.right || currentNode.indexRange.right < prevNode.indexRange.right) {
			panic(fmt.Errorf("ranges out of order: prev (%s), current(%s)", prevNode.indexRange, currentNode.indexRange))
		}

		f(currentNode.indexRange)

		prevNode = currentNode
		currentNode = currentNode.next
	}
}
