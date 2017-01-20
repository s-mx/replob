package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type Consensuser interface {
	Propose(cont.Carry)
	OnBroadcast(cont.Message, nodes.NodeId)
	OnDisconnect(nodes.NodeId)
}

// states for replicas
const (
	ToVote       = iota
	MayCommit    = iota
	CannotCommit = iota
	Completed    = iota
)

type MyConsensuser struct {
    State       int
	Id          nodes.NodeId
	myStamp     cont.Stamp
	Stamps      []cont.Stamp
	NodesInfo   nodes.NodesInfo
	VotedSet    cont.Set
	CarriesSet  cont.CarriesSet
	broadcaster Broadcaster
	committer   Committer
}

func NewConsensuser(broadcaster Broadcaster, committer Committer,
	conf *MasterlessConfiguration, id nodes.NodeId) *MyConsensuser {
	ptrConsensuser := new(MyConsensuser)
	ptrConsensuser.Id = id
	ptrConsensuser.Stamps = make([]cont.Stamp, conf.GetNumberNodes())
	ptrConsensuser.State = ToVote
	ptrConsensuser.NodesInfo = conf.Info
	ptrConsensuser.broadcaster = broadcaster
	ptrConsensuser.committer = committer
	return ptrConsensuser
}

func (consensuser *MyConsensuser) messageIsOutdated(msg cont.Message, idFrom nodes.NodeId) bool {
	return consensuser.Stamps[uint32(idFrom)] >= msg.Stamp
}

func (consensuser *MyConsensuser) updateMessageStamp(msg cont.Message, idFrom nodes.NodeId) {
	if consensuser.messageIsOutdated(msg, idFrom) == false {
		consensuser.Stamps[int(idFrom)] = msg.Stamp
	}
}

func (consensuser *MyConsensuser) NextStamp() cont.Stamp {
	consensuser.myStamp += 1
	return consensuser.myStamp
}

func (consensuser *MyConsensuser) Propose(carrier cont.Carry) {
	votedSet := cont.NewSet(0)
	votedSet.Insert(uint32(consensuser.Id))
	nodesSet := consensuser.NodesInfo.GetSet()
    consensuser.State = ToVote

	carrySet := cont.NewCarriesSet(carrier)
	stamp := consensuser.NextStamp()
	msg := cont.NewMessageVote(stamp, carrySet, votedSet, &nodesSet)
	consensuser.broadcaster.Broadcast(*msg, consensuser.Id)
    consensuser.OnBroadcast(*msg, consensuser.Id)
}

func (consensuser *MyConsensuser) OnBroadcast(msg cont.Message, idFrom nodes.NodeId) {
    if consensuser.Id != idFrom && consensuser.messageIsOutdated(msg, idFrom) {
        return
    }

    if consensuser.State == Completed {
		return
	}

	if consensuser.NodesInfo.ConsistsId(idFrom) == false {
		return
	}

	consensuser.updateMessageStamp(msg, idFrom)

	if consensuser.State == MayCommit && consensuser.CarriesSet.NotEqual(&msg.CarriesSet) {
		consensuser.State = CannotCommit
	}

	consensuser.CarriesSet.AddSet(&msg.CarriesSet)

	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(idFrom)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))

	if consensuser.NodesInfo.NodesNotEqual(&msg.NodesSet) {
		if consensuser.State == MayCommit {
			consensuser.State = CannotCommit
		}

		consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
		consensuser.VotedSet.Clear()
		//consensuser.VotedSet.Intersect(&msg.NodesSet)
	}

	nodesSet := consensuser.NodesInfo.GetSet()
	if consensuser.VotedSet.Equal(&nodesSet) {
		if consensuser.State == MayCommit {
			consensuser.committer.CommitSet(consensuser.CarriesSet)
			//consensuser.broadcaster.Broadcast(cont.NewMessageCommit(cont.Commit, &consensuser.CarriesSet), consensuser.Id) // We need more too looong lines
		} else {
			consensuser.State = ToVote
			consensuser.VotedSet.Clear()
		}
	}

	if consensuser.State == ToVote {
		consensuser.State = MayCommit
		stamp := consensuser.NextStamp()
		msg := cont.NewMessageVote(stamp, &consensuser.CarriesSet, &consensuser.VotedSet, &nodesSet)
		consensuser.broadcaster.Broadcast(*msg, consensuser.Id)
	}
}

func (consensuser *MyConsensuser) OnDisconnect(idFrom nodes.NodeId) {
	consensuser.NodesInfo.Erase(idFrom)
	if consensuser.State == ToVote {
		consensuser.NodesInfo.Erase(idFrom)
	} else {
		set := consensuser.NodesInfo.GetSet()
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		stamp := consensuser.NextStamp()
		votedSet := consensuser.VotedSet.Diff(otherSet)
		msg := cont.NewMessageVote(stamp, &consensuser.CarriesSet, votedSet, set.Diff(otherSet))
		consensuser.broadcaster.Broadcast(*msg, idFrom)
		//consensuser.broadcaster.Broadcast(cont.NewMessageVote(cont.Vote, set.Diff(otherSet), &set), idFrom)
	}
}
