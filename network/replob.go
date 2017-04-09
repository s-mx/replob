package network

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/consensuser"
	"log"
)

type Replob interface {
	consensuser.Committer
	Propose(carry cont.Carry)
	getCarry() (cont.Carry, bool)
}

type LocalReplob struct {
	batcher		*Batcher
}

func NewLocalReplob() *LocalReplob {
	return &LocalReplob{
		batcher:NewBatcher(),
	}
}

func (replob *LocalReplob) CommitSet(carries cont.CarriesSet) {
	for ind := 0; ind < carries.Size(); ind++ {
		log.Printf("Committed carry %d", carries.Get(ind).Id)
	}
}

func (replob *LocalReplob) Propose(carry cont.Carry) {
	replob.batcher.Propose(carry)
}

func (replob *LocalReplob) getCarry() (cont.Carry, bool) {
	if replob.batcher.IsEmpty() {
		return nil, false
	}

	carry := replob.batcher.GetCarry()
	return carry, true
}

type Batcher struct {
	queue	cont.QueueCarry
}

func NewBatcher() *Batcher {
	return &Batcher{}
}

func (batcher *Batcher) Propose(carry cont.Carry) {
	batcher.queue.Push(carry)
}

func (batcher *Batcher) IsEmpty() bool {
	return batcher.queue.Size() == 0
}

func (batcher *Batcher) GetCarry() cont.Carry {
	return batcher.queue.Pop()
}
