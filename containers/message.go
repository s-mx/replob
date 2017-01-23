package containers

// States for messages
const (
	Vote       = iota
	Commit     = iota
)

type NodeId int
type Stamp int

type Message struct {
	typeMessage int
	Stamp       Stamp
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
	IdFrom		NodeId
}

func NewMessageVote(stamp Stamp, carrySet *CarriesSet, votedSet *Set, nodesSet *Set, idFrom NodeId) *Message {
	return &Message{
        typeMessage:Vote,
        Stamp:stamp,
        CarriesSet:*carrySet,
        VotedSet:*votedSet,
        NodesSet:*nodesSet,
        IdFrom:idFrom,
    }
}

func NewMessageCommit(stamp Stamp, CarrySet *CarriesSet) *Message {
	return &Message{
        typeMessage:Commit,
        Stamp:stamp,
        CarriesSet:*CarrySet,
    }
}

func (msg *Message) GetType() int {
	return msg.typeMessage
}

// For testing purposes
func (msg *Message) notEqual(otherMsg *Message) bool {
	return msg.VotedSet.NotEqual(&otherMsg.VotedSet)
}
