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

	replob					Replob
	batcher					*Batcher

	channelStop				chan interface{}

	isRunning     			*int32
	mutexRunning			sync.Mutex

	waitGroup				sync.WaitGroup
}

func NewNetworkDispatcher(id int, config *Configuration, reblob Replob) *NetworkDispatcher {
	tmp := int32(0)
	ptr := &NetworkDispatcher{
		id:id,
		config:*config,
		ServerService:NewServerService(id, config),
		ClientServices:make([]*ClientService, config.numberNodes),
		myStepId:0,
		myStamp:0,
		nodesStamps:make([]cont.Stamp, config.numberNodes),
		channelStop:make(chan interface{}),
		isRunning:&tmp,
		replob:reblob,
	}

	ptr.batcher = NewBatcher(ptr)

	for ind := 0; ind < config.numberNodes; ind++ {
		if ind == id {
			continue
		}

		ptr.ClientServices[ind] = NewClientService(id, config.serviceServer[ind])
	}

	ptr.ServerService = NewServerService(id, config)
	return ptr
}

func NewConsensuser(id int, config* Configuration, replob Replob) (*NetworkDispatcher, *consensus.CalmConsensuser) {
	disp := NewNetworkDispatcher(id, config, replob)
	cons := consensus.NewCalmConsensuser(disp, config.GetMasterlessConfiguration(), id)
	disp.cons = cons
	return disp, cons
}

func (dispatcher *NetworkDispatcher) Propose(carry cont.Carry) {
	dispatcher.batcher.Propose(carry)
}

func (dispatcher *NetworkDispatcher) canPropose() bool {
	return dispatcher.cons.GetState() == consensus.Initial
}

func (dispatcher *NetworkDispatcher) ProposeElementaryCarry(carry cont.ElementaryCarry) {
	dispatcher.batcher.Propose(*cont.NewCarry([]cont.ElementaryCarry{carry}))
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
		case _=<-dispatcher.channelStop:
			return
		default:
		}

		if atomic.LoadInt32(dispatcher.isRunning) == STOPPED {
			time.Sleep(500 * time.Microsecond)
			continue
		}

		select {
		case message := <-dispatcher.ServerService.channelMessage:
			dispatcher.OnReceive(message)
			break
		case carry := <-dispatcher.batcher.channel:
			dispatcher.cons.Propose(carry)
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

func (dispatcher *NetworkDispatcher) Fail(codeReason int) {
	// TODO: recovery here
	if codeReason == consensus.LOSTMAJORITY {
		log.Printf("Dispatcher [%d]: Lost majority in consensuser", dispatcher.id)
	}
}

func (dispatcher *NetworkDispatcher) Stop() bool {
	if atomic.CompareAndSwapInt32(dispatcher.isRunning, RUNNING, STOPPED) == false {
		return true
	}

	return false
}

func (dispatcher *NetworkDispatcher) StopWait() {
	defer dispatcher.mutexRunning.Unlock()
	dispatcher.mutexRunning.Lock()
	if atomic.CompareAndSwapInt32(dispatcher.isRunning, RUNNING, STOPPED) == false {
		log.Printf("INFO dispatcher[%d]: Attempt to stop. Dispatcher has already stopped", dispatcher.id) // Консэнсусер может сюда войти
		return
	}

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

func (dispatcher *NetworkDispatcher) CommitSet(carries cont.Carry) {
	dispatcher.replob.CommitSet(dispatcher.myStepId, carries)

	if carry, ok := dispatcher.batcher.popBatch(); ok {
		dispatcher.Propose(*carry)
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
	message.IdFrom = cont.NodeId(dispatcher.id)
	message.Stamp = dispatcher.nextStamp()
	message.StepId = dispatcher.getStep()

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

func (dispatcher *NetworkDispatcher) getStep() cont.StepId {
	return dispatcher.myStepId
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
