package main

import (
    "github.com/s-mx/replob/containers"
    //"sync"
    //"net/rpc"
)

type MasterlessConfiguration struct {
    Info containers.NodesInfo
}

func NewMasterlessConfiguration(numberNodes uint32) *MasterlessConfiguration {
    conf := new(MasterlessConfiguration)
    conf.Info = *containers.NewNodesInfo(numberNodes)
    return conf
}

type Carrier struct {
    Id    uint32
    Value int
}

type Message struct {
    typeMessage int
    VotedSet    containers.SetNodes
    CarrySet    containers.SetNodes
    NodesSet    containers.SetNodes
}

type Broadcaster interface {
    Broadcast(Message, containers.NodeId)
}

type MyMainBroadcaster struct {
    queue containers.QueueMessages
}

func (broadcaster MyMainBroadcaster) addMessage(msg Message, id containers.NodeId) {
    broadcaster.queue.Push(msg, id)
}

type MyBroadcaster struct {
    info            *containers.NodesInfo
    mainBroadcaster *MyMainBroadcaster
}

func (broadcaster MyBroadcaster) addMessage(msg Message, idFrom containers.NodeId) {
    broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}

func (broadcaster MyBroadcaster) Broadcast(msg Message, idFrom containers.NodeId) {
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
    OnBroadcast(Message, containers.NodeId)
    OnDisconnect(containers.NodeId)
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
    Id                  containers.NodeId
    NodesInfo           containers.NodesInfo
    VotedSet            containers.SetNodes
    CarriesSet          containers.SetNodes
    BroadcasterCommiter BroadcasterCommiter
}

func (consensuser MyConsensuser) Propose(carrier Carrier) {
    votedSet := *containers.NewSetNodes(consensuser.NodesInfo.Size())
    votedSet.Change(uint32(consensuser.Id), true)
    consensuser.BroadcasterCommiter.Broadcast(Message{typeMessage:containers.Vote, VotedSet:votedSet}, consensuser.Id)
}

func (consensuser MyConsensuser) OnBroadcast(msg Message, idFrom containers.NodeId) {
    if consensuser.State == Completed {
        return
    }

    if consensuser.NodesInfo.ConsistsId(idFrom) == false {
        return
    }

    if consensuser.State == MayCommit && consensuser.CarriesSet.NotEqual(msg.CarrySet) {
        consensuser.State = CannotCommit
    }

    consensuser.CarriesSet.AddSet(msg.CarrySet)

    consensuser.VotedSet.AddSet(containers.NewSetFromValue(idFrom))
    consensuser.VotedSet.AddSet(containers.NewSetNodes(0).Insert(consensuser.Id))

    if consensuser.NodesInfo.NotEqual(msg.NodeSet) {
        if consensuser.State == MayCommit {
            consensuser.State = CannotCommit
        }

        consensuser.NodesInfo.multiply(msg.NodesSet)
        consensuser.VotedSet.multiply(msg.NodesSet)
    }

    if consensuser.VotedSet.Equal(consensuser.NodesInfo.GetSet()) {
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

func NewMasterlessConsensus(broadcasterCommitter BroadcasterCommiter,
conf MasterlessConfiguration, id NodeId) Consensuser {
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
