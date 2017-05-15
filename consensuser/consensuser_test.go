package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
	"math/rand"
	"log"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carries := cont.NewCarriesN(1)
	LocalDispatchers := NewLocalDispatchers(1, conf, t)

	helper := newTestCommitHelper(1, carries.GetElementaryCarries(), LocalDispatchers)
	cm := NewTestLocalCommitter(0, helper)
	LocalDispatchers[0].committer = cm
	dsp := LocalDispatchers[0]
	cons := NewCalmConsensuser(dsp, conf, 0)
	dsp.cons = cons

	dsp.Propose(carries)
	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func TestTwoNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(2)
	carries := cont.NewCarriesN(2)
	LocalDispatchers := NewLocalDispatchers(2, conf, t)

	helper := newTestCommitHelper(2, carries.GetElementaryCarries(), LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	LocalDispatchers[0].committer = cm1
	LocalDispatchers[1].committer = cm2
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]
	cons1 := NewCalmConsensuser(dsp1, conf, 0)
	LocalDispatchers[0].cons = cons1
	cons2 := NewCalmConsensuser(dsp2, conf, 1)
	LocalDispatchers[1].cons = cons2

	carry, _ := carries.GetCarry(0)
	cons1.Propose(*carry)
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[0].ClearQueues()
	LocalDispatchers[1].ClearQueues()

	carry, _ = carries.GetCarry(1)
	cons2.Propose(*carry)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[0].proceedFirstMessage(1)

	if helper.CheckSafety() == false {
		t.Error("Safety is broken")
	}
}


func TestThreeNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(3)
	carries := cont.NewCarriesN(2)
	LocalDispatchers := NewLocalDispatchers(3, conf, t)

	helper := newTestCommitHelper(3, carries.GetElementaryCarries(), LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cm3 := NewTestLocalCommitter(2, helper)

	LocalDispatchers[0].committer = cm1
	LocalDispatchers[1].committer = cm2
	LocalDispatchers[2].committer = cm3
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]
	dsp3 := LocalDispatchers[2]
	cons1 := NewCalmConsensuser(dsp1, conf, 0)
	cons2 := NewCalmConsensuser(dsp2, conf, 1)
	cons3 := NewCalmConsensuser(dsp3, conf, 2)
	LocalDispatchers[0].cons = cons1
	LocalDispatchers[1].cons = cons2
	LocalDispatchers[2].cons = cons3

	carry, _ := carries.GetCarry(0)
	cons1.Propose(*carry)
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[2].proceedFirstMessage(0)
	LocalDispatchers[2].proceedFirstMessage(0)

	LocalDispatchers[0].ClearQueues()
	LocalDispatchers[1].ClearQueues()
	LocalDispatchers[2].ClearQueues()

	carry, _ = carries.GetCarry(1)
	cons2.Propose(*carry)
	LocalDispatchers[1].proceedFirstMessage(0)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[2].proceedFirstMessage(0)

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

func RunRandomTest(numberNodes int, numberCarries int, seed int64, t *testing.T) {
	Source := rand.NewSource(seed)
	generator := rand.New(Source)

	log.Printf("===START TEST===")

	conf := NewMasterlessConfiguration(uint32(numberNodes))
	carries := cont.NewCarriesN(numberCarries)
	LocalDispatchers := NewLocalDispatchers(numberNodes, conf, t)

	helper := newTestCommitHelper(numberNodes, carries.GetElementaryCarries(), LocalDispatchers)
	consensusers := []*CalmConsensuser{}
	for ind := 0; ind < numberNodes; ind++ {
		cm := NewTestLocalCommitter(cont.NodeId(ind), helper)
		cons := NewCalmConsensuser(LocalDispatchers[ind], conf, ind)
		LocalDispatchers[ind].cons = cons
		LocalDispatchers[ind].committer = cm
		consensusers = append(consensusers, cons)
	}

	carry, _ := carries.GetCarry(0)
	consensusers[0].Propose(*carry)

	numberProposedCarries := 1
	for numberProposedCarries != numberCarries {
		for true {
			flag := false
			for ind := 0; ind < numberNodes; ind++ {
				if LocalDispatchers[ind].proceedRandomMessage(generator, 0) == true {
					flag = true
				}
			}

			ind := helper.findIndLastCommit(numberProposedCarries)
			if ind != -1 && numberProposedCarries < numberCarries {
				carry, _ = carries.GetCarry(numberProposedCarries)
				consensusers[ind].Propose(*carry)
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
			if LocalDispatchers[ind].proceedRandomMessage(generator, 0) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if helper.CheckSafety() == false {
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
	conf := NewMasterlessConfiguration(3)
	carries := cont.NewCarriesN(1)
	LocalDispatchers := NewLocalDispatchers(3, conf, t)
	dsp1 := LocalDispatchers[0]
	dsp2 := LocalDispatchers[1]
	dsp3 := LocalDispatchers[2]

	helper := newTestCommitHelper(3, carries.GetElementaryCarries(), LocalDispatchers)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cm3 := NewTestLocalCommitter(2, helper)

	cons1 := NewCalmConsensuser(dsp1, conf, 0)
	cons2 := NewCalmConsensuser(dsp2, conf, 1)
	cons3 := NewCalmConsensuser(dsp3, conf, 2)
	LocalDispatchers[0].cons = cons1
	LocalDispatchers[0].committer = cm1
	LocalDispatchers[1].cons = cons2
	LocalDispatchers[1].committer = cm2
	LocalDispatchers[2].cons = cons3
	LocalDispatchers[2].committer = cm3

	carry, _ := carries.GetCarry(0)
	cons1.Propose(*carry)
	LocalDispatchers[0].proceedFirstMessage(1)
	LocalDispatchers[0].proceedFirstMessage(2)
	LocalDispatchers[0].Stop()
	cons2.OnDisconnect(0)
	cons3.OnDisconnect(0)

	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)
	LocalDispatchers[1].proceedFirstMessage(2)
	LocalDispatchers[2].proceedFirstMessage(1)

	if helper.CheckSafety() == false {
		t.Error("Carry isn't committed")
	}
}

type Probabilities struct {
	probDisconnect	float32
	probSwap		float32
}

func DisconnectNode(subsetDisconnectedNodes *cont.Set, indDisconnect uint32, LocalDispatchers []*TestLocalDispatcher,
					numberNodes int, consensusers []*CalmConsensuser) {
	subsetDisconnectedNodes.Erase(indDisconnect)

	LocalDispatchers[indDisconnect].Stop()
	for ind := 0; ind < numberNodes; ind++ {
		if LocalDispatchers[ind].IsRunning() {
			consensusers[ind].OnDisconnect(cont.NodeId(indDisconnect))
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

	conf := NewMasterlessConfiguration(uint32(numberNodes))
	carries := cont.NewCarriesN(numberCarries)
	LocalDispatchers := NewLocalDispatchers(numberNodes, conf, t)

	subsetDisconnectedNodes := cont.NewRandomSubset(conf.Info, numberDisconnects, generator)
	helper := newTestCommitHelper(numberNodes, carries.GetElementaryCarries(), LocalDispatchers)
	consensusers := []*CalmConsensuser{}
	for ind := 0; ind < numberNodes; ind++ {
		cm := NewTestLocalCommitter(cont.NodeId(ind), helper)
		cons := NewCalmConsensuser(LocalDispatchers[ind], conf, ind)
		LocalDispatchers[ind].cons = cons
		LocalDispatchers[ind].committer = cm
		consensusers = append(consensusers, cons)
	}

	carry, _ := carries.GetCarry(0)
	consensusers[0].Propose(*carry)

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
					DisconnectNode(&subsetDisconnectedNodes, indDisconnect, LocalDispatchers, numberNodes, consensusers)
				}
			}

			for ind := 0; ind < numberNodes; ind++ {
				if LocalDispatchers[ind].proceedRandomMessage(generator, prob.probSwap) == true {
					flag = true
				}
			}

			if numberProposedCarries < numberCarries {
				ind := helper.findIndLastCommit(numberProposedCarries)
				if ind != -1 {
					carry, _ = carries.GetCarry(numberProposedCarries)
					consensusers[ind].Propose(*carry)
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
			if LocalDispatchers[ind].proceedRandomMessage(generator, prob.probSwap) == true {
				flag = true
			}
		}

		if flag == false {
			break
		}
	}

	if helper.CheckSafety() == false {
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
	prob := Probabilities{0.05,0}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(5, 100, 2, seed, prob, t)
	}
}

func TestRandomDisconnect10_10(t *testing.T) {
	prob := Probabilities{0.05,0}
	for seed := int64(1); seed <= 42; seed++ {
		RunRandomDisconnectTest(10, 10, 4, seed, prob, t)
	}
}

func TestRandomDisconnect10_100(t *testing.T) {
	prob := Probabilities{0.05,0}
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
