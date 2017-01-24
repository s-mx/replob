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

func NewMessageVote(stamp Stamp, stepId StepId, carrySet *CarriesSet, votedSet *Set, nodesSet *Set, idFrom NodeId) *Message {
	return &Message{
        typeMessage:Vote,
        Stamp:stamp,
		StepId:stepId,
        CarriesSet:*carrySet,
        VotedSet:*votedSet,
        NodesSet:*nodesSet,
        IdFrom:idFrom,
    }
}

func NewMessageCommit(stamp Stamp, stepId StepId, CarrySet *CarriesSet) *Message {
	return &Message{
        typeMessage:Commit,
        Stamp:stamp,
		StepId:stepId,
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
