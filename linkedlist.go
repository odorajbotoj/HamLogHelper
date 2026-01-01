package main

type LinkedList struct {
	maxPos uint64
	head   *LinkedListNode
}

type LinkedListNode struct {
	data LogLine
	next *LinkedListNode
}

func newLinkedList() *LinkedList {
	p := new(LinkedList)
	p.head = new(LinkedListNode)
	return p
}

func (ll *LinkedList) set(pos uint64, d *LogLine) {
	current := ll.head
	for range pos {
		if current.next == nil {
			current.next = new(LinkedListNode)
		}
		current = current.next
	}
	current.data = *d
	if pos > ll.maxPos {
		ll.maxPos = pos
	}
}

func (ll *LinkedList) dumpTill(pos uint64) []LogLine {
	ret := make([]LogLine, pos)
	current := ll.head
	for i := range pos {
		if current.next == nil {
			current.next = new(LinkedListNode)
		}
		current = current.next
		ret[i] = current.data
	}
	if pos > ll.maxPos {
		ll.maxPos = pos
	}
	return ret
}
