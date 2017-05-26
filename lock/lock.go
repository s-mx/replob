package lock

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/network"
	"log"
	"time"
	"errors"
)

type action struct {
	TypeAction string
	Message    string
}

const (
	OK			= iota
	FAIL
	LOCKED_YOU
	LOCKED_OTHER
	NOT_LOCKED
	TIMEOUT
)

const DURATION_TIME = time.Second

type lockRecord struct {
	clientId  string
	timestamp time.Time
}

func (record *lockRecord) isEarlier(timestamp time.Time) bool {
	return record.timestamp.Before(timestamp)
}

func (record *lockRecord) isExpired() bool {
	return record.timestamp.Before(time.Now())
}

func (record *lockRecord) updateLease(timestamp time.Time) {
	record.timestamp = timestamp
}

func newLockRecord(message *Message) *lockRecord {
	return &lockRecord{
		clientId:  message.ClientId,
		timestamp: time.Now().Add(message.Duration),
	}
}

type lockReplob struct {
	dispatcher    *network.NetworkDispatcher
	actionChannels map[string]chan action	// GC не может трогать его. Нужно вызывать lock.close() !!!

	lockTable map[string]lockRecord
}

func newLockReplob() *lockReplob {
	return &lockReplob{
		actionChannels:make(map[string]chan action),
		lockTable:make(map[string]lockRecord),
	}
}

func (replob *lockReplob) ExecuteAcquireLock(message *Message) (resultCode int) {
	record, ok := replob.lockTable[message.LockId]
	endTimestamp := time.Now().Add(message.Duration)
	if ok {
		if record.clientId == message.ClientId && record.isEarlier(endTimestamp) {
			record.updateLease(endTimestamp)
			return OK
		}

		return FAIL
	}

	replob.lockTable[message.LockId] = *newLockRecord(message)
	return OK
}

func (replob *lockReplob) ExecuteUnlock(message *Message) (resultCode int) {
	record, ok := replob.lockTable[message.LockId]
	if ok {
		if expired := record.isExpired(); expired {
			delete(replob.lockTable, message.LockId)
			return NOT_LOCKED
		}

		if record.clientId == message.ClientId {
			delete(replob.lockTable, message.LockId)
			return OK
		} else {
			return LOCKED_OTHER
		}
	}

	return NOT_LOCKED
}

func (replob *lockReplob) ExecuteCarry(carry cont.ElementaryCarry) {
	message := Unmarshall(carry)
	if message.TypeMessage == "lock" {
		res := replob.ExecuteAcquireLock(message)
		actionChan, ok := replob.actionChannels[message.ClientId]
		if ok {
			var act action
			if res == OK {
				act = action{TypeAction:"lock", Message:message.LockId}
			} else {
				act = action{TypeAction:"lock", Message:"-1"} // костыль здесь :(
			}

			actionChan <- act
		}
	} else if message.TypeMessage == "unlock" {
		res := replob.ExecuteUnlock(message)
		actionChan, ok := replob.actionChannels[message.ClientId]
		if ok {
			var act action
			if res == OK {
				act = action{TypeAction:"unlock"}
			} else if res == LOCKED_OTHER {
				act = action{TypeAction:"error", Message:"LOCKED_OTHER"}
			} else if res == NOT_LOCKED {
				act = action{TypeAction:"error", Message:"NOT_LOCKED"}
			}

			actionChan <- act
		}
	} else {
		log.Printf("INFO LOCK: wrong carry [%d], TypeMessage: %s", carry.GetId(), message.TypeMessage)
	}
}

func (replob *lockReplob) CommitSet(id cont.StepId, carry cont.Carry) {
	for _, elemCarry := range carry.GetElementaryCarries() {
		replob.ExecuteCarry(elemCarry)
	}
}

func (replob *lockReplob) Propose(value cont.Carry) {
	replob.dispatcher.Propose(value)
}

func (replob *lockReplob) ProposeWithClient(value cont.Carry, clientId string) chan action {
	actionChan, ok := replob.actionChannels[clientId]
	if !ok {
		actionChan = make(chan action)
		replob.actionChannels[clientId] = actionChan
	}

	replob.Propose(value)
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

func (replob *lockReplob) NewLock(clientId string) *Lock {
	return &Lock{
		clientId:clientId,
		impl:replob,
	}
}

func (lock *Lock) createCarry(message *Message) *cont.Carry {
	bytes := message.Marshall()
	messageCarry := MessageCarry{message.TypeMessage, bytes.Bytes()}
	elemCarry := cont.NewElementaryCarry(0, cont.Payload(messageCarry))
	return cont.NewCarry([]cont.ElementaryCarry{elemCarry})
}

func (lock *Lock) AcquireLock(lockId string) (int, error) {
	message := &Message{
		TypeMessage: "lock",
		LockId:      lockId,
		ClientId:    lock.clientId,
		Duration:    DURATION_TIME,
	}

	actionChan := lock.impl.ProposeWithClient(*lock.createCarry(message), lock.clientId)
	timeOutChan := time.After(DURATION_TIME) // FIXME: Configure here
	for {
		select {
		case action := <-actionChan:
			if action.TypeAction == "lock" && action.Message == lockId {
				return OK, nil
			} else {
				return LOCKED_OTHER, errors.New("False-lock happened")
			}

			break
		case <-timeOutChan:
			return TIMEOUT, errors.New("TimeOut happened") // FIXME: нужно отличать TimeOut от False-lock
		}
	}
}

func (lock *Lock) Unlock(lockId string) (int, error) {
	message := &Message{
		TypeMessage: "unlock",
		LockId:      lockId,
		ClientId:    lock.clientId,
		Duration:    DURATION_TIME,
	}

	actionChan := lock.impl.ProposeWithClient(*lock.createCarry(message), lock.clientId)
	timeOutChan := time.After(DURATION_TIME)
	for {
		select {
		case action := <-actionChan:
			if action.TypeAction == "unlock" {
				return OK, nil
			} else if action.TypeAction == "error" {
				if action.Message == "OTHER" {
					return LOCKED_OTHER, errors.New("The lock is acquired by other client")
				} else if action.Message == "NOT_LOCKED" {
					return NOT_LOCKED, errors.New("The lock isn't acquired")
				} else {
					log.Printf("LOCK ERROR: %s\n", action.Message)
				}
			}

		case <-timeOutChan:
			return TIMEOUT, errors.New("TimeOut happened")
		}
	}
}

func (lock *Lock) Close() {
	delete(lock.impl.actionChannels, lock.clientId)
}
