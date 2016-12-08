package main

import (
    "github.com/s-mx/replob/containers"
    "github.com/s-mx/replob/nodes"
    //"sync"
    //"net/rpc"
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

type Broadcaster interface {
    Broadcast(*containers.Message, nodes.NodeId)
}

type MyMainBroadcaster struct {
    queue containers.QueueMessages
}

func (broadcaster *MyMainBroadcaster) addMessage(msg *containers.Message, id nodes.NodeId) {
    broadcaster.queue.Push(msg, uint32(id))
}

type MyBroadcaster struct {
    info            *nodes.NodesInfo
    mainBroadcaster *MyMainBroadcaster
}

func (broadcaster *MyBroadcaster) addMessage(msg *containers.Message, idFrom nodes.NodeId) {
    broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}

func (broadcaster *MyBroadcaster) Broadcast(msg *containers.Message, idFrom nodes.NodeId) {
    broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}

type Commiter interface {
    Commit(Carrier)
}

type MyCommiter struct {

}

func (commiter MyCommiter) Commit(carrier Carrier) {

}

type BroadcasterCommiter interface {
    Broadcaster
    Commiter
}

type MyBroadcasterCommiter struct {
    MyBroadcaster
    MyCommiter
}

type Consensuser interface {
    Propose(Carrier)
    OnBroadcast(containers.Message, nodes.NodeId)
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
    State               int
    Id                  nodes.NodeId
    NodesInfo           nodes.NodesInfo
    VotedSet            containers.Set
    CarriesSet          containers.Set
    BroadcasterCommiter BroadcasterCommiter
}

func (consensuser MyConsensuser) Propose(carrier Carrier) {
    votedSet := *containers.NewSet(consensuser.NodesInfo.Size())
    votedSet.Change(uint32(consensuser.Id), true)
    consensuser.BroadcasterCommiter.Broadcast(containers.NewMessageVote(containers.Vote, votedSet), consensuser.Id)
}

func (consensuser MyConsensuser) OnBroadcast(msg containers.Message, idFrom nodes.NodeId) {
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

    consensuser.VotedSet.AddSet(containers.NewSetFromValue(uint32(idFrom)))
    consensuser.VotedSet.AddSet(containers.NewSetFromValue(uint32(consensuser.Id)))

    if consensuser.NodesInfo.NodesNotEqual(&msg.NodesSet) {
        if consensuser.State == MayCommit {
            consensuser.State = CannotCommit
        }

        consensuser.NodesInfo.IntersectNodes(&msg.NodesSet)
        consensuser.VotedSet.Intersect(&msg.NodesSet)
    }

    if consensuser.VotedSet.Equal(&consensuser.NodesInfo.GetSet()) { // trouble
        if consensuser.State == MayCommit {
            consensuser.BroadcasterCommiter.Broadcast(Message{typeMessage:Commit, CarrySet:consensuser.CarriesSet})
        } else {
            consensuser.State = ToVote
        }
    }

    if consensuser.State == ToVote {
        consensuser.State = MayCommit
        consensuser.BroadcasterCommiter.Broadcast(Message{typeMessage:Vote, CarrySet:consensuser.CarriesSet,
            NodesSet:consensuser.NodesInfo.GetSet()})
    }
}

func (consensuser MyConsensuser) OnDisconnect(idFrom NodeId) {
    consensuser.NodesInfo.erase(idFrom)
    if consensuser.State == Initial {
        consensuser.NodesInfo.erase(idFrom)
    } else {
        set := consensuser.NodesInfo.GetSet()
        otherSet := containers.SetNodes{}
        otherSet.insert(idFrom)
        consensuser.BroadcasterCommiter.Broadcast(Message{typeMessage:Vote, set.diff(otherSet)})
    }
}

func NewMasterlessConsensus(broadcasterCommitter *BroadcasterCommiter,
                            conf *MasterlessConfiguration, id NodeId) Consensuser {
    var consensuser MyConsensuser
    consensuser.Id = id
    consensuser.State = ToVote
    consensuser.NodesInfo = conf.Info
    consensuser.BroadcasterCommiter = broadcasterCommitter
    return consensuser
}

func NewBroadcasterCommiter(numberNodes uint32) BroadcasterCommiter {
    broadcasterCommiter := *new(MyBroadcasterCommiter)
    return broadcasterCommiter
}

func main() {
    var number uint32
    number = 5
    arr := make([]Consensuser, number)

    bc := NewBroadcasterCommiter(number)
    conf := NewMasterlessConfiguration(number)

    for i := 0; i < int(number); i++ {
        arr[i] = NewMasterlessConsensus(bc, *conf, NodeId(i))
    }

    arr[0].Propose(Carrier{1, 12345})
}
