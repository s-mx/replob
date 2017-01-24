package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Consensuser interface {
	Propose(cont.Carry)
	OnBroadcast(cont.Message)
	OnDisconnect(cont.NodeId)
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
	myStamp    cont.Stamp
	Stamps     []cont.Stamp
	NodesInfo  nodes.NodesInfo
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
		NodesInfo:   conf.Info,
		curStepId:   cont.StepId(0),
	}
}

func (consensuser *CalmConsensuser) messageIsOutdated(msg *cont.Message) bool {
	return consensuser.Stamps[uint32(msg.IdFrom)] >= msg.Stamp
}

func (consensuser *CalmConsensuser) updateMessageStamp(msg *cont.Message) {
	if consensuser.messageIsOutdated(msg) == false {
		consensuser.Stamps[int(msg.IdFrom)] = msg.Stamp
	}
}

func (consensuser *CalmConsensuser) NextStamp() cont.Stamp {
	consensuser.myStamp += 1
	return consensuser.myStamp
}

func (consensuser *CalmConsensuser) BroadcastCommit(carriesSet *cont.CarriesSet) {
	// FIXME: add broadcast
	// FIXME: move commit outside broadcast
	consensuser.CommitSet(int(consensuser.Id), *carriesSet)
}

func (consensuser *CalmConsensuser) Propose(carrier cont.Carry) {
	// TODO: propose may be invoked only in Initial state
	votedSet := cont.NewSet(0) // TODO: should be removed
	votedSet.Insert(uint32(consensuser.Id))
	nodesSet := consensuser.NodesInfo.GetSet()
	consensuser.State = Initial

	carrySet := cont.NewCarriesSet(carrier)
	msg := consensuser.newVote(carrySet, &nodesSet)
	consensuser.Broadcast(msg) // FIXME: remove broadcast
	consensuser.OnVote(msg)
}

// checks that all nodes are agreed on sequence of carries and nodes group
func (consensuser *CalmConsensuser) checkInvariant(msg *cont.Message) bool {
	return consensuser.CarriesSet.Equal(&msg.CarriesSet) &&
		consensuser.NodesInfo.NodesEqual(&msg.NodesSet)
}

func (consensuser *CalmConsensuser) mergeVotes(msg *cont.Message) {
	consensuser.CarriesSet.AddSet(&msg.CarriesSet)
	consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
}

// FIXME: pass by value all arguments
// FIXME: generate next stamp inside broadcaster
func (consensuser *CalmConsensuser) newVote(carrySet *cont.CarriesSet, nodesSet *cont.Set) *cont.Message {
	stamp := consensuser.NextStamp()
	return cont.NewMessageVote(stamp, consensuser.curStepId, carrySet, &consensuser.VotedSet, nodesSet, consensuser.Id)
}

/*
FIXME: use OnBroadcast method to dispatch OnVote & OnCommit
use common logic to check for outdated message
 */
// FIXME: add updateMessageStamp to increase stamp
func (consensuser *CalmConsensuser) OnVote(msg *cont.Message) {
	// TODO: check for stepId from future
	if consensuser.messageIsOutdated(msg) ||
		consensuser.NodesInfo.ConsistsId(msg.IdFrom) == false ||
		consensuser.curStepId.NotEqual(&msg.StepId) {
		return
	}

	if consensuser.State == MayCommit && consensuser.checkInvariant(msg) == false {
		consensuser.State = CannotCommit
	}

	consensuser.mergeVotes(msg) // don't use msg right after this line
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
	// FIXME: use nodes group from NodesInfo instead of message
	consensuser.VotedSet.Intersect(&msg.NodesSet)

	if consensuser.State == Initial {
		consensuser.State = MayCommit
		// TODO: implement consensuser.Broadcast() without additional args
		nodesSet := consensuser.NodesInfo.GetSet()
		voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
		consensuser.Broadcast(voteMsg)
	}

	// FIXME: compare with current nodes group instead of message groups
	if consensuser.VotedSet.Equal(&msg.NodesSet) {
		if consensuser.State == MayCommit {
			consensuser.OnCommit()
		} else {
			consensuser.State = MayCommit
			consensuser.VotedSet.Clear()
			// TODO: use consensuser.Broadcast() as above
			nodesSet := consensuser.NodesInfo.GetSet()
			voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
			consensuser.Broadcast(voteMsg)
		}
	}
}

func (consensuser *CalmConsensuser) OnCommit() {
	consensuser.BroadcastCommit(&consensuser.CarriesSet)
	// TODO: consensuser.PrepareNextStep()
	consensuser.CleanUp()
	consensuser.curStepId.Inc()

	/*
	 Refactor:
	 0. outdated check
	 1. commit
	 2. broadcast commit
	 3. prepare next step
	  */

}

// TODO: rename to PrepareNextStep
func (consensuser *CalmConsensuser) CleanUp() {
	consensuser.CarriesSet.Clear()
	consensuser.State = Initial
	// TODO: increment current step
	consensuser.VotedSet.Clear()
}

func (consensuser *CalmConsensuser) OnDisconnect(idFrom cont.NodeId) {
	consensuser.NodesInfo.Erase(idFrom)
	if consensuser.State == Initial {
		consensuser.NodesInfo.Erase(idFrom)
	} else {
		set := consensuser.NodesInfo.GetSet()
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		// FIXME: remove next stamp
		stamp := consensuser.NextStamp()
		votedSet := consensuser.VotedSet.Diff(otherSet)
		// FIXME: invoke onVote synchronously without broadcast
		msg := cont.NewMessageVote(stamp, consensuser.curStepId, &consensuser.CarriesSet, votedSet, set.Diff(otherSet), idFrom)
		consensuser.Broadcast(msg)
	}
}
