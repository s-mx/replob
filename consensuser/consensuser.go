package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"github.com/s-mx/replob/nodes"
)

type MasterlessConfiguration struct {
	Info nodes.NodesInfo
}

func NewMasterlessConfiguration(numberNodes uint32) *MasterlessConfiguration {
	conf := new(MasterlessConfiguration)
	conf.Info = *nodes.NewNodesInfo(numberNodes)
	return conf
}

type Carrier struct {
	Id    uint32
	Value int
}

func NewCarrier(id uint32, val int) *Carrier {
	ptr := new(Carrier)
	ptr.Id = id
	ptr.Value = val
}

type Consensuser interface {
	Propose(Carrier)
	OnBroadcast(cont.Message, nodes.NodeId)
	OnDisconnect(nodes.NodeId)
}

// states for replicas
const (
	Initial      = iota
	ToVote       = iota
	MayCommit    = iota
	CannotCommit = iota
	Completed    = iota
)

type MyConsensuser struct {
	State       int
	Id          nodes.NodeId
	NodesInfo   nodes.NodesInfo
	VotedSet    cont.Set
	CarriesSet  cont.Set // Полная лажа. Надо исправить :(
	broadcaster Broadcaster
	committer   Committer
}

func NewConsensuser(broadcaster Broadcaster, committer Committer,
	conf *MasterlessConfiguration, id nodes.NodeId) *MyConsensuser {
	ptrConsensuser := new(MyConsensuser)
	ptrConsensuser.Id = id
	ptrConsensuser.State = ToVote
	ptrConsensuser.NodesInfo = conf.Info
	ptrConsensuser.broadcaster = broadcaster
	ptrConsensuser.committer = committer
	return ptrConsensuser
}

func (consensuser *MyConsensuser) Propose(carrier Carrier) {
	votedSet := cont.NewSet(consensuser.NodesInfo.Size())
	votedSet.Insert(uint32(consensuser.Id))
	nodesSet := consensuser.NodesInfo.GetSet()
	consensuser.broadcaster.Broadcast(cont.NewMessageVote(cont.Vote, votedSet, &nodesSet), consensuser.Id)
}

func (consensuser *MyConsensuser) OnBroadcast(msg cont.Message, idFrom nodes.NodeId) {
	if consensuser.State == Completed {
		return
	}

	if consensuser.NodesInfo.ConsistsId(idFrom) == false {
		return
	}

	if consensuser.State == MayCommit && consensuser.CarriesSet.NotEqual(&msg.CarrySet) {
		consensuser.State = CannotCommit
	}

	consensuser.CarriesSet.AddSet(&msg.CarrySet)

	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(idFrom)))
	consensuser.VotedSet.AddSet(cont.NewSetFromValue(uint32(consensuser.Id)))

	if consensuser.NodesInfo.NodesNotEqual(&msg.NodesSet) { // Не правильно. У каждой реплики свой nodes_
		if consensuser.State == MayCommit {
			consensuser.State = CannotCommit
		}

		consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
		consensuser.VotedSet.Intersect(&msg.NodesSet)
	}

	nodesSet := consensuser.NodesInfo.GetSet()
	if consensuser.VotedSet.Equal(&nodesSet) {
		if consensuser.State == MayCommit {
			consensuser.committer.CommitSet(consensuser.CarriesSet)
			//consensuser.broadcaster.Broadcast(cont.NewMessageCommit(cont.Commit, &consensuser.CarriesSet), consensuser.Id) // We need more too looong lines
		} else {
			consensuser.State = ToVote
		}
	}

	if consensuser.State == ToVote {
		consensuser.State = MayCommit
		consensuser.broadcaster.Broadcast(cont.NewMessageVote(cont.Vote, &consensuser.CarriesSet, &nodesSet), consensuser.Id)
	}
}

func (consensuser MyConsensuser) OnDisconnect(idFrom nodes.NodeId) {
	consensuser.NodesInfo.Erase(idFrom)
	if consensuser.State == Initial {
		consensuser.NodesInfo.Erase(idFrom)
	} else {
		set := consensuser.NodesInfo.GetSet()
		otherSet := cont.NewSetFromValue(uint32(idFrom))
		consensuser.broadcaster.Broadcast(cont.NewMessageVote(cont.Vote, set.Diff(otherSet), &set), idFrom)
	}
}
