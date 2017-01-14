package containers

type queueItem struct {
	msg Message
	id  uint32
}

func newQueueItem(msg Message, idFrom uint32) *queueItem {
	return &queueItem{msg, idFrom}
}

type QueueMessages struct {
	arr []queueItem
}

func NewQueueMessages() *QueueMessages {
	ptr := new(QueueMessages)
	ptr.arr = make([]queueItem, 0)
	return ptr
}

func (queue *QueueMessages) Size() int {
	return len(queue.arr)
}

func (queue *QueueMessages) Push(msg Message, idFrom uint32) {
	if len(queue.arr) == cap(queue.arr) {
		queue.reallocate()
	}

	queue.arr = append(queue.arr, *newQueueItem(msg, idFrom))
}

func (queue *QueueMessages) reallocate() {
	if len(queue.arr) < cap(queue.arr) {
		return
	}

	newPtr := make([]queueItem, len(queue.arr), (len(queue.arr)+1)*2)
	copy(newPtr, queue.arr)
	queue.arr = newPtr
}

func (queue *QueueMessages) Pop() (Message, uint32) {
	firstElem := queue.arr[0]
	queue.arr = append(queue.arr[:0], queue.arr[1:]...)
	return firstElem.msg, firstElem.id
}
