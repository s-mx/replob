package main

import (
	//"log"
	//"sync"
    //"net/rpc"
)

type NodeId int

type SetNodes struct {
    markerNodes map[uint32]bool
}

func (set *SetNodes) Change(ind uint32, val bool) {
    set.markerNodes[ind] = val
}

func NewSetNodes(numberNodes uint32) *SetNodes {
    ptr_set := new(SetNodes)
    ptr_set.markerNodes = make(map[uint32]bool)
    for i := uint32(0); i < numberNodes; i++ {
        ptr_set.markerNodes[i] = false
    }
    return ptr_set
}

type NodesInfo struct {
    numberNodes uint32
    plugInNodes []bool
}

func NewNodesInfo(numberNodes uint32) *NodesInfo {
    info := new(NodesInfo)
    info.numberNodes = numberNodes
    info.plugInNodes = make([]bool, numberNodes)
    for i := range info.plugInNodes {
        info.plugInNodes[i] = true
    }
    return info
}

type MasterlessConfiguration struct {
    Info NodesInfo
}

func NewMasterlessConfiguration(numberNodes uint32) *MasterlessConfiguration {
    conf := new(MasterlessConfiguration)
    conf.Info = *NewNodesInfo(numberNodes)
    return conf
}

type Carrier struct {
    Id uint32
    Value int
}

// States for messages
const (
    Vote = iota
    Commit = iota
    Disconnect = iota
)

type Message struct {
    typeMessage int
    VotedSet SetNodes
    CarrySet SetNodes
    NodesSet SetNodes
}

type Broadcaster interface {
    Broadcast(Message, NodeId)
}

type QueueMessages struct {

}

type MyMainBroadcaster struct {
    queue QueueMessages
}

func (broadcaster MyMainBroadcaster) addMessage(msg Message, id NodeId) {
    broadcaster.queue.push(msg, id)
}

type MyBroadcaster struct {
    info *NodesInfo
    mainBroadcaster *MyMainBroadcaster
}

func (broadcaster MyBroadcaster) addMessage(msg Message, idFrom NodeId) {
    broadcaster.mainBroadcaster.addMessage(msg, idFrom)
}

func (broadcaster MyBroadcaster) Broadcast(msg Message, idFrom NodeId) {
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
    OnBroadcast(Message, NodeId)
    OnDisconnect(NodeId)
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
    State int
    Id NodeId
    NodesInfo NodesInfo
    VotedSet SetNodes
    CarriesSet SetNodes
    BroadcasterCommiter BroadcasterCommiter
}

func (consensuser MyConsensuser) Propose(carrier Carrier) {
    votedSet := *NewSetNodes(consensuser.NodesInfo.numberNodes)
    votedSet.Change(uint32(consensuser.Id), true)
    consensuser.BroadcasterCommiter.Broadcast(Message{typeMessage:Vote, VotedSet:votedSet}, consensuser.Id)
}

func (consensuser MyConsensuser) OnBroadcast(msg Message, idFrom NodeId) {
    if consensuser.State == Completed {
        return
    }

    if !consensuser.NodesInfo.consists(idFrom) {
        return
    }

    if consensuser.State == MayCommit && consensuser.carriesSet.NotEqual(msg.carrySet) {
        consensuser.State = CannotCommit
    }

    consensuser.CarriesSet.addSet(msg.carrySet)

    consensuser.VotedSet.addSet(NewSetNodes().insert(idFrom))
    consensuser.VotedSet.addSet(NewSetNodes().insert(consensuser.Id))

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
        otherSet := SetNodes{}
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
