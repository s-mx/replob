package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Committer interface {
	CommitSet(set cont.CarriesSet)
}

type testCommitHelper struct {
	arrNodeCarries [][]cont.Carry
	DesiredCarries []cont.Carry
	arrDispatchers []*TestLocalDispatcher
}

func newTestCommitHelper(numberNodes int,
						 desiredCarries []cont.Carry,
	                     arrDispatchers []*TestLocalDispatcher) *testCommitHelper {
	return &testCommitHelper{
		arrNodeCarries: make([][]cont.Carry, numberNodes),
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
	// Кажется, в Go нету функции минимума для целых чисел
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
			if helper.arrNodeCarries[indNode][indCarry].NotEqual(helper.DesiredCarries[indCarry]) {
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
	curCarries := &committer.ptrAllCarries.arrNodeCarries[committer.idNode]
	*curCarries = append(*curCarries, carry)
}

func (committer *TestLocalCommiter) CommitSet(set cont.CarriesSet) {
	sizeSet := set.Size()
	for ind := 0; ind < sizeSet; ind++ {
		carry := set.Get(ind)
		log.Printf("     Carry %d committied", carry.Id)
		committer.Commit(carry)
	}
}

func (committer *TestLocalCommiter) CheckLastCarry(id int, carry cont.Carry) bool {
	carries := committer.ptrAllCarries.arrNodeCarries[id]
	lastInd := len(carries) - 1
	if lastInd < 0 {
		return false
	}

	return carries[lastInd].Equal(carry)
}
