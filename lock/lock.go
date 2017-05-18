package lock

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/network"
	"log"
	"time"
)

type action struct {
	TypeAction string
	Message    string
}

const (
	OK			= iota
	FAIL
	FREE
	LOCKED_YOU
	LOCKED_OTHER
)

/*
1) Нужно по lock_id быстро находить клиента.
2) Нужно
 */

type lockRecord struct {
	clientId	string
	timeStamp	time.Time
}

func (record *lockRecord) isEarlier(message *Message) bool {
	return record.timeStamp.Before(message.timeStamp)
}

func (record *lockRecord) isExpired() bool {
	return record.timeStamp.Before(time.Now())
}

func (record *lockRecord) updateLease(message *Message) {
	record.timeStamp = message.timeStamp
}

func newLockRecord(message *Message) *lockRecord {
	return &lockRecord{
		clientId:message.clientId,
		timeStamp:message.timeStamp,
	}
}

type lockReplob struct {
	dispatcher    *network.NetworkDispatcher
	actionChannels map[string]chan action

	lockTable map[string]lockRecord
}

func newLockReplob(dispatcher *network.NetworkDispatcher) *lockReplob {
	return &lockReplob{
		dispatcher:dispatcher,
		actionChannels:make(map[string]chan action),
		lockTable:make(map[string]lockRecord),
	}
}

func (replob *lockReplob) ExecuteAcquireLock(message *Message) (resultCode int) {
	record, ok := replob.lockTable[message.lockId]
	if ok {
		if record.clientId == message.clientId && record.isEarlier(message) {
			record.updateLease(message)
			return OK
		}

		return FAIL
	}

	replob.lockTable[message.lockId] = *newLockRecord(message)
	return OK
}

func (replob *lockReplob) ExecuteUnlock(message *Message) (resultCode int) {
	record, ok := replob.lockTable[message.lockId]
	if ok {
		if expired := record.isExpired(); expired {
			delete(replob.lockTable, message.lockId) // FIXME: consensus operation here
			return FAIL
		}

		delete(replob.lockTable, message.lockId)
		return OK
	}

	return FAIL
}

func (replob *lockReplob) ExecuteIsLocked(message *Message) (resultCode int) {
	record, ok := replob.lockTable[message.lockId]
	if ok {
		if expired := record.isExpired(); expired {
			delete(replob.lockTable, message.lockId) // FIXME: consensus operation here
			return FREE
		}

		if record.clientId == message.clientId {
			return LOCKED_YOU
		} else {
			return LOCKED_OTHER
		}
	}

	return FAIL
}

func (replob *lockReplob) ExecuteCarry(carry cont.ElementaryCarry) {
	message := Unmarshall(carry)
	if message.typeMessage == "lock" {
		replob.ExecuteAcquireLock(message) // TODO: send action through chan here
	} else if message.typeMessage == "unlock" {
		replob.ExecuteUnlock(message)
	} else if message.typeMessage == "isLocked" {
		replob.ExecuteIsLocked(message)
	} else {
		log.Printf("INFO LOCK: wrong carry [%d]", carry.GetId())
	}
}

func (replob *lockReplob) CommitSet(id cont.StepId, carry cont.Carry) {
	for _, elemCarry := range carry.GetElementaryCarries() {
		replob.ExecuteCarry(elemCarry)
	}
}

func (replob *lockReplob) Propose(value cont.Carry, clientId string) chan action {
	actionChan, ok := replob.actionChannels[clientId]
	if !ok {
		actionChan = make(chan action)
		replob.actionChannels[clientId] = actionChan

	}

	replob.dispatcher.Propose(value)
	return actionChan
}

func (replob *lockReplob) GetSnapshot() (cont.Carry, bool) {
	// TODO: IMPLEMENT
	return cont.Carry{}, true
}

type Lock struct {
	clientId	string
	impl		*lockReplob
}

func NewLock(clientId string, disp *network.NetworkDispatcher) *Lock {
	return &Lock{
		clientId:clientId,
		impl:newLockReplob(disp),
	}
}

func (lock *Lock) createCarry(lock_id string) *cont.Carry {
	return cont.NewCarry([]cont.ElementaryCarry{})
}

func (lock *Lock) AcquireLock(lockId string) (bool, error) {
	actionChan := lock.impl.Propose(*lock.createCarry(lockId), lock.clientId)
	timeOutChan := time.After(5 * time.Second)
	for {
		select {
		case action := <-actionChan:
			if action.TypeAction == "lock" && action.Message == lockId {
				return true, nil
			}

			break
		case <-timeOutChan:
			return false, nil // FIXME: нужно отличать TimeOut от False-lock
		}
	}
}

func (lock *Lock) Unlock() (bool, error) {
	return true, nil
}

func (lock *Lock) IsChecked() {

}
