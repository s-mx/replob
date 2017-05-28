package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
	"log"
	"math/rand"
)

type Context struct {
	numberNodes int
	conf		Configuration
	dispatchers	[]*TestLocalDispatcher
	helper		*testCommitHelper
}

func CreateNodes(numberNodes int, carry cont.Carry, t *testing.T) *Context {
	context := &Context{
		numberNodes:numberNodes,
		conf:NewMasterlessConfiguration(numberNodes),
	}

	context.dispatchers = NewLocalDispatchers(numberNodes, context.conf, t)
	context.helper = newTestCommitHelper(numberNodes, carry.GetElementaryCarries(), context.dispatchers)

	for ind := 0; ind < numberNodes; ind++ {
		committer := NewTestLocalCommitter(cont.NodeId(ind), context.helper)
		context.dispatchers[ind].committer = committer
		currentDispatcher := context.dispatchers[ind]
		context.dispatchers[ind].cons = NewCalmConsensuser(currentDispatcher, context.conf, ind)
	}

	return context
}

func TestOneNode(t *testing.T) {
	carries := cont.NewIntCarriesN(1, 0)
	context := CreateNodes(1, carries, t)

	context.dispatchers[0].Propose(carries)
	if context.helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestTwoNodes(t *testing.T) {
	carries := cont.NewIntCarriesN(2, 0)
	context := CreateNodes(2, carries, t)

	carry, _ := carries.GetCarry(0)
	context.dispatchers[0].Propose(*carry)
	context.dispatchers[0].proceedFirstMessage(1)
	context.dispatchers[1].proceedFirstMessage(0)
	context.dispatchers[0].ClearQueues()
	context.dispatchers[1].ClearQueues()

	carry, _ = carries.GetCarry(1)
	context.dispatchers[1].Propose(*carry)
	context.dispatchers[1].proceedFirstMessage(0)
	context.dispatchers[0].proceedFirstMessage(1)

	if context.helper.CheckSafety() == false {
		t.Error("Safety is broken")
	}
}

func TestThreeNodes(t *testing.T) {
	carries := cont.NewIntCarriesN(2, 0)
	context := CreateNodes(3, carries, t)

	carry, _ := carries.GetCarry(0)
	context.dispatchers[0].Propose(*carry)
	context.dispatchers[0].proceedFirstMessage(1)
	context.dispatchers[0].proceedFirstMessage(2)
	context.dispatchers[1].proceedFirstMessage(0)
	context.dispatchers[1].proceedFirstMessage(2)
	context.dispatchers[2].proceedFirstMessage(1)
	context.dispatchers[2].proceedFirstMessage(0)
	context.dispatchers[2].proceedFirstMessage(0)

	context.dispatchers[0].ClearQueues()
	context.dispatchers[1].ClearQueues()
	context.dispatchers[2].ClearQueues()

	carry, _ = carries.GetCarry(1)
	context.dispatchers[1].Propose(*carry)
	context.dispatchers[1].proceedFirstMessage(0)
	context.dispatchers[1].proceedFirstMessage(2)
	context.dispatchers[0].proceedFirstMessage(1)
	context.dispatchers[0].proceedFirstMessage(2)
	context.dispatchers[2].proceedFirstMessage(1)
	context.dispatchers[2].proceedFirstMessage(0)

	if context.helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func RunRandomTest(numberNodes int, numberCarries int, seed int64, t *testing.T) {
	Source := rand.NewSource(seed)
	generator := rand.New(Source)

	log.Printf("===START TEST===")

	carries := cont.NewIntCarriesN(numberCarries, 0)
	context := CreateNodes(numberNodes, carries, t)

	carry, _ := carries.GetCarry(0)
	context.dispatchers[0].Propose(*carry)

	numberProposedCarries := 1
	for numberProposedCarries != numberCarries {
		for true {
			flag := false
			for ind := 0; ind < numberNodes; ind++ {
				if context.dispatchers[ind].proceedRandomMessage(generator, 0) == true {
					flag = true
				}
			}

			ind := context.helper.findIndLastCommit(numberProposedCarries)
			if ind != -1 && numberProposedCarries < numberCarries {
				carry, _ = carries.GetCarry(numberProposedCarries)
				context.dispatchers[ind].Propose(*carry)
				numberProposedCarries += 1
				continue
			}

			if flag == false {
				break
			}
		}
	}

	for true {
		flag := false
		for ind := 0; ind < numberNodes; ind++ {
			if context.dispatchers[ind].proceedRandomMessage(generator, 0) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if context.helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestRandomMessages2(t *testing.T) {
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomTest(2, 1, seed, t)
	}
}

func TestRandomMessages5(t *testing.T) {
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomTest(5, 10, seed, t)
	}
}

func TestRandomMessages5_100(t *testing.T) {
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomTest(5, 100, seed, t)
	}
}

func TestRandomMessages10_10(t *testing.T) {
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomTest(10, 10, seed, t)
	}
}

func TestRandomMessages10_100(t *testing.T) {
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomTest(10, 100, seed, t)
	}
}

func TestDisconnectThreeNodes(t *testing.T) {
	carries := cont.NewIntCarriesN(1, 0)
	context := CreateNodes(3, carries, t)


	carry, _ := carries.GetCarry(0)
	context.dispatchers[0].Propose(*carry)
	context.dispatchers[0].proceedFirstMessage(1)
	context.dispatchers[0].proceedFirstMessage(2)
	context.dispatchers[0].Stop()
	context.dispatchers[1].cons.OnDisconnect(0)
	context.dispatchers[2].cons.OnDisconnect(0)

	context.dispatchers[1].proceedFirstMessage(2)
	context.dispatchers[2].proceedFirstMessage(1)
	context.dispatchers[1].proceedFirstMessage(2)
	context.dispatchers[2].proceedFirstMessage(1)

	if context.helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

type Probabilities struct {
	probDisconnect	float32
	probSwap		float32
}

func DisconnectNode(subsetDisconnectedNodes *cont.Set, indDisconnect uint32, context *Context) {
	subsetDisconnectedNodes.Erase(indDisconnect)

	context.dispatchers[indDisconnect].Stop()
	for ind := 0; ind < context.numberNodes; ind++ {
		if context.dispatchers[ind].IsRunning() {
			context.dispatchers[ind].cons.OnDisconnect(cont.NodeId(indDisconnect))
		}
	}
}

func RunRandomDisconnectTest(numberNodes int, numberCarries int, numberDisconnects int, seed int64, prob Probabilities, t *testing.T) {
	if numberDisconnects * 2 > numberNodes {
		log.Fatalf("%d disconnected nodes can become the majority of %d nodes", numberDisconnects, numberNodes)
	}

	log.Printf("===START TEST===   seed: %d", seed)

	Source := rand.NewSource(seed)
	generator := rand.New(Source)

	carries := cont.NewIntCarriesN(numberCarries, 0)
	context := CreateNodes(numberNodes, carries, t)

	subsetDisconnectedNodes := cont.NewRandomSubset(context.conf.Info, numberDisconnects, generator)

	carry, _ := carries.GetCarry(0)
	context.dispatchers[0].Propose(*carry)

	indLastPropose := uint32(0)
	numberProposedCarries := 1
	for numberProposedCarries != numberCarries {
		for true {
			flag := false
			// Disconnect this
			if subsetDisconnectedNodes.Size() > 0 &&  generator.Float32() < prob.probDisconnect {
				indDisconnect := subsetDisconnectedNodes.Get(0)
				if indDisconnect == indLastPropose && subsetDisconnectedNodes.Size() > 1 {
					indDisconnect = subsetDisconnectedNodes.Get(1)
				}

				if indDisconnect != indLastPropose {
					DisconnectNode(&subsetDisconnectedNodes, indDisconnect, context)
				}
			}

			for ind := 0; ind < numberNodes; ind++ {
				if context.dispatchers[ind].proceedRandomMessage(generator, prob.probSwap) == true {
					flag = true
				}
			}

			if numberProposedCarries < numberCarries {
				ind := context.helper.findIndLastCommit(numberProposedCarries)
				if ind != -1 {
					carry, _ = carries.GetCarry(numberProposedCarries)
					context.dispatchers[ind].cons.Propose(*carry)
					indLastPropose = uint32(ind)
					numberProposedCarries += 1
					continue
				}
			}

			if flag == false {
				break
			}
		}
	}

	for true {
		flag := false
		for ind := 0; ind < numberNodes; ind++ {
			if context.dispatchers[ind].proceedRandomMessage(generator, prob.probSwap) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if context.helper.CheckSafety() == false {
		log.Printf("Carry isn't committed")
		t.Fatal()
	}
}

func TestRandomDisconnect5(t *testing.T) {
	prob := Probabilities{0.05,0.1}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(5, 10, 2, seed, prob, t)
	}
}

func TestRandomDisconnect5_100(t *testing.T) {
	prob := Probabilities{0.05,0.1}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(5, 100, 2, seed, prob, t)
	}
}

func TestRandomDisconnect10_10(t *testing.T) {
	prob := Probabilities{0.05,0.1}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(10, 10, 4, seed, prob, t)
	}
}

func TestRandomDisconnect10_100(t *testing.T) {
	prob := Probabilities{0.05,0.1}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(10, 100, 4, seed, prob, t)
	}
}

/*
Tests TODO:
1. Disconnect + liveness checks.
	+ change safety check: all prefixes with the same length must be the same
	+ no disconnects && no message drops => all lengths must be the same
	+ minor disconnnects without message drops => there are majority nodes with desired messages
	- on limit dropped message on each step: full safety check
	- on dropped message: if there is no any message in the queues
	 	=> resend the latest message again to each client from each node
	- tests with swaps and drops
2. Propose must be right after commit.
 */
