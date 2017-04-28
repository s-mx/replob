package network

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Replob interface {
	CommitSet(id cont.StepId, set cont.CarriesSet) // TODO: Are we need this?
	Propose(value int)
	getSnapshot(lastStepId cont.StepId, curStepId cont.StepId) (cont.CarriesSet, bool)
}

type element struct {
	carry	cont.Carry
	stepId	cont.StepId
}

type LocalReplob struct {
	counter		int
	disp		NetworkDispatcher
	storage		*Storage
}

func NewLocalReplob() *LocalReplob {
	return &LocalReplob{
		counter:0,
		storage:NewStorage(),
	}
}

func (replob *LocalReplob) CommitSet(stepId cont.StepId, carries cont.CarriesSet) {
	for ind := 0; ind < carries.Size(); ind++ {
		replob.storage.Commit(*carries.Get(ind), stepId)
		log.Printf("Committed carry %d", carries.Get(ind).GetId())
	}
}

func (replob *LocalReplob) Propose(value int) { // TODO: here something else against int
	replob.disp.ProposeElementaryCarry(cont.NewElementaryCarry(replob.counter, cont.NewSimpleInt(value)))
	replob.counter++
}

func (replob *LocalReplob) getSnapshot(curStepId cont.StepId, lastStepId cont.StepId) (cont.CarriesSet, bool) {
	return replob.storage.GetSnapshot(curStepId, lastStepId+1)
}
