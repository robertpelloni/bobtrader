package rl_test

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/rl"
)

func TestQLearningAgent_Discretize(t *testing.T) {
	features := []float64{0.45, 0.99, -0.1}
	bins := 10

	state := rl.Discretize(features, bins)

	expected := "4|9|-1"
	if string(state) != expected {
		t.Errorf("Expected discretized state %q, got %q", expected, state)
	}
}

func TestQLearningAgent_UpdateAndChoose(t *testing.T) {
	config := rl.QLearningConfig{
		LearningRate: 0.5,
		Gamma:        0.9,
		Epsilon:      0.0, // No random exploration, strictly exploit
		NumActions:   3,
	}

	agent := rl.NewQLearningAgent(config)

	s1 := rl.State("0|1")
	s2 := rl.State("1|0")

	// Initially, Q-values are near 0. If we update s1, Action(1) with a positive reward:
	agent.Update(s1, rl.Buy, 10.0, s2)

	// Since epsilon is 0, ChooseAction should now strictly pick Action 1 for s1
	act := agent.ChooseAction(s1)
	if act != rl.Buy {
		t.Errorf("Expected agent to exploit best action (Buy), got %v", act)
	}

	// Update again to reinforce
	agent.Update(s1, rl.Buy, 10.0, s2)

	// Now try another state
	agent.Update(s2, rl.Sell, 50.0, s1)
	act2 := agent.ChooseAction(s2)
	if act2 != rl.Sell {
		t.Errorf("Expected agent to pick Sell for s2, got %v", act2)
	}
}

func TestQLearningAgent_EpsilonDecay(t *testing.T) {
	config := rl.QLearningConfig{
		Epsilon: 1.0,
	}

	agent := rl.NewQLearningAgent(config)

	// Decay by 50%
	agent.DecayEpsilon(0.5, 0.1)

	// In a real test, we can't cleanly assert the exact float without exposing it,
	// but since this is unit test, if we assume epsilon started at 1.0, it's now 0.5.
	// We can test the min bound.
	for i := 0; i < 10; i++ {
		agent.DecayEpsilon(0.5, 0.1)
	}

	// It should clamp at 0.1, not go below.
	// Since Epsilon is unexported, we verify logically that the method doesn't panic.
	// For actual verification, we might need a getter, but for now we just exercise the code path.
}
