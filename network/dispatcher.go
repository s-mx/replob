package network

import (
	calmConsensus "github.com/s-mx/replob/consensuser"
	cont    "github.com/s-mx/replob/containers"
	"log"
)

type NetworkDispatcher struct {
	calmConsensus.Dispatcher
	id             			int
	config         			Configuration
	channelServer			chan string // FIXME: remove
	channelServerMessage	chan cont.Message // FIXME: remove
	ServerService			*ServerService
	ClientServices			[]*ClientService

	cons					*calmConsensus.CalmConsensuser

	myStepId				cont.StepId
	myStamp					cont.Stamp
	nodesStamps				[]cont.Stamp
	isRunning     			bool
}

// FIXME: use Consensuser instance (interface) here
func NewNetworkDispatcher(id int, config Configuration) *NetworkDispatcher {
	ptr := &NetworkDispatcher{
		id:id,
		config:config,
		ClientServices:make([]*ClientService, config.numberNodes),
		myStepId:0,
		myStamp:0,
		nodesStamps:make([]cont.Stamp, config.numberNodes),
		isRunning:false,
	}

	for ind := 0; ind < config.numberNodes; ind++ {
		if ind == id {
			continue
		}

		ptr.ClientServices[ind] = NewClientService(id, config.serviceClient[ind])
	}

	ptr.ServerService = NewServerService(id, config)
	return ptr
}

func (dispatcher *NetworkDispatcher) RunClients() {
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		dispatcher.ClientServices[ind].Start()
	}
}

// FIXME: Use consistent names, see services .Start()
func (dispatcher *NetworkDispatcher) Run() {
	if dispatcher.cons == nil {
		log.Panicf("ERROR dispatcher[%d]: consensuser isn't created\n", dispatcher.id)
	}

	dispatcher.RunClients()
	dispatcher.ServerService.Start()

	dispatcher.isRunning = true

	// FIXME: extract to another method: Loop()
	// FIXME: check for stopping
	for {
		message := <-dispatcher.ServerService.channelMessage
		dispatcher.OnReceive(message)
	}
}

func (dispatcher *NetworkDispatcher) messageIsOutdated(message cont.Message) bool {
	return dispatcher.nodesStamps[uint(message.IdFrom)] >= message.Stamp
}

func (dispatcher *NetworkDispatcher) updateMessageStamp(message cont.Message) {
	if dispatcher.messageIsOutdated(message) == false {
		dispatcher.myStamp = message.Stamp
	}
}

func (dispatcher *NetworkDispatcher) OnReceive(message cont.Message) {
	if dispatcher.isRunning == false {
		return
	}

	if message.StepId > dispatcher.myStepId {
		dispatcher.isRunning = false
		dispatcher.logOutdatedStepId(message)
		return
	}

	if dispatcher.messageIsOutdated(message) ||
		dispatcher.myStepId > message.StepId {
		// Message is outdated by stamp or by stepId
		dispatcher.logDroppedMessage(message)
		return
	}

	dispatcher.updateMessageStamp(message)
	dispatcher.cons.OnBroadcast(message)
}

func (dispatcher *NetworkDispatcher) Broadcast(message cont.Message) {
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		// FIXME: increase size of channel to avoid blocking
		// FIXME: in case of blocking just drop the message
		// In future we need something else for that.
		// May be set high buffer for channel.
		dispatcher.ClientServices[ind].channelMessage<-message
	}
}

func (dispatcher *NetworkDispatcher) IncStep() {
	dispatcher.myStepId += 1
}

func (dispatcher *NetworkDispatcher) Stop() {
	dispatcher.isRunning = false
	dispatcher.ServerService.Stop()
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		dispatcher.ClientServices[ind].Stop()
	}
}

func (dispatcher *NetworkDispatcher) nextStamp() cont.Stamp {
	dispatcher.myStamp += 1
	return dispatcher.myStamp
}

func (dispatcher *NetworkDispatcher) IsRunning() bool {
	return dispatcher.isRunning
}

func (dispatcher *NetworkDispatcher) logDroppedMessage(message cont.Message) {
	if dispatcher.messageIsOutdated(message) {
		log.Printf("INFO Dispatcher[%d]: Message dropped by stamp. Message stamp=%d, dispatcher stamp=%d\n",
			dispatcher.id, message.Stamp, dispatcher.myStamp)
	} else if dispatcher.myStepId > message.StepId {
		log.Printf("INFO Dispatcher[%d]: Message dropped by StepId. Message StepId=%d, dispatcher StepId=%d\n",
			dispatcher.id, message.StepId, dispatcher.myStepId)
	}
}

func (dispatcher *NetworkDispatcher) logOutdatedStepId(message cont.Message) {
	log.Printf("WARNING: StepId of dispatcher is outdated: Message StepId=%d, dispatcher StepId=%d\n", message.StepId, dispatcher.myStepId)
}
