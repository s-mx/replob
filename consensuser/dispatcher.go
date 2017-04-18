package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
	"math/rand"
	"log"
)

const (
	LOSTMAJORITY = iota
)

type Dispatcher interface {
	CommitSet(carries cont.CarriesSet)
	Broadcast(message cont.Message)
	IncStep()
	Stop() bool
	StopWait()
	Fail(int)
	Propose(carry cont.Carry)
}

type TestLocalDispatcher struct {
	nodeId      cont.NodeId
	conf        Configuration
	cons        Consensuser
	myStepId    cont.StepId
	myStamp     cont.Stamp
	nodesStamps []cont.Stamp
	queues      []cont.QueueMessages
	dispatchers []*TestLocalDispatcher
	committer	Committer
	isRunning   bool

	t 				*testing.T
}

func NewLocalDispatchers(numberDispatchers int, conf Configuration, t *testing.T) []*TestLocalDispatcher {
	arrPtr := make([]*TestLocalDispatcher, numberDispatchers)
	for i := 0; i < numberDispatchers; i++ {
		arrPtr[i] = NewLocalDispatcher(cont.NodeId(i), conf, numberDispatchers, t)
		arrPtr[i].dispatchers[i] = arrPtr[i]
		for j := 0; j < i; j++ {
			arrPtr[i].dispatchers[j] = arrPtr[j]
			arrPtr[j].dispatchers[i] = arrPtr[i]
		}
	}

	return arrPtr
}

func NewLocalDispatcher(id cont.NodeId, conf Configuration, numberDispatchers int, t *testing.T) *TestLocalDispatcher {
	return &TestLocalDispatcher{
		nodeId:      id,
		conf:        conf,
		myStepId:    0,
		myStamp:     0,
		nodesStamps: make([]cont.Stamp, numberDispatchers),
		queues:      cont.NewQueueMessagesN(numberDispatchers),
		dispatchers: make([]*TestLocalDispatcher, numberDispatchers),
		isRunning:   true,
		t:           t,
	}
}

func (dispatcher *TestLocalDispatcher) Propose(carry cont.Carry) {
	if dispatcher.cons.GetState() == Initial {
		dispatcher.cons.Propose(carry)
	} else {
		log.Printf("Dispatcher [%d]: state of consenuser isn't Initial", dispatcher.nodeId)
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
	if dispatcher.isRunning == false {
		return
	}

	if message.StepId > dispatcher.myStepId {
		//dispatcher.isRunning = false // FIXME: продолжить работать, запросить прошлый коммит
		log.Printf("WARNING: StepId of dispatcher[%d] is outdated: Message StepId=%d, dispatcher StepId=%d\n", dispatcher.nodeId, message.StepId, dispatcher.myStepId)
		return
	}

	if dispatcher.messageIsOutdated(message) ||
	   dispatcher.myStepId > message.StepId {
		// Message is outdated by stamp or by stepId
		return
	}

	dispatcher.updateMessageStamp(message)
	dispatcher.cons.OnBroadcast(message)
}

func (dispatcher *TestLocalDispatcher) CommitSet(carries cont.CarriesSet) {
	// TODO: разделить операции по типу и занести в лог здесь
	dispatcher.committer.CommitSet(carries)
}

func (dispatcher *TestLocalDispatcher) IsRunning() bool {
	return dispatcher.isRunning
}

func (dispatcher *TestLocalDispatcher) Stop() bool {
	dispatcher.isRunning = false
	return true
}

func (dispatcher *TestLocalDispatcher) StopWait() {
	dispatcher.isRunning = false
}

func (dispatcher *TestLocalDispatcher) Fail(codeReason int) {
	// TODO: recovery here
	if codeReason == LOSTMAJORITY {
		log.Printf("Dispatcher [%d]: Lost majority in consensuser", dispatcher.nodeId)
	}

	dispatcher.isRunning = false
}

func (dispatcher *TestLocalDispatcher) proceedFirstMessage(toId int) {
	if dispatcher.queues[toId].Size() == 0 {
		dispatcher.t.Error("Empty message queue")
	}

	message := dispatcher.queues[toId].Pop()
	dispatcher.dispatchers[toId].OnReceive(message)
}

func (dispatcher *TestLocalDispatcher) proceedRandomMessage(generator *rand.Rand, probSwap float32) bool {
	if dispatcher.IsRunning() == false {
		return false
	}

	result := false
	for ind := 0; ind < int(dispatcher.conf.Size()); ind++ {
		if dispatcher.queues[ind].Size() == 0 {
			continue
		}

		if dispatcher.queues[ind].Size() >= 2 && generator.Float32() < probSwap {
			dispatcher.queues[ind].Swap(0, 1)
		}

		result = true
		message := dispatcher.queues[ind].Pop()
		dispatcher.dispatchers[ind].OnReceive(message)
	}

	return result
}

func (dispatcher *TestLocalDispatcher) ClearQueues() {
	for ind := 0; ind < len(dispatcher.queues); ind++ {
		dispatcher.queues[ind].Clear()
	}
}