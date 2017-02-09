package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	log  "log"
)

// FIXME: OnBroadcast->OnReceive
// FIXME: use separated interface for OnReceive to filter incoming messages (drop outdated)
type Consensuser interface {
	Propose(cont.Carry)
	OnBroadcast(cont.Message)
	OnDisconnect(cont.NodeId)
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
	Committer

	dispatcher   *Dispatcher
	State        CalmState
	Id           cont.NodeId // Только для логирования.
	Nodes		 cont.Set
	CurrentNodes cont.Set
	VotedSet     cont.Set
	CarriesSet   cont.CarriesSet
}

func NewCalmConsensuser(dispatcher *Dispatcher, committer Committer,
	conf Configuration, id cont.NodeId) *CalmConsensuser {
	return &CalmConsensuser{
		dispatcher:   dispatcher,
		Committer:    committer,
		State:        Initial,
		Id:           id,
		Nodes:		  conf.Info,
		CurrentNodes: conf.Info,
	}
}

func (consensuser *CalmConsensuser) Broadcast() {
	msg := cont.NewMessageVote(consensuser.CarriesSet, consensuser.VotedSet, consensuser.CurrentNodes)
	(*consensuser.dispatcher).Broadcast(*msg) // FIXME
}

func (consensuser *CalmConsensuser) newVote(carry cont.Carry) cont.Message {
	return *cont.NewMessageVote(*cont.NewCarriesSet(carry), consensuser.VotedSet, consensuser.CurrentNodes)
}

// checks that all nodes are agreed on sequence of carries and nodes group
func (consensuser *CalmConsensuser) Propose(carry cont.Carry) {
	if consensuser.State != Initial {
		log.Fatalf("state of consenuser %d isn't Initial on propose", consensuser.Id)
	}

	consensuser.OnVote(consensuser.newVote(carry))
}

func (consensuser *CalmConsensuser) checkInvariant(msg cont.Message) bool {
	return consensuser.CarriesSet.Equal(msg.CarriesSet) &&
		   consensuser.CurrentNodes.Equal(msg.NodesSet)
}

func (consensuser *CalmConsensuser) OnBroadcast(msg cont.Message) {
	// Думаю, это чекать нужно здесь.
	if consensuser.CurrentNodes.Consist(uint32(msg.IdFrom)) == false {
		return
	}

	if msg.GetType() == cont.Vote {
		consensuser.OnVote(msg)
	} else if msg.GetType() == cont.Commit {
		consensuser.OnCommit()
	}
}

func (consensuser *CalmConsensuser) mergeVotes(msg cont.Message) {
	consensuser.CarriesSet.AddSet(msg.CarriesSet)
	consensuser.CurrentNodes.Intersect(msg.NodesSet)
}

func (consensuser *CalmConsensuser) OnVote(msg cont.Message) {
	if consensuser.State == MayCommit && consensuser.checkInvariant(msg) == false {
		consensuser.State = CannotCommit
	}

	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
	consensuser.mergeVotes(msg) // don't use msg right after this line
	consensuser.VotedSet.Intersect(consensuser.CurrentNodes)
	if consensuser.Nodes.Size() > consensuser.CurrentNodes.Size() * 2 {
		log.Println("current set of nodes of %d consensuser become less than majority", consensuser.Id)
		(*consensuser.dispatcher).Stop()
	}
	// TODO: check for majority by comparing with initial number of nodes in Initial state, log it and stop processing

	if consensuser.State == Initial {
		consensuser.State = MayCommit
		consensuser.Broadcast()
	}

	if consensuser.VotedSet.Equal(consensuser.CurrentNodes) {
		if consensuser.State == MayCommit {
			consensuser.OnCommit()
		} else {
			consensuser.State = MayCommit
			consensuser.VotedSet.Clear()
			consensuser.Broadcast()
		}
	}
}

func (consensuser *CalmConsensuser) OnCommit() {
	consensuser.Committer.CommitSet(consensuser.CarriesSet)
	(*consensuser.dispatcher).Broadcast(consensuser.newCommitMessage()) // FIXME
	consensuser.PrepareNextStep()
}

func (consensuser *CalmConsensuser) newCommitMessage() cont.Message {
	return *cont.NewMessageCommit(consensuser.CarriesSet)
}

func (consensuser *CalmConsensuser) CleanUp() {
	consensuser.State = Initial
	consensuser.Nodes = consensuser.CurrentNodes
	consensuser.CarriesSet.Clear()
	consensuser.VotedSet.Clear()
}

func (consensuser *CalmConsensuser) PrepareNextStep() {
	consensuser.CleanUp()
	(*consensuser.dispatcher).IncStep() // FIXME
}

func (consensuser *CalmConsensuser) OnDisconnect(idFrom cont.NodeId) {
	if consensuser.State == Initial {
		return
	}

	disconnectedSet := cont.NewSetFromValue(uint32(idFrom))
	consensuser.OnVote(*cont.NewMessageVote(consensuser.CarriesSet,
											consensuser.VotedSet.Diff(disconnectedSet),
											consensuser.CurrentNodes.Diff(disconnectedSet)))
}
