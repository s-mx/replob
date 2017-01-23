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

// states for replicas
const (
	ToVote       = iota
	MayCommit    = iota
	CannotCommit = iota
	Completed    = iota
)

type MyConsensuser struct {
    State       int
	Id          cont.NodeId
	myStamp     cont.Stamp
	Stamps      []cont.Stamp
	NodesInfo   nodes.NodesInfo
	VotedSet    cont.Set
	CarriesSet  cont.CarriesSet
	broadcaster Broadcaster
	committer   Committer
    activatedFlag   bool
}

func NewConsensuser(broadcaster Broadcaster, committer Committer,
	conf *MasterlessConfiguration, id cont.NodeId) *MyConsensuser {
	return &MyConsensuser{
        State:ToVote,
        Id:id,
        myStamp:cont.Stamp(0),
        Stamps:make([]cont.Stamp, conf.GetNumberNodes()),
        NodesInfo:conf.Info,
        broadcaster:broadcaster,
        committer:committer,
        activatedFlag:false,
    }
}

func (consensuser *MyConsensuser) messageIsOutdated(msg cont.Message) bool {
	return consensuser.Stamps[uint32(msg.IdFrom)] >= msg.Stamp
}

func (consensuser *MyConsensuser) updateMessageStamp(msg cont.Message) {
	if consensuser.messageIsOutdated(msg) == false {
		consensuser.Stamps[int(msg.IdFrom)] = msg.Stamp
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
    consensuser.Activate()

	carrySet := cont.NewCarriesSet(carrier)
	stamp := consensuser.NextStamp()
	msg := cont.NewMessageVote(stamp, carrySet, votedSet, &nodesSet, consensuser.Id)
	consensuser.broadcaster.Broadcast(*msg)
    consensuser.OnBroadcast(*msg)
}

func (consensuser *MyConsensuser) checkInvariant(msg *cont.Message) bool {
    return consensuser.CarriesSet.Equal(&msg.CarriesSet) &&
           consensuser.NodesInfo.NodesEqual(&msg.NodesSet)
}

func (consensuser *MyConsensuser) mergeVotes(msg *cont.Message) {
    consensuser.CarriesSet.AddSet(&msg.CarriesSet)
    consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
}

func (consensuser *MyConsensuser) newVote(carrySet *cont.CarriesSet, nodesSet *cont.Set) *cont.Message {
    stamp := consensuser.NextStamp()
    return cont.NewMessageVote(stamp, carrySet, &consensuser.VotedSet, nodesSet, consensuser.Id)
}

func (consensuser *MyConsensuser) Activated() bool {
    return consensuser.activatedFlag
}

func (consensuser *MyConsensuser) NotActivated() bool {
    return ! consensuser.Activated()
}

func (consensuser *MyConsensuser) Activate() {
    consensuser.activatedFlag = true
}

func (consensuser *MyConsensuser) OnVote(msg cont.Message) {
    if consensuser.messageIsOutdated(msg) ||
       consensuser.NodesInfo.ConsistsId(msg.IdFrom) == false {
        return
    }

    if consensuser.State == MayCommit && consensuser.checkInvariant(&msg) == false {
        consensuser.State = CannotCommit
    }

    consensuser.mergeVotes(&msg)
    consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))
    consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
    consensuser.VotedSet.Intersect(&msg.NodesSet)

    if consensuser.NotActivated() {
        consensuser.Activate()
		nodesSet := consensuser.NodesInfo.GetSet()
        voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
        consensuser.broadcaster.Broadcast(*voteMsg)
    }

    if consensuser.VotedSet.Equal(&msg.NodesSet) {
        if consensuser.State == MayCommit {
            consensuser.State = Completed
            consensuser.committer.BroadcastSet(consensuser.CarriesSet)
            consensuser.OnCommit()
            consensuser.CleanUp()
        } else {
            consensuser.State = MayCommit
            consensuser.VotedSet.Clear()
			nodesSet := consensuser.NodesInfo.GetSet()
			voteMsg := consensuser.newVote(&consensuser.CarriesSet, &nodesSet)
            consensuser.broadcaster.Broadcast(*voteMsg)
        }
    }
}

func (consensuser *MyConsensuser) OnCommit() {
	consensuser.committer.CommitSet(consensuser.CarriesSet)
}

func (consensuser *MyConsensuser) CleanUp() {
	consensuser.CarriesSet.Clear()
	consensuser.State = ToVote
	consensuser.VotedSet.Clear()
}

func (consensuser *MyConsensuser) OnBroadcast(msg cont.Message) {
    if consensuser.Id != msg.IdFrom && consensuser.messageIsOutdated(msg) {
        return
    }

    if consensuser.State == Completed {
		return
	}

	if consensuser.NodesInfo.ConsistsId(msg.IdFrom) == false {
		return
	}

	consensuser.updateMessageStamp(msg)

	if consensuser.State == MayCommit && consensuser.CarriesSet.NotEqual(&msg.CarriesSet) {
		consensuser.State = CannotCommit
	}

	consensuser.CarriesSet.AddSet(&msg.CarriesSet)

	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(msg.IdFrom)))
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
		msg := cont.NewMessageVote(stamp, &consensuser.CarriesSet, &consensuser.VotedSet, &nodesSet, consensuser.Id)
		consensuser.broadcaster.Broadcast(*msg)
	}
}

func (consensuser *MyConsensuser) OnDisconnect(idFrom cont.NodeId) {
	consensuser.NodesInfo.Erase(idFrom)
	if consensuser.State == ToVote {
		consensuser.NodesInfo.Erase(idFrom)
	} else {
		set := consensuser.NodesInfo.GetSet()
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		stamp := consensuser.NextStamp()
		votedSet := consensuser.VotedSet.Diff(otherSet)
		msg := cont.NewMessageVote(stamp, &consensuser.CarriesSet, votedSet, set.Diff(otherSet), idFrom)
		consensuser.broadcaster.Broadcast(*msg)
    }
}
