package containers

type Queue struct {
	arr []interface{}
}

func NewQueue() *Queue {
	return &Queue{
		arr:make([]interface{}, 0),
	}
}

func (q *Queue) Size() int {
	return len(q.arr)
}

func (q *Queue) Push(elem interface{}) {
	if len(q.arr) == cap(q.arr) {
		q.reallocate()
	}

	q.arr = append(q.arr, elem)
}

func (q *Queue) reallocate() {
	if len(q.arr) < cap(q.arr) {
		return
	}

	newPtr := make([]interface{}, len(q.arr), (len(q.arr)+1)*2)
	copy(newPtr, q.arr)
	q.arr = newPtr
}

func (q *Queue) Pop() interface{} {
	firstElem := q.arr[0]
	q.arr = append(q.arr[:0], q.arr[1:]...)
	return firstElem
}

func (q *Queue) Swap(i, j int) {
	q.arr[i], q.arr[j] = q.arr[j], q.arr[i]
}

func (q *Queue) Clear() {
	q.arr = make([]interface{}, 0)
}

type QueueMessages struct {
	*Queue
}

func NewQueueMessages() *QueueMessages {
	return &QueueMessages{
		Queue:NewQueue(),
	}
}

func (queue *QueueMessages) Size() int {
	return queue.Queue.Size()
}

func (queue *QueueMessages) Push(msg Message) {
	queue.Queue.Push(msg)
}

func (queue *QueueMessages) Pop() Message {
	return queue.Queue.Pop().(Message)
}

func (queue *QueueMessages) Clear() {
	queue.Queue.Clear()
}

type QueueCarry struct {
	*Queue
}

func NewQueueCarry() *QueueCarry {
	return &QueueCarry{
		Queue:NewQueue(),
	}
}

func (queue *QueueCarry) Size() int {
	return queue.Queue.Size()
}

func (queue *QueueCarry) Push(carry Carry) {
	queue.Queue.Push(carry)
}

func (queue *QueueCarry) Pop() Carry {
	return queue.Queue.Pop().(Carry)
}

func (queue *QueueCarry) Clear() {
	queue.Queue.Clear()
}
