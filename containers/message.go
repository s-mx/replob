package containers

// States for messages
const (
	Vote       = iota
	Commit     = iota
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

type Message struct {
	typeMessage int
	Stamp       Stamp
	StepId		StepId
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
	IdFrom		NodeId
}

func NewMessageVote(carrySet CarriesSet, votedSet Set, nodesSet Set) *Message {
	return &Message{
        typeMessage:Vote,
        CarriesSet:carrySet,
        VotedSet:votedSet,
        NodesSet:nodesSet,
    }
}

func NewMessageCommit(CarrySet CarriesSet) *Message {
	return &Message{
        typeMessage:Commit,
        CarriesSet:CarrySet,
    }
}

func (msg *Message) GetType() int {
	return msg.typeMessage
}

// For testing purposes
func (msg *Message) notEqual(otherMsg *Message) bool {
	return msg.VotedSet.NotEqual(otherMsg.VotedSet)
}
