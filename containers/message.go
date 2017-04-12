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

type Message struct {
	MessageType int
	Stamp       Stamp
	StepId      StepId
	VotedSet    Set
	CarriesSet  CarriesSet
	NodesSet    Set
	IdFrom      NodeId
}

func NewMessageVote(carrySet CarriesSet, votedSet Set, nodesSet Set) Message {
	return Message{
        MessageType: Vote,
        CarriesSet:  carrySet,
        VotedSet:    votedSet,
        NodesSet:    nodesSet,
    }
}

func NewMessageCommit(CarrySet CarriesSet) *Message {
	return &Message{
        MessageType: Commit,
        CarriesSet:  CarrySet,
    }
}

func NewEmptyMessage() Message {
	return Message{MessageType: Empty}
}

func (msg *Message) GetType() int {
	return msg.MessageType
}

// For testing purposes
func (msg *Message) notEqual(otherMsg *Message) bool {
	return msg.VotedSet.NotEqual(otherMsg.VotedSet)
}

func (msg *Message) Equal(other Message) bool {
	// FIXME: додумать этот кусок
	if msg.MessageType != other.MessageType || msg.IdFrom != other.IdFrom {
		return false
	}

	if msg.MessageType == Empty {
		return true
	}

	if msg.MessageType == Vote {
		if msg.NodesSet.Equal(other.NodesSet) && msg.VotedSet.Equal(other.VotedSet) &&
		   msg.Stamp == other.Stamp && msg.StepId == other.StepId {
			return true
		}
	}

	if msg.MessageType == Commit {
		if msg.NodesSet.Equal(other.NodesSet) && msg.CarriesSet.Equal(other.CarriesSet) &&
			msg.Stamp == other.Stamp && msg.StepId == other.StepId {
			return true
		}
	}

	return false
}

func (msg *Message) NotEqual(other Message) bool {
	return !msg.Equal(other)
}
