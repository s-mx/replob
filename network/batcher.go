package network

import (
	cont "github.com/s-mx/replob/containers"
	"sync"
)

type Batcher struct {
	batchSize	int
	carry 		cont.Carry
	channel		chan cont.Carry
	dispatcher	*NetworkDispatcher

	mutex		sync.Mutex
}

func NewBatcher(dispatcher *NetworkDispatcher) *Batcher {
	return &Batcher{
		batchSize:10, // FIXME: use flags here
		carry:cont.NewCarry([]cont.ElementaryCarry{}),
		dispatcher:dispatcher,
	}
}

func (batcher *Batcher) Propose(carry cont.Carry) {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()

	for _, elemCarry := range carry.GetElementaryCarries() {
		batcher.carry.Append(elemCarry) // FIXME: добавлять до batchSize
	}

	if batcher.dispatcher.canPropose() { // TODO: потестить
		batcher.dispatcher.Propose(batcher.carry)
		batcher.carry = cont.NewCarry([]cont.ElementaryCarry{})
	}
}

func (batcher *Batcher) IsEmpty() bool {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()
	return batcher.carry.Size() == 0
}

func (batcher *Batcher) hasBatch() bool {
	return batcher.IsEmpty() == false
}

func (batcher *Batcher) popBatch() (cont.Carry, bool) {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()

	if batcher.IsEmpty() {
		return cont.NewCarry([]cont.ElementaryCarry{}), false
	}

	carry := batcher.carry
	batcher.carry = cont.NewCarry([]cont.ElementaryCarry{})
	return carry, true
}
