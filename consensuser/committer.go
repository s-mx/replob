package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Committer interface {
	Commit(cont.Carry)
	CommitSet(cont.CarriesSet)
}

type MyCommitter struct {
}

func (commiter *MyCommitter) Commit(carrier cont.Carry) {

}

type SimpleCommitter struct {
	resultStatus []bool
}

func NewSimpleCommitter(info nodes.NodesInfo) *SimpleCommitter {
	ptr := new(SimpleCommitter)
	ptr.resultStatus = make([]bool, info.Size())
	return ptr
}

func (committer *SimpleCommitter) Commit(carry cont.Carry) {
	committer.resultStatus[carry.Id] = true
}

func (committer *SimpleCommitter) CommitSet(set cont.CarriesSet) {
	sizeSet := set.Size()
	for ind := 0; ind < sizeSet; ind++ {
		committer.Commit(set.Get(ind))
	}
}

func (committer *SimpleCommitter) GetStatus(ind int) bool {
	return committer.resultStatus[ind]
}
