package consensuser

import (
	cont "github.com/s-mx/replob/containers"
)

type Committer interface {
	CommitSet(cont.CarriesSet)
}

type testCommitHelper struct {
	Carries 		[][]cont.Carry
	DesiredCarries 	[]cont.Carry
	arrDispatchers	[]*TestLocalDispatcher
}

func newTestCommitHelper(numberNodes int,
						 desiredCarries []cont.Carry,
	                     arrDispatchers []*TestLocalDispatcher) *testCommitHelper {
	return &testCommitHelper{
		Carries:make([][]cont.Carry, numberNodes),
		DesiredCarries:desiredCarries,
		arrDispatchers:arrDispatchers,
	}
}

func (helper *testCommitHelper) isRunning(ind int) bool {
	return helper.arrDispatchers[ind].IsRunning()
}

func (helper *testCommitHelper) getCommonLength(ind int) int {
	a := len(helper.DesiredCarries)
	b := len(helper.Carries[ind])
	// Кажется, в Go нету функции минимума для целых чисел
	if a < b {
		return a
	} else {
		return b
	}
}

func (helper *testCommitHelper) CheckSafety() bool {
	for indNode := 0; indNode < len(helper.Carries); indNode++ {
		if helper.arrDispatchers[indNode].IsRunning() &&
		   len(helper.Carries[indNode]) != len(helper.DesiredCarries) {
			return false
		}

		commonLength := helper.getCommonLength(indNode)
		for indCarry := 0; indCarry < commonLength; indCarry++ {
			if helper.Carries[indNode][indCarry].NotEqual(helper.DesiredCarries[indCarry]) {
				return false
			}
		}
	}

	return true
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