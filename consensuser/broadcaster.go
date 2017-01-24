package consensuser

import (
	"errors"
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Broadcaster interface {
	Broadcast(cont.Message)
}

type SimpleBroadcaster struct {
    info nodes.NodesInfo
    queues []cont.QueueMessages
}

func NewSimpleBroadcaster(info nodes.NodesInfo) *SimpleBroadcaster {
	return &SimpleBroadcaster{info:info, queues:make([]cont.QueueMessages, info.Size())}
}

func (broadcaster *SimpleBroadcaster) Broadcast(msg cont.Message) {
	for ind := 0; uint32(ind) < broadcaster.info.Size(); ind++ {
        if ind != int(msg.IdFrom) {
            broadcaster.queues[ind].Push(&msg)
        }
    }
}

func (broadcaster *SimpleBroadcaster) proceedMessage(cons *MyConsensuser) error {
    if broadcaster.queues[cons.Id].Size() == 0 {
        return errors.New("The queue is empty")
    }

    msg := broadcaster.queues[cons.Id].Pop()
    cons.OnVote(msg)
	return nil
}