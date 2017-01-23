package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Committer interface {
	Commit(cont.Carry)
	CommitSet(cont.CarriesSet)
	Broadcast(cont.Carry)
	BroadcastSet(set cont.CarriesSet)
}

type SimpleCommitter struct {
	Carries []cont.Carry
}

func NewSimpleCommitter(info nodes.NodesInfo) *SimpleCommitter {
	return &SimpleCommitter{
		Carries:make([]cont.Carry, info.Size()),
	}
}

func (committer *SimpleCommitter) Commit(carry cont.Carry) {
	committer.Carries = append(committer.Carries, carry)
}

func (committer *SimpleCommitter) CommitSet(set cont.CarriesSet) {
	sizeSet := set.Size()
	for ind := 0; ind < sizeSet; ind++ {
		committer.Commit(set.Get(ind))
	}
}

func (committer *SimpleCommitter) Broadcast(carry cont.Carry) {
	// пока что оставим так
}

func (committer *SimpleCommitter) BroadcastSet(carriesSet cont.CarriesSet) {
	// пока что оставим так
}

func (committer *SimpleCommitter) CheckLastCarry(carry *cont.Carry) bool {
	lastInd := len(committer.Carries) - 1
	if lastInd < 0 {
		return false
	}

	return committer.Carries[lastInd].Equal(*carry)
}