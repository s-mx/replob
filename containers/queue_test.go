package containers

import (
	"testing"
)

func TestPushPop(t *testing.T) {
	queue := NewQueueMessages()

	if queue.Size() != 0 {
		t.Error("Size of empty queue not zero")
	}

	messages := make([]Message, 0)
	votedSet := NewSet(10)
	carriesSet := NewCarriesSet()
	nodesSet := NewSet(10)

	for ind := 0; ind < 10; ind++ {
		messages = append(messages, *NewMessageVote(Stamp(ind), 0, *carriesSet, votedSet, nodesSet, NodeId(ind)))
		votedSet.Insert(uint32(ind))
		queue.Push(messages[ind])
		if queue.Size() != ind+1 {
			t.Error("Size of queue after push is wrong")
		}
	}

	for ind := 9; ind >= 0; ind-- {
		msg:= queue.Pop()
		IdFrom := uint32(msg.IdFrom)

		condition1 := queue.Size() != ind
		condition2 := msg.notEqual(&messages[ind])
		condition3 := IdFrom != uint32(9-ind)

		if condition1 || condition2 || condition3 {
			if condition1 {
				t.Errorf("Size of queue after pop is wrong: %d given, %d expected", queue.Size(), ind)
			}

			if condition2 {
				t.Error("Wrong message after pop")
			}

			if condition3 {
				t.Errorf("Wrong id after pop: %d given, %d expected", IdFrom, 9-ind)
			}
		}
	}
}
