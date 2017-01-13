package containers

// States for messages
const (
	Vote       = iota
	Commit     = iota
	Disconnect = iota
)

type Message struct {
	typeMessage int
	Stamp       int
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
}

func NewMessageVote(typeMessage int, stamp int, carrySet *CarriesSet, votedSet *Set, nodesSet *Set) *Message {
	ptrMessage := new(Message)
	ptrMessage.typeMessage = typeMessage
	ptrMessage.Stamp = stamp
	ptrMessage.CarriesSet = *carrySet
	ptrMessage.VotedSet = *votedSet
	ptrMessage.NodesSet = *nodesSet
	return ptrMessage
}

func NewMessageCommit(typeMessage int, stamp int, CarrySet *CarriesSet) *Message {
	ptrMessage := new(Message)
	ptrMessage.Stamp = stamp
	ptrMessage.typeMessage = typeMessage
	ptrMessage.CarriesSet = *CarrySet
	return ptrMessage
}

func (msg *Message) GetType() int {
	return msg.typeMessage
}
