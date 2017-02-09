package consensuser

import (
	cont "github.com/s-mx/replob/containers"
)

type Committer interface {
	Commit(cont.Carry)
	CommitSet(cont.CarriesSet)
}

type testCommitHelper struct {
	Carries [][]cont.Carry
}

func newTestCommitHelper(numberNodes int) *testCommitHelper {
	return &testCommitHelper{
		Carries:make([][]cont.Carry, numberNodes),
	}
}

type TestLocalCommiter struct {
	idNode  cont.NodeId
	ptrAllCarries *testCommitHelper
}

func NewTestLocalCommitter(idNode cont.NodeId, ptrHellper *testCommitHelper) *TestLocalCommiter {
	return &TestLocalCommiter{
		idNode:idNode,
		ptrAllCarries:ptrHellper,
	}
}

func (committer *TestLocalCommiter) Commit(carry cont.Carry) {
	curCarries := &committer.ptrAllCarries.Carries[committer.idNode]
	*curCarries = append(*curCarries, carry)
}

func (committer *TestLocalCommiter) CommitSet(set cont.CarriesSet) {
	sizeSet := set.Size()
	for ind := 0; ind < sizeSet; ind++ {
		committer.Commit(set.Get(ind))
	}
}

func (committer *TestLocalCommiter) CheckLastCarry(id int, carry cont.Carry) bool {
	carries := committer.ptrAllCarries.Carries[id]
	lastInd := len(carries) - 1
	if lastInd < 0 {
		return false
	}

	return carries[lastInd].Equal(carry)
}