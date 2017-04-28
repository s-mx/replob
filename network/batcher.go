package network

import (
	cont "github.com/s-mx/replob/containers"
	"sync"
)

type Batcher struct {
	counter		int
	batchSize	int
	carry 		cont.Carry
	channel		chan cont.Carry

	mutex		sync.Mutex
}

func NewBatcher() *Batcher {
	return &Batcher{
		counter:0,
		batchSize:10, // FIXME: use flags here
		carry:cont.NewCarry(0, []cont.ElementaryCarry{}),
	}
}

func (batcher *Batcher) Propose(carry cont.ElementaryCarry) {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()

	batcher.carry.Append(carry)
}

func (batcher *Batcher) IsEmpty() bool {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()
	return batcher.carry.Size() == 0
}

func (batcher *Batcher) getCarry() (cont.Carry, bool) {
	batcher.mutex.Lock()
	defer batcher.mutex.Unlock()

	if batcher.IsEmpty() {
		return cont.NewCarry(-1, []cont.ElementaryCarry{}), false
	}

	carry := batcher.carry
	batcher.carry = cont.NewCarry(batcher.counter, []cont.ElementaryCarry{})
	batcher.counter++
	return carry, true
}
