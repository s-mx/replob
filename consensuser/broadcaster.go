package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Broadcaster interface {
	Broadcast(cont.Message, nodes.NodeId)
}

type SimpleBroadcaster struct {
    info nodes.NodesInfo
    queues []cont.QueueMessages
}

func NewSimpleBroadcaster(info nodes.NodesInfo) *SimpleBroadcaster {
	return &SimpleBroadcaster{info:info, queues:make([]cont.QueueMessages, info.Size())}
}

func (broadcaster *SimpleBroadcaster) Broadcast(msg cont.Message, idFrom nodes.NodeId) {
	for ind := 0; uint32(ind) < broadcaster.info.Size(); ind++ {
        if ind != int(idFrom) {
            broadcaster.queues[ind].Push(&msg, uint32(idFrom))
        }
    }
}

func (broadcaster *SimpleBroadcaster) proceedMessage(idFrom int, cons *MyConsensuser) {
    if broadcaster.queues[idFrom].Size() == 0 {
        return
    }

    msg, id := broadcaster.queues[idFrom].Pop()
    cons.OnBroadcast(msg, nodes.NodeId(id))
}