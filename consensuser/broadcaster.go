package consensuser

import (
	"errors"
	cont "github.com/s-mx/replob/containers"
)

type Broadcaster interface {
	Broadcast(cont.Message)
}

type TestLocalBroadcaster struct {
    info cont.Set
    queues []cont.QueueMessages
}

func NewTestLocalBroadcaster(info cont.Set) *TestLocalBroadcaster {
	return &TestLocalBroadcaster{info:info, queues:make([]cont.QueueMessages, info.Size())}
}

func (broadcaster *TestLocalBroadcaster) Broadcast(msg cont.Message) {
	for ind := 0; uint32(ind) < broadcaster.info.Size(); ind++ {
        if ind != int(msg.IdFrom) {
            broadcaster.queues[ind].Push(msg)
        }
    }
}

// FIXME: use interface instead of concrete implementation
func (broadcaster *TestLocalBroadcaster) proceedMessage(cons *CalmConsensuser) error {
    if broadcaster.queues[cons.Id].Size() == 0 {
        return errors.New("The queue is empty")
    }

    msg := broadcaster.queues[cons.Id].Pop()
	cons.OnBroadcast(msg)
	return nil
}
