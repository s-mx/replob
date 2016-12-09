package containers

// States for messages
const (
    Vote       = iota
    Commit     = iota
    Disconnect = iota
)

type Message struct {
    typeMessage int
    VotedSet    Set
    CarrySet    Set
    NodesSet    Set
}

func NewMessageVote(typeMessage int, votedSet Set) *Message {
    ptrMessage := new(Message)
    ptrMessage.typeMessage = typeMessage
    ptrMessage.VotedSet = votedSet
    return ptrMessage
}

func NewMessageCommit(typeMessage int, CarrySet Set) *Message {
    ptrMessage := new(Message)
    ptrMessage.typeMessage = typeMessage
    ptrMessage.VotedSet = CarrySet
    return ptrMessage
}

func (msg *Message) GetType() int {
    return msg.typeMessage
}
