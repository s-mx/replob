package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Broadcaster interface {
	Broadcast(*cont.Message, nodes.NodeId)
}

type SimpleBroadcaster struct {
}

type MyMainBroadcaster struct {
	queue cont.QueueMessages
}

func (broadcaster *MyMainBroadcaster) addMessage(msg *cont.Message, id nodes.NodeId) {
	broadcaster.queue.Push(msg, uint32(id))
}

type MyBroadcaster struct {
	info            *nodes.NodesInfo
	mainBroadcaster *MyMainBroadcaster
}

func (broadcaster *MyBroadcaster) addMessage(msg *cont.Message, idFrom nodes.NodeId) {
	broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}

func (broadcaster *MyBroadcaster) Broadcast(msg *cont.Message, idFrom nodes.NodeId) {
	broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}
