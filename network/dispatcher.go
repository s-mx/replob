package network

import (
	consensus "github.com/s-mx/replob/consensuser"
	cont    "github.com/s-mx/replob/containers"
	"log"
	"sync"
	"time"
	"sync/atomic"
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

	// Есть функция Pause, которую может позвать и консэнсусер и пользователь диспетчера
	// Для исключительного доступа используется мьютекс
	isRunning     			*int32
	mutexRunning			sync.Mutex

	waitGroup				sync.WaitGroup
}

func NewNetworkDispatcher(id int, config *Configuration) *NetworkDispatcher {
	tmp := int32(0)
	ptr := &NetworkDispatcher{
		id:id,
		config:*config,
		ClientServices:make([]*ClientService, config.numberNodes),
		myStepId:0,
		myStamp:0,
		nodesStamps:make([]cont.Stamp, config.numberNodes),
		isRunning:&tmp,
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

		if atomic.LoadInt32(dispatcher.isRunning) == STOPPED {
			//time.Sleep(500 * time.Microsecond)
			continue
		}

		select {
		case message := <-dispatcher.ServerService.channelMessage:
			dispatcher.OnReceive(message)
			break
		case <-time.After(100*time.Millisecond): // FIXME: use flags
		}
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

	if atomic.LoadInt32(dispatcher.isRunning) == RUNNING {
		return
	}

	dispatcher.StartClients()
	dispatcher.ServerService.Start()

	atomic.StoreInt32(dispatcher.isRunning, RUNNING)
	dispatcher.goLoop()
}

func (dispatcher *NetworkDispatcher) Pause() {
	atomic.StoreInt32(dispatcher.isRunning, STOPPED)
}

func (dispatcher *NetworkDispatcher) Stop() {
	// Операция сравнения и присваивания неразделимые операции
	dispatcher.mutexRunning.Lock()
	if atomic.LoadInt32(dispatcher.isRunning) == STOPPED {
		log.Printf("INFO dispatcher[%d]: Attempt to stop. Dispatcher has already stopped", dispatcher.id) // Консэнсусер может сюда войти
		return
	}

	atomic.StoreInt32(dispatcher.isRunning, STOPPED)
	dispatcher.mutexRunning.Unlock()

	dispatcher.ServerService.Stop()
	dispatcher.channelStop<-0
	dispatcher.waitGroup.Wait()
	for ind := 0; ind < dispatcher.config.numberNodes; ind++ {
		if ind == dispatcher.id {
			continue
		}

		dispatcher.ClientServices[ind].Stop()
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
	if atomic.LoadInt32(dispatcher.isRunning) == STOPPED {
		return
	}

	if message.StepId > dispatcher.myStepId {
		atomic.StoreInt32(dispatcher.isRunning, STOPPED)
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

		// In future we need something else for that.
		// May be set high buffer for channel.
		select {
		case dispatcher.ClientServices[ind].channelMessage<-message:
			break
		default:
		}
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
	return atomic.LoadInt32(dispatcher.isRunning) == RUNNING
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
