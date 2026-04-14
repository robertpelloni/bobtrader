package rl

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Action represents a discrete trading decision.
type Action int

const (
	Hold Action = 0
	Buy  Action = 1
	Sell Action = 2
)

// State is an abstract representation of the market.
// It is typically a discretized string of concatenated features (e.g., "RSI:High,MACD:Bull").
type State string

// QLearningAgent implements a tabular reinforcement learning agent.
// It uses the Bellman equation to learn optimal policies.
type QLearningAgent struct {
	qTable   map[State][]float64
	nActions int
	lr       float64 // Learning rate (alpha)
	gamma    float64 // Discount factor
	epsilon  float64 // Exploration rate
	rng      *rand.Rand
	mu       sync.RWMutex
}

// QLearningConfig defines the hyperparameters for the RL agent.
type QLearningConfig struct {
	LearningRate float64
	Gamma        float64
	Epsilon      float64
	NumActions   int // Defaults to 3 (Hold, Buy, Sell)
}

// NewQLearningAgent creates a new tabular Q-Learning agent.
func NewQLearningAgent(config QLearningConfig) *QLearningAgent {
	if config.NumActions <= 0 {
		config.NumActions = 3
	}
	if config.LearningRate <= 0 {
		config.LearningRate = 0.1
	}
	if config.Gamma <= 0 {
		config.Gamma = 0.95
	}

	return &QLearningAgent{
		qTable:   make(map[State][]float64),
		nActions: config.NumActions,
		lr:       config.LearningRate,
		gamma:    config.Gamma,
		epsilon:  config.Epsilon,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ensureState initialized the Q-values for a new state.
func (a *QLearningAgent) ensureState(s State) {
	if _, ok := a.qTable[s]; !ok {
		a.qTable[s] = make([]float64, a.nActions)
		// Small random initialization can help break ties
		for i := 0; i < a.nActions; i++ {
			a.qTable[s][i] = (a.rng.Float64() - 0.5) * 0.01 // between -0.005 and 0.005
		}
	}
}

// ChooseAction selects an action using an epsilon-greedy policy.
func (a *QLearningAgent) ChooseAction(s State) Action {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ensureState(s)

	// Explore
	if a.rng.Float64() < a.epsilon {
		return Action(a.rng.Intn(a.nActions))
	}

	// Exploit
	return a.bestAction(s)
}

// bestAction returns the action with the highest Q-value for a given state.
func (a *QLearningAgent) bestAction(s State) Action {
	qVals := a.qTable[s]
	maxQ := qVals[0]
	bestA := 0

	for i := 1; i < a.nActions; i++ {
		if qVals[i] > maxQ {
			maxQ = qVals[i]
			bestA = i
		}
	}

	return Action(bestA)
}

// Update Q-value using the Bellman equation:
// Q(s,a) = Q(s,a) + lr * (Reward + gamma * maxQ(s') - Q(s,a))
func (a *QLearningAgent) Update(s State, act Action, reward float64, nextS State) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ensureState(s)
	a.ensureState(nextS)

	oldQ := a.qTable[s][act]
	nextMaxQ := a.qTable[nextS][a.bestAction(nextS)]

	newQ := oldQ + a.lr*(reward+(a.gamma*nextMaxQ)-oldQ)
	a.qTable[s][act] = newQ
}

// DecayEpsilon reduces the exploration rate over time.
func (a *QLearningAgent) DecayEpsilon(decayRate, minEpsilon float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.epsilon *= decayRate
	if a.epsilon < minEpsilon {
		a.epsilon = minEpsilon
	}
}

// Discretize converts a continuous feature array into a discrete state string.
// This matches the Python implementation logic of converting floats to bins.
func Discretize(features []float64, bins int) State {
	var parts []string
	for _, f := range features {
		// e.g., if f is 0.45 and bins is 10, bin is 4
		bin := int(f * float64(bins))
		parts = append(parts, fmt.Sprintf("%d", bin))
	}
	return State(strings.Join(parts, "|"))
}
