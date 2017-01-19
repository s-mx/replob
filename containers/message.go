package containers

// States for messages
const (
	Vote       = iota
	Commit     = iota
)

type Stamp int

type Message struct {
	typeMessage int
	Stamp       Stamp
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
}

func NewMessageVote(stamp Stamp, carrySet *CarriesSet, votedSet *Set, nodesSet *Set) *Message {
	ptrMessage := new(Message)
	ptrMessage.typeMessage = Vote
	ptrMessage.Stamp = stamp
	ptrMessage.CarriesSet = *carrySet
	ptrMessage.VotedSet = *votedSet
	ptrMessage.NodesSet = *nodesSet
	return ptrMessage
}

func NewMessageCommit(stamp Stamp, CarrySet *CarriesSet) *Message {
	ptrMessage := new(Message)
	ptrMessage.Stamp = stamp
	ptrMessage.typeMessage = Commit
	ptrMessage.CarriesSet = *CarrySet
	return ptrMessage
}

func (msg *Message) GetType() int {
	return msg.typeMessage
}

// For testing purposes
func (msg *Message) notEqual(otherMsg *Message) bool {
	return msg.VotedSet.NotEqual(&otherMsg.VotedSet)
}
