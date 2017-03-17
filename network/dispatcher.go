package network

import (
	consensus "github.com/s-mx/replob/consensuser"
	cont    "github.com/s-mx/replob/containers"
	"log"
	"sync"
	"time"
)

type NetworkDispatcher struct {
	consensus.Dispatcher
	id             			int
	config         			Configuration
	ServerService			*ServerService
	ClientServices			[]*ClientService

	cons					consensus.Consensuser

	myStepId				cont.StepId
	myStamp					cont.Stamp
	nodesStamps				[]cont.Stamp

	channelLoopEnd			chan interface{}
	channelStop				chan interface{}

	// Есть функция Stop, которую может позвать и консэнсусер и пользователь диспетчера
	// Для исключительного доступа используется мьютекс
	isRunning     			bool
	mutexRunning			sync.Mutex

	waitGroup				sync.WaitGroup
}

func NewNetworkDispatcher(id int, config *Configuration) *NetworkDispatcher {
	ptr := &NetworkDispatcher{
		id:id,
		config:*config,
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

		ptr.ClientServices[ind] = NewClientService(id, config.serviceServer[ind])
	}

	ptr.ServerService = NewServerService(id, config)
	return ptr
}

func (dispatcher *NetworkDispatcher) StartClients() {
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		dispatcher.ClientServices[ind].Start()
	}
}

func (dispatcher *NetworkDispatcher) Loop() {
	defer func() {
		log.Printf("Dispatcher [%d]: stop working", dispatcher.id)
		dispatcher.waitGroup.Done()
	}()

	for {
		select {
		case <-dispatcher.channelStop:
			return
		default:
		}

		select {
		case message := <-dispatcher.ServerService.channelMessage: // FIXME: select
			dispatcher.OnReceive(message)
		case <-time.After(100*time.Millisecond):
		}

		// TODO:
		// Может ли быть дедлок?
	}
}

func (dispatcher *NetworkDispatcher) goLoop() {
	dispatcher.waitGroup.Add(1)
	go dispatcher.Loop()
}

func (dispatcher *NetworkDispatcher) Start() {
	defer dispatcher.mutexRunning.Unlock()
	dispatcher.mutexRunning.Lock()

	if dispatcher.cons == nil {
		log.Panicf("ERROR dispatcher[%d]: consensuser isn't created\n", dispatcher.id)
	}

	if dispatcher.isRunning {
		return
	}

	dispatcher.StartClients()
	dispatcher.ServerService.Start()

	dispatcher.isRunning = true
	dispatcher.goLoop()
}

func (dispatcher *NetworkDispatcher) Stop() {
	defer dispatcher.mutexRunning.Unlock() // FIXME: заменить на channel
	dispatcher.mutexRunning.Lock()
	if dispatcher.isRunning == false {
		log.Panicf("Dispatcher isn't working") // Консэнсусер может сюда войти
	}

	dispatcher.channelStop<-0
	dispatcher.waitGroup.Wait()

	dispatcher.ServerService.Stop()
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		dispatcher.ClientServices[ind].Stop()
	}

	dispatcher.isRunning = false
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

		// FIXME: in case of blocking just drop the message
		// In future we need something else for that.
		// May be set high buffer for channel.
		dispatcher.ClientServices[ind].channelMessage<-message
	}
}

func (dispatcher *NetworkDispatcher) IncStep() {
	dispatcher.myStepId += 1
}

func (dispatcher *NetworkDispatcher) nextStamp() cont.Stamp {
	dispatcher.myStamp += 1
	return dispatcher.myStamp
}

func (dispatcher *NetworkDispatcher) IsRunning() bool {
	defer dispatcher.mutexRunning.Unlock()
	dispatcher.mutexRunning.Lock()
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
