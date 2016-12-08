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

func NewMessageVote(typeMesssage int, votedSet Set) *Message {
    ptrMessage := new(Message)
    ptrMessage.typeMessage = typeMesssage
    ptrMessage.VotedSet = votedSet
    return ptrMessage
}

func (msg *Message) GetType() int {
    return msg.typeMessage
}
