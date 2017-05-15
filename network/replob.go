package network

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Replob interface {
	CommitSet(id cont.StepId, set cont.Carry) // TODO: Are we need this?
	Propose(cont.Carry)
	GetSnapshot(lastStepId cont.StepId, curStepId cont.StepId) (cont.Carry, bool)
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

func (replob *LocalReplob) CommitSet(stepId cont.StepId, carries cont.Carry) {
	for ind := 0; ind < carries.Size(); ind++ {
		carry, _ := carries.Get(ind)
		// TODO: разобраться с этим
		// replob.storage.Commit(*carry, stepId)
		log.Printf("Committed carry %d", carry.GetId()) // FIXME:
	}
}

func (replob *LocalReplob) Propose(carry cont.Carry) { // TODO: here something else against int
	replob.disp.Propose(carry)
	replob.counter++
}

func (replob *LocalReplob) getSnapshot(curStepId cont.StepId, lastStepId cont.StepId) (cont.Carry, bool) {
	return replob.storage.GetSnapshot(curStepId, lastStepId+1)
}
