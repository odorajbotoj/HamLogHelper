package main

type LinkedList struct {
	head *LinkedListNode
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

func (ll *LinkedList) set(pos int64, d *LogLine) {
	if pos < 0 {
		pos = -pos // mark deleted
	}
	current := ll.head
	for range pos {
		if current.next == nil {
			current.next = new(LinkedListNode)
		}
		current = current.next
	}
	current.data = *d
}

func (ll *LinkedList) cleanAndDump() []LogLine {
	var ret []LogLine
	current := ll.head
	var reindex int64 = 0
	for current.next != nil {
		if current.next.data.Index <= 0 {
			current.next = current.next.next // delete
			continue
		}
		current = current.next
		current.data.Index = reindex + 1
		ret = append(ret, current.data)
		reindex++
	}
	return ret
}
