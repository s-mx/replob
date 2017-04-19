package network

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Replob interface {
	CommitSet(id cont.StepId, set cont.CarriesSet)
	Propose(value int)
	getCarry() (cont.Carry, bool)
	getSnapshot(lastStepId cont.StepId, curStepId cont.StepId) (cont.CarriesSet, bool)
}

type element struct {
	carry	cont.Carry
	stepId	cont.StepId
}

type Storage struct {
	queue	cont.Queue
}

func NewStorage() *Storage {
	return &Storage{}
}

func (storage *Storage) Commit(carry cont.Carry, id cont.StepId) {
	storage.queue.Push(element{carry:carry, stepId:id})
}

func (storage *Storage) checkIndexes(begin int, end int) bool {
	if begin > end || begin < 0 || storage.queue.Empty() {
		return false
	}

	if storage.queue.Back().(element).stepId < cont.StepId(begin) {
		return false
	}

	return false
}

func binsearch(arr cont.Queue, elem cont.StepId) int {
	left := 0
	right := arr.Size()
	for ; left + 1 < right; left++ {
		mid := (left + right) + 1
		if arr.Get(mid).(element).stepId <= elem {
			left = mid
		} else {
			right = mid
		}
	}

	return left
}

func (storage *Storage) findRange(beginStep cont.StepId, endStep cont.StepId) (int, int) {
	begin := binsearch(storage.queue, beginStep)
	end := binsearch(storage.queue, endStep-1)
	return begin, end
}

func (storage *Storage) GetSnapshot(begin cont.StepId, end cont.StepId) (cont.CarriesSet, bool) {
	if storage.checkIndexes(int(begin), int(end)) == false {
		return cont.CarriesSet{}, false
	}

	resultSet := cont.CarriesSet{}
	beginInd, endInd := storage.findRange(begin, end)
	for ; beginInd < endInd; beginInd++ {
		resultSet.ArrCarry = append(resultSet.ArrCarry, storage.queue.Get(int(beginInd)).(cont.Carry))
	}

	return cont.CarriesSet{}, true
}

type LocalReplob struct {
	counter		int
	batcher		*Batcher
	storage		*Storage
}

func NewLocalReplob() *LocalReplob {
	return &LocalReplob{
		counter:0,
		batcher:NewBatcher(),
		storage:NewStorage(),
	}
}

func (replob *LocalReplob) CommitSet(stepId cont.StepId, carries cont.CarriesSet) {
	for ind := 0; ind < carries.Size(); ind++ {
		replob.storage.Commit(carries.Get(ind), stepId)
		log.Printf("Committed carry %d", carries.Get(ind).id)
	}
}

func (replob *LocalReplob) Propose(value int) { // TODO: here something else against int
	replob.batcher.Propose(cont.NewElementaryCarry(replob.counter, cont.Payload(value)))
	replob.counter++
}

func (replob *LocalReplob) getCarry() (cont.Carry, bool) {
	if replob.batcher.IsEmpty() {
		return cont.Carry{}, false
	}

	carry := replob.batcher.GetCarry()
	return carry, true
}

func (replob *LocalReplob) getSnapshot(curStepId cont.StepId, lastStepId cont.StepId) (cont.CarriesSet, bool) {
	return replob.storage.GetSnapshot(curStepId, lastStepId+1)
}

type Batcher struct {
	counter		int
	batchSize	int
	queue		*cont.QueueCarry
}

func NewBatcher() *Batcher {
	return &Batcher{
		counter:0,
		batchSize:10, // FIXME: use flags here
		queue:cont.NewQueueCarry(),
	}
}

func (batcher *Batcher) Propose(carry cont.ElementaryCarry) {
	if batcher.queue.Empty() {
		batcher.queue.Push(cont.NewCarry(batcher.counter, carry))
		batcher.counter++
		return
	} else {
		lastCarry := batcher.queue.Back().(cont.Carry)
		if lastCarry.Size() < batcher.batchSize {
			lastCarry.Append(carry)
		}
	}
}

func (batcher *Batcher) IsEmpty() bool {
	return batcher.queue.Size() == 0
}

func (batcher *Batcher) GetCarry() cont.Carry {
	return batcher.queue.Pop()
}
