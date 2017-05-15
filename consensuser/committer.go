package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Committer interface {
	CommitSet(set cont.Carry)
}

type testCommitHelper struct {
	arrNodeCarries [][]cont.ElementaryCarry
	DesiredCarries []cont.ElementaryCarry
	arrDispatchers []*TestLocalDispatcher
}

func newTestCommitHelper(numberNodes int,
						 desiredCarries []cont.ElementaryCarry,
	                     arrDispatchers []*TestLocalDispatcher) *testCommitHelper {
	return &testCommitHelper{
		arrNodeCarries: make([][]cont.ElementaryCarry, numberNodes),
		DesiredCarries: desiredCarries,
		arrDispatchers: arrDispatchers,
	}
}

func (helper *testCommitHelper) findIndLastCommit(lastLength int) int {
	for ind := 0; ind < len(helper.arrNodeCarries); ind++ {
		if helper.isRunning(ind) && lastLength == len(helper.arrNodeCarries[ind]) {
			return ind
		}
	}

	return -1
}

func (helper *testCommitHelper) isRunning(ind int) bool {
	return helper.arrDispatchers[ind].IsRunning()
}

func (helper *testCommitHelper) getCommonLength(ind int) int {
	a := len(helper.DesiredCarries)
	b := len(helper.arrNodeCarries[ind])
	if a < b {
		return a
	} else {
		return b
	}
}

func (helper *testCommitHelper) CheckSafety() bool {
	for indNode := 0; indNode < len(helper.arrNodeCarries); indNode++ {
		if helper.arrDispatchers[indNode].IsRunning() &&
		   len(helper.arrNodeCarries[indNode]) != len(helper.DesiredCarries) {
			return false
		}

		commonLength := helper.getCommonLength(indNode)
		for indCarry := 0; indCarry < commonLength; indCarry++ {
			if helper.arrNodeCarries[indNode][indCarry].NotEqual(&helper.DesiredCarries[indCarry]) {
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

func (committer *TestLocalCommiter) Commit(elemCarry cont.ElementaryCarry) {
	curCarries := &committer.ptrAllCarries.arrNodeCarries[committer.idNode]
	*curCarries = append(*curCarries, elemCarry)
}

func (committer *TestLocalCommiter) CommitSet(carry cont.Carry) {
	sizeSet := carry.Size()
	for ind := 0; ind < sizeSet; ind++ {
		elem, _ := carry.Get(ind)
		log.Printf("     Carry %d committied", elem.GetId())
		committer.Commit(*elem)
	}
}

func (committer *TestLocalCommiter) CheckLastCarry(id int, elemCarry cont.ElementaryCarry) bool {
	carries := committer.ptrAllCarries.arrNodeCarries[id]
	lastInd := len(carries) - 1
	if lastInd < 0 {
		return false
	}

	return carries[lastInd].Equal(&elemCarry)
}
