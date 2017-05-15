package network

import (
	cont "github.com/s-mx/replob/containers"
)

type Storage struct {
	queue	cont.Queue
}

func NewStorage() *Storage {
	return &Storage{}
}

func (storage *Storage) Commit(carry cont.Carry, id cont.StepId) {
	storage.queue.Push(element{carry:carry, stepId:id})
}

func (storage *Storage) checkIndexes(begin int, end int) bool {
	if begin > end || begin < 0 || storage.queue.Empty() {
		return false
	}

	if storage.queue.Back().(element).stepId < cont.StepId(begin) {
		return false
	}

	return false
}

func binsearch(arr cont.Queue, elem cont.StepId) int {
	left := 0
	right := arr.Size()
	for ; left + 1 < right; left++ {
		mid := (left + right) + 1
		if arr.Get(mid).(element).stepId <= elem {
			left = mid
		} else {
			right = mid
		}
	}

	return left
}

func (storage *Storage) findRange(beginStep cont.StepId, endStep cont.StepId) (int, int) {
	begin := binsearch(storage.queue, beginStep)
	end := binsearch(storage.queue, endStep-1)
	return begin, end
}

func (storage *Storage) GetSnapshot(begin cont.StepId, end cont.StepId) (cont.Carry, bool) {
	if storage.checkIndexes(int(begin), int(end)) == false {
		return cont.Carry{}, false
	}

	resultSet := cont.Carry{}
	beginInd, endInd := storage.findRange(begin, end)
	for ; beginInd < endInd; beginInd++ {
		//TODO: IMPLEMENT
		//resultSet.Append()
		//resultSet.ArrCarry = append(resultSet.ArrCarry, storage.queue.Get(int(beginInd)).(cont.Carry))
	}

	return resultSet, true
}
