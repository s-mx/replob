package containers

// States for messages
const (
	Vote       = iota
	Commit
	Empty
)

type NodeId int
type Stamp int
type StepId int

func (step *StepId) Inc() {
	*step += 1
}

func (step *StepId) Equal(rghStep *StepId) bool {
	return *step == *rghStep
}

func (step *StepId) NotEqual(rghStep *StepId) bool {
	return ! step.Equal(rghStep)
}

// TODO: Реализовать интерфейс Marshall, UnMarshall
type Message struct {
	TypeMessage int
	Stamp       Stamp
	StepId      StepId
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
	IdFrom      NodeId
}

func NewMessageVote(carrySet CarriesSet, votedSet Set, nodesSet Set) Message {
	return Message{
        TypeMessage: Vote,
        CarriesSet:  carrySet,
        VotedSet:    votedSet,
        NodesSet:    nodesSet,
    }
}

func NewMessageCommit(CarrySet CarriesSet) *Message {
	return &Message{
        TypeMessage: Commit,
        CarriesSet:  CarrySet,
    }
}

func NewEmptyMessage() Message {
	return Message{TypeMessage: Empty}
}

func (msg *Message) GetType() int {
	return msg.TypeMessage
}

// For testing purposes
func (msg *Message) notEqual(otherMsg *Message) bool {
	return msg.VotedSet.NotEqual(otherMsg.VotedSet)
}
