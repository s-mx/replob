package lock

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/network"
	"time"
	"log"
)

type action struct {
	TypeAction string
	Message    string
}

type lockReplob struct {
	dispatcher		*network.NetworkDispatcher
	actionChannel	chan action

	timeStamps		map[string]map[string]string
}

func newLockReplob() *lockReplob {
	return &lockReplob{}
}

func (replob *lockReplob) ExecuteLock(message *Message) {
	table, ok := replob.timeStamps[message.clienId]
}

func (replob *lockReplob) ExecuteUnlock(message *Message) {

}

func (replob *lockReplob) ExecuteIsLocked(message *Message) {

}

func (replob *lockReplob) ExecuteCarry(carry cont.ElementaryCarry) {
	message := Unmarshall(carry)
	if message.typeMessage == "lock" {
		replob.ExecuteLock(message)
	} else if message.typeMessage == "unlock" {
		replob.ExecuteUnlock(message)
	} else if message.typeMessage == "isLocked" {
		replob.ExecuteIsLocked(message)
	} else {
		log.Printf("INFO LOCK: wrong carry [%d]", carry.GetId())
	}
}

func (replob *lockReplob) CommitSet(id cont.StepId, set cont.CarriesSet) {
	carry := set.ArrCarry[0]
	for _, elemCarry := range(carry.GetElementaryCarries()) {
		replob.ExecuteCarry(elemCarry)
	}
}

func (replob *lockReplob) Propose(value cont.Carry) {
	replob.dispatcher.Propose(value)
}

func (replob *lockReplob) GetSnapshot() (cont.CarriesSet, bool) {
	return cont.CarriesSet{}, true
}

type Lock struct {
	impl	*lockReplob
}

func NewLock() *Lock {
	return &Lock{}
}

func (lock *Lock) createCarry(lock_id string) cont.Carry {
	return cont.NewCarry([]cont.ElementaryCarry{})
}

func (lock *Lock) Lock(lockId string) (bool, error) {
	lock.impl.Propose(lock.createCarry(lockId))
	endTimeStamp := time.Now().Add(5 * time.Second)
	for {
		select {
		case action := <-lock.impl.actionChannel:
			if action.TypeAction == "lock" && action.Message == lockId {
				return true, nil
			}

			break
		case <-time.After(endTimeStamp.Sub(time.Now())):
			return false, nil
		}
	}
}

func (lock *Lock) Unlock() (bool, error) {
	return true, nil
}

func (lock *Lock) IsChecked() {

}