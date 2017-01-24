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
type State int

const (
	Initial      State = iota
	ToVote             = iota
	MayCommit          = iota
	CannotCommit       = iota
	Completed          = iota
)

type CalmConsensuser struct {
	State       State
	Id          cont.NodeId
	myStamp     cont.Stamp
	Stamps      []cont.Stamp
	NodesInfo   nodes.NodesInfo
	VotedSet    cont.Set
	CarriesSet  cont.CarriesSet
	broadcaster *Broadcaster
	committer   *Committer
	curStepId   cont.StepId
}

func NewCalmConsensuser(broadcaster *Broadcaster, committer *Committer,
	conf *MasterlessConfiguration, id cont.NodeId) *CalmConsensuser {
	return &CalmConsensuser{
		State:       Initial,
		Id:          id,
		myStamp:     cont.Stamp(0),
		Stamps:      make([]cont.Stamp, conf.GetNumberNodes()),
		NodesInfo:   conf.Info,
		broadcaster: broadcaster,
		committer:   committer,
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

func (consensuser *CalmConsensuser) Broadcast(msg *cont.Message) {
	(*consensuser.broadcaster).Broadcast(*msg)
}

func (consensuser *CalmConsensuser) BroadcastCommit(carriesSet *cont.CarriesSet) {
	(*consensuser.committer).CommitSet(int(consensuser.Id), *carriesSet)
}

func (consensuser *CalmConsensuser) Propose(carrier cont.Carry) {
	votedSet := cont.NewSet(0)
	votedSet.Insert(uint32(consensuser.Id))
	nodesSet := consensuser.NodesInfo.GetSet()
	consensuser.State = Initial

	carrySet := cont.NewCarriesSet(carrier)
	msg := consensuser.newVote(carrySet, &nodesSet)
	consensuser.Broadcast(msg)
	consensuser.OnVote(*msg)
}

func (consensuser *CalmConsensuser) checkInvariant(msg *cont.Message) bool {
	return consensuser.CarriesSet.Equal(&msg.CarriesSet) &&
		consensuser.NodesInfo.NodesEqual(&msg.NodesSet)
}

func (consensuser *CalmConsensuser) mergeVotes(msg *cont.Message) {
	consensuser.CarriesSet.AddSet(&msg.CarriesSet)
	consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
}

func (consensuser *CalmConsensuser) newVote(carrySet *cont.CarriesSet, nodesSet *cont.Set) *cont.Message {
	stamp := consensuser.NextStamp()
	return cont.NewMessageVote(stamp, consensuser.curStepId, carrySet, &consensuser.VotedSet, nodesSet, consensuser.Id)
}

func (consensuser *CalmConsensuser) OnVote(msg cont.Message) {
	if consensuser.messageIsOutdated(msg) ||
		consensuser.NodesInfo.ConsistsId(msg.IdFrom) == false ||
		consensuser.curStepId.NotEqual(&msg.StepId) {
		return
	}

	if consensuser.State == MayCommit && consensuser.checkInvariant(&msg) == false {
		consensuser.State = CannotCommit
	}

	consensuser.mergeVotes(&msg)
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
	consensuser.VotedSet.Intersect(&msg.NodesSet)

	if consensuser.State == Initial {
		consensuser.State = MayCommit
		nodesSet := consensuser.NodesInfo.GetSet()
		voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
		consensuser.Broadcast(voteMsg)
	}

	if consensuser.VotedSet.Equal(&msg.NodesSet) {
		if consensuser.State == MayCommit {
			consensuser.State = Completed
			consensuser.OnCommit()
		} else {
			consensuser.State = MayCommit
			consensuser.VotedSet.Clear()
			nodesSet := consensuser.NodesInfo.GetSet()
			voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
			consensuser.Broadcast(voteMsg)
		}
	}
}

func (consensuser *CalmConsensuser) OnCommit() {
	consensuser.BroadcastCommit(&consensuser.CarriesSet)
	consensuser.CleanUp()
	consensuser.curStepId.Inc()
}

func (consensuser *CalmConsensuser) CleanUp() {
	consensuser.CarriesSet.Clear()
	consensuser.State = ToVote
	consensuser.VotedSet.Clear()
}

func (consensuser *CalmConsensuser) OnDisconnect(idFrom cont.NodeId) {
	consensuser.NodesInfo.Erase(idFrom)
	if consensuser.State == ToVote {
		consensuser.NodesInfo.Erase(idFrom)
	} else {
		set := consensuser.NodesInfo.GetSet()
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		stamp := consensuser.NextStamp()
		votedSet := consensuser.VotedSet.Diff(otherSet)
		msg := cont.NewMessageVote(stamp, consensuser.curStepId, &consensuser.CarriesSet, votedSet, set.Diff(otherSet), idFrom)
		consensuser.Broadcast(msg)
	}
}
