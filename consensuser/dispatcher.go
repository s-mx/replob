package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"errors"
)

type Dispatcher interface {
	Broadcast(message cont.Message)
	IncStep()
	Stop()
}

type TestLocalDispatcher struct {
	nodeId			cont.NodeId
	conf 			Configuration
	cons			*CalmConsensuser
	myStepId		cont.StepId
	myStamp			cont.Stamp
	nodesStamps		[]cont.Stamp
	queues			[]cont.QueueMessages
	dispatchers     []*TestLocalDispatcher
	isStopReceiving	bool
}

func NewLocalDispatchers(numberDispatchers int, conf Configuration) []*TestLocalDispatcher {
	arrPtr := make([]*TestLocalDispatcher, numberDispatchers)
	for i := 0; i < numberDispatchers; i++ {
		arrPtr[i] = NewLocalDispatcher(cont.NodeId(i), conf, numberDispatchers)
		for j := 0; j < i; j++ {
			arrPtr[i].dispatchers[j] = arrPtr[j]
			arrPtr[j].dispatchers[i] = arrPtr[i]
		}
	}

	return arrPtr
}

func NewLocalDispatcher(id cont.NodeId, conf Configuration, numberDispatchers int ) *TestLocalDispatcher {
	return &TestLocalDispatcher{
		nodeId:id,
		conf:conf,
		myStepId:0,
		myStamp:0,
		nodesStamps:make([]cont.Stamp, numberDispatchers),
		queues:make([]cont.QueueMessages, numberDispatchers),
		dispatchers:make([]*TestLocalDispatcher, numberDispatchers),
		isStopReceiving:false,
	}
}

func (dispatcher *TestLocalDispatcher) nextStamp() cont.Stamp {
	dispatcher.myStamp += 1
	return dispatcher.myStamp
}

func (dispatcher *TestLocalDispatcher) getStep() cont.StepId {
	return dispatcher.myStepId
}

func (dispatcher *TestLocalDispatcher) Broadcast(message cont.Message) {
	message.IdFrom = dispatcher.nodeId
	message.Stamp = dispatcher.nextStamp()
	message.StepId = dispatcher.getStep()

	for ind := 0; uint(ind) < dispatcher.conf.Size(); ind++ {
		if ind != int(dispatcher.nodeId) {
			dispatcher.queues[ind].Push(message)
		}
	}
}

func (dispatcher *TestLocalDispatcher) messageIsOutdated(message cont.Message) bool {
	return dispatcher.nodesStamps[uint(message.IdFrom)] >= message.Stamp
}

func (dispatcher *TestLocalDispatcher) updateMessageStamp(message cont.Message) {
	if dispatcher.messageIsOutdated(message) == false {
		dispatcher.myStamp = message.Stamp
	}
}

func (dispatcher *TestLocalDispatcher) IncStep() {
	dispatcher.myStepId += 1
}

func (dispatcher *TestLocalDispatcher) OnReceive(message cont.Message) {
	if dispatcher.isStopReceiving {
		return
	}

	if message.StepId > dispatcher.myStepId {
		dispatcher.isStopReceiving = true
		return
	}

	if dispatcher.messageIsOutdated(message) ||
	   dispatcher.myStepId > message.StepId {
		return
	}

	dispatcher.updateMessageStamp(message)
	dispatcher.cons.OnBroadcast(message)
}

func (dispatcher *TestLocalDispatcher) Stop() {
	dispatcher.isStopReceiving = true
}

func (dispatcher *TestLocalDispatcher) proceedFirstMessage(toId int) error {
	if dispatcher.queues[toId].Size() == 0 {
		return errors.New("Empty message queue")
	}

	message := dispatcher.queues[toId].Pop()
	dispatcher.dispatchers[toId].OnReceive(message)
	return nil
}