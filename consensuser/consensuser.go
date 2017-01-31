package consensuser

import (
	cont "github.com/s-mx/replob/containers"
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
	Broadcaster
	Committer

	State      CalmState
	Id         cont.NodeId
	myStamp    cont.Stamp // TODO: move to Broadcaster implementation
	Stamps     []cont.Stamp // TODO: move to Receiver implementation
	Nodes      cont.Set
	VotedSet   cont.Set
	CarriesSet cont.CarriesSet
	curStepId  cont.StepId
}

func NewCalmConsensuser(broadcaster Broadcaster, committer Committer,
	conf *MasterlessConfiguration, id cont.NodeId) *CalmConsensuser {
	return &CalmConsensuser{
		Broadcaster: broadcaster,
		Committer:   committer,
		State:       Initial,
		Id:          id,
		myStamp:     cont.Stamp(0),
		Stamps:      make([]cont.Stamp, conf.GetNumberNodes()),
		Nodes:       conf.Info,
		curStepId:   cont.StepId(0),
	}
}

func (consensuser *CalmConsensuser) messageIsOutdated(msg cont.Message) bool {
	return consensuser.Stamps[uint32(msg.IdFrom)] >= msg.Stamp
}

func (consensuser *CalmConsensuser) updateMessageStamp(msg cont.Message) {
	if consensuser.messageIsOutdated(msg) == false {
		consensuser.Stamps[int(msg.IdFrom)] = msg.Stamp
	}
}

func (consensuser *CalmConsensuser) NextStamp() cont.Stamp {
	consensuser.myStamp += 1
	return consensuser.myStamp
}

func (consensuser *CalmConsensuser) Broadcast() {
	// FIXME: generate next stamp inside broadcaster
	msg := cont.NewMessageVote(consensuser.NextStamp(), consensuser.curStepId,
	 					       consensuser.CarriesSet, consensuser.VotedSet, consensuser.Nodes,
							   consensuser.Id)
	consensuser.Broadcaster.Broadcast(*msg)
}

func (consensuser *CalmConsensuser) Propose(carrier cont.Carry) {
	if consensuser.State != Initial {
		return // FIXME: Log.Fatal
	}

	carrySet := cont.NewCarriesSet(carrier)
	msg := consensuser.newVote(*carrySet, consensuser.Nodes)
	consensuser.OnVote(*msg) // FIXME: refactor to use single line
}

// checks that all nodes are agreed on sequence of carries and nodes group
func (consensuser *CalmConsensuser) checkInvariant(msg cont.Message) bool {
	return consensuser.CarriesSet.Equal(msg.CarriesSet) &&
		   consensuser.Nodes.Equal(msg.NodesSet)
}

func (consensuser *CalmConsensuser) mergeVotes(msg cont.Message) {
	consensuser.CarriesSet.AddSet(msg.CarriesSet)
	consensuser.Nodes.Intersect(msg.NodesSet)
}

// FIXME: generate next stamp inside broadcaster
// FIXME: use return by value
func (consensuser *CalmConsensuser) newVote(carrySet cont.CarriesSet, nodesSet cont.Set) *cont.Message {
	stamp := consensuser.NextStamp()
	return cont.NewMessageVote(stamp, consensuser.curStepId, carrySet, consensuser.VotedSet, nodesSet, consensuser.Id)
}

func (consensuser *CalmConsensuser) OnBroadcast(msg cont.Message) {
	if consensuser.curStepId < msg.StepId {
		// TODO: use ConsensusController.LostSteps(...)
		consensuser.PrepareFutureStep(msg.StepId) // TODO: need to reinitialize consensuser state (group membership changes)
		// FIXME: change the state to stop receiving any messages: just log & return
	}

	if consensuser.messageIsOutdated(msg) ||
	   consensuser.Nodes.Consist(uint32(msg.IdFrom)) == false ||
	   consensuser.curStepId > msg.StepId {
		return
	}

	consensuser.updateMessageStamp(msg)
	if msg.GetType() == cont.Vote {
		consensuser.OnVote(msg)
	} else if msg.GetType() == cont.Commit {
		consensuser.OnVote(msg) // FIXME: use OnCommit
	}
}

func (consensuser *CalmConsensuser) OnVote(msg cont.Message) {
	if consensuser.State == MayCommit && consensuser.checkInvariant(msg) == false {
		consensuser.State = CannotCommit
	}

	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
	consensuser.mergeVotes(msg) // don't use msg right after this line
	consensuser.VotedSet.Intersect(consensuser.Nodes)
	// TODO: check for majority by comparing with initial number of nodes in Initial state, log it and stop processing

	if consensuser.State == Initial {
		consensuser.State = MayCommit
		consensuser.Broadcast()
	}

	if consensuser.VotedSet.Equal(consensuser.Nodes) {
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
	consensuser.Broadcaster.Broadcast(consensuser.newCommitMessage())
	consensuser.PrepareNextStep()
}

func (consensuser *CalmConsensuser) newCommitMessage() cont.Message {
	// FIXME: generate stamp inside broadcaster
	stamp := consensuser.NextStamp()
	return *cont.NewMessageCommit(stamp, consensuser.curStepId, consensuser.CarriesSet)
}

func (consensuser *CalmConsensuser) CleanUp() {
	consensuser.CarriesSet.Clear()
	consensuser.State = Initial
	consensuser.VotedSet.Clear()
}

func (consensuser *CalmConsensuser) PrepareNextStep() {
	consensuser.CleanUp()
	consensuser.curStepId.Inc()
}

func (consensuser *CalmConsensuser) PrepareFutureStep(step cont.StepId) {
	if consensuser.curStepId < step {
		consensuser.CleanUp()
		consensuser.curStepId = step
	}
}

func (consensuser *CalmConsensuser) OnDisconnect(idFrom cont.NodeId) {
	/*
	FIXME:
	algorithm:
	1. nodes must be updated in OnVote only
	2. votes must not be changed (may be updated only in OnVote)
	 */
	consensuser.Nodes.Erase(idFrom)
	if consensuser.State != Initial {
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		votedSet := consensuser.VotedSet.Diff(otherSet)
		msg := cont.NewMessageVote(consensuser.myStamp, consensuser.curStepId, // FIXME: stamp and stepId may be equal to zero here
								   consensuser.CarriesSet, votedSet, consensuser.Nodes, consensuser.Id)
		consensuser.OnVote(*msg)
	}
}
