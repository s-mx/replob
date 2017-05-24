package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"log"
)

type Consensuser interface {
	Propose(cont.Carry)
	OnBroadcast(cont.Message)
	OnDisconnect(cont.NodeId)
	GetState() CalmState
}

type ConsensusController interface {
	Stop()
	/*
	on LostSteps the following actions must be applied:
	1. Update commits based on currentStepId and latestStepId
	2. Add current node to the group
	 */
	LostSteps(currentStepId cont.StepId, latestStepId cont.StepId)
}

// states for consensus algorithm
type CalmState int

const (
	Initial CalmState = iota
	MayCommit
	CannotCommit
)

type CalmConsensuser struct {
	Dispatcher

	state        CalmState
	id           cont.NodeId // Только для логирования.
	nodes        cont.Set
	currentNodes cont.Set
	votedSet     cont.Set
	carry        *cont.Carry
}

func NewCalmConsensuser(dispatcher Dispatcher, conf Configuration, id int) *CalmConsensuser {
	return &CalmConsensuser{
		Dispatcher:		dispatcher,
		state:			Initial,
		id:				cont.NodeId(id),
		nodes:			conf.Info,
		currentNodes:	conf.Info,
		carry:			cont.NewCarry([]cont.ElementaryCarry{}),
	}
}

func (consensuser *CalmConsensuser) doBroadcast() {
	msg := cont.NewMessageVote(*consensuser.carry, consensuser.votedSet, consensuser.currentNodes)
	consensuser.Broadcast(msg)
}

func (consensuser *CalmConsensuser) newVote(carry cont.Carry) cont.Message {
	return cont.NewMessageVote(carry, consensuser.votedSet, consensuser.currentNodes)
}

// checks that all nodes are agreed on sequence of carries and nodes group
func (consensuser *CalmConsensuser) Propose(carry cont.Carry) {
	if consensuser.state != Initial {
		log.Fatalf("state of consenuser %d isn't Initial on propose", consensuser.id)
	}

	firstCarry, _ := carry.GetFirst()
	log.Printf("Consensuser [%d]: Propose %d", consensuser.id, firstCarry.GetId())
	consensuser.OnVote(consensuser.newVote(carry))
}

// FIXME: уточнить в названии функции смысл инварианта.
//        слово invariant употреблено не удачно, т. к. инвариант - это что-то, что выполняется всегда.
func (consensuser *CalmConsensuser) checkInvariant(msg cont.Message) bool {
	return consensuser.carry.Equal(msg.Carry) &&
		   consensuser.currentNodes.Equal(msg.NodeSet)
}

func (consensuser *CalmConsensuser) OnBroadcast(msg cont.Message) {
	if consensuser.currentNodes.Consist(uint32(msg.IdFrom)) == false {
		log.Printf("Consensuser [%d]: dropped message out of group. idFrom: %d", consensuser.id, msg.IdFrom)
		return
	}

	if msg.GetType() == cont.Vote {
		consensuser.OnVote(msg)
	} else if msg.GetType() == cont.Commit {
		consensuser.onCommit()
	}
}

func (consensuser *CalmConsensuser) mergeVotes(msg cont.Message) {
	consensuser.carry = consensuser.carry.Union(&msg.Carry)
	consensuser.currentNodes.Intersect(msg.NodeSet)
}

func (consensuser *CalmConsensuser) OnVote(msg cont.Message) {
	if consensuser.state == MayCommit && consensuser.checkInvariant(msg) == false {
		log.Printf("Consensuser [%d]: It's broken invariant", consensuser.id)
		consensuser.state = CannotCommit
	}

	consensuser.votedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.id)))
	consensuser.votedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
	consensuser.mergeVotes(msg) // don't use msg right after this line
	consensuser.votedSet.Intersect(consensuser.currentNodes)
	if consensuser.nodes.Size() >= consensuser.currentNodes.Size() * 2 {
		log.Printf("Consensuser [%d]: Current set of nodes has become less than majority", consensuser.id)
		consensuser.Fail(LOSTMAJORITY)
		return
	}

	if consensuser.state == Initial {
		consensuser.state = MayCommit
		consensuser.doBroadcast()
	}

	if consensuser.votedSet.Equal(consensuser.currentNodes) {
		if consensuser.state == MayCommit {
			consensuser.onCommit()
		} else {
			log.Printf("Consensuser [%d]: Consensuser state has become MayCommit", consensuser.id)
			consensuser.state = MayCommit
			consensuser.votedSet.Clear()
			consensuser.votedSet.Insert(uint32(consensuser.id))
			consensuser.doBroadcast()
		}
	}
}

func (consensuser *CalmConsensuser) onCommit() {
	log.Printf("Consensuser [%d] has committed:", consensuser.id)
	consensuser.CommitSet(*consensuser.carry)
	consensuser.Broadcast(consensuser.newCommitMessage())
	consensuser.prepareNextStep()
}

func (consensuser *CalmConsensuser) newCommitMessage() cont.Message {
	return *cont.NewMessageCommit(*consensuser.carry)
}

func (consensuser *CalmConsensuser) cleanUp() {
	consensuser.state = Initial
	consensuser.nodes = consensuser.currentNodes
	consensuser.carry.Clear()
	consensuser.votedSet.Clear()
}

func (consensuser *CalmConsensuser) prepareNextStep() {
	consensuser.cleanUp()
	consensuser.IncStep()
}

func (consensuser *CalmConsensuser) GetId() cont.NodeId {
	return consensuser.id
}

func (consensuser *CalmConsensuser) GetState() CalmState {
	return consensuser.state
}

func (consensuser *CalmConsensuser) OnDisconnect(idFrom cont.NodeId) {
	if consensuser.currentNodes.Consist(uint32(idFrom)) == false {
		log.Printf("Consensuser [%d]: Disconnect message out of group from %d", consensuser.id, idFrom)
		return
	}

	log.Printf("Consensuser [%d]: Disconnect %d node", consensuser.id, int(idFrom))
	disconnectedSet := cont.NewSetFromValue(uint32(idFrom))
	consensuser.OnVote(cont.NewMessageVote(*consensuser.carry,
		                                   consensuser.votedSet.Diff(disconnectedSet),
										   consensuser.currentNodes.Diff(disconnectedSet)))
}
