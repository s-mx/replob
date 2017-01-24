package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Committer interface {
	Commit(int, cont.Carry)
	CommitSet(int, cont.CarriesSet)
	//Broadcast(cont.Carry)
	//BroadcastSet(set cont.CarriesSet)
}

type SimpleCommitter struct {
	Carries [][]cont.Carry
}

func NewSimpleCommitter(info nodes.NodesInfo) *SimpleCommitter {
	return &SimpleCommitter{
		Carries:make([][]cont.Carry, info.Size()),
	}
}

func (committer *SimpleCommitter) Commit(id int, carry cont.Carry) {
	committer.Carries[id] = append(committer.Carries[id], carry)
}

func (committer *SimpleCommitter) CommitSet(id int, set cont.CarriesSet) {
	sizeSet := set.Size()
	for ind := 0; ind < sizeSet; ind++ {
		committer.Commit(id, set.Get(ind))
	}
}

func (committer *SimpleCommitter) Broadcast(carry cont.Carry) {
	// пока что оставим так
}

func (committer *SimpleCommitter) BroadcastSet(carriesSet cont.CarriesSet) {
	// пока что оставим так
}

func (committer *SimpleCommitter) CheckLastCarry(id int, carry *cont.Carry) bool {
	lastInd := len(committer.Carries[id]) - 1
	if lastInd < 0 {
		return false
	}

	return committer.Carries[id][lastInd].Equal(*carry)
}