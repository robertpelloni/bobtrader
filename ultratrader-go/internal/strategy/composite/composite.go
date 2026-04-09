package composite

import (
	"context"
	"fmt"
	"sync"
)

// Signal represents a trading signal from a strategy.
type Signal int

const (
	SignalNone   Signal = iota
	SignalBuy
	SignalSell
)

func (s Signal) String() string {
	switch s {
	case SignalBuy:
		return "BUY"
	case SignalSell:
		return "SELL"
	case SignalNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

// Confidence represents signal strength from 0.0 to 1.0.
type Confidence float64

const (
	ConfidenceLow      Confidence = 0.25
	ConfidenceMedium   Confidence = 0.50
	ConfidenceHigh     Confidence = 0.75
	ConfidenceAbsolute  Confidence = 1.00
)

// SignalResult is the output of a signal evaluation.
type SignalResult struct {
	Signal      Signal
	Confidence  Confidence
	Source      string // Strategy name
	Reason      string // Human-readable explanation
}

// SignalEvaluator is the interface for strategies that produce signals.
type SignalEvaluator interface {
	Name() string
	Evaluate(ctx context.Context) (SignalResult, error)
}

// ResolutionMode determines how conflicting signals are resolved.
type ResolutionMode int

const (
	// Unanimous requires all strategies to agree.
	Unanimous ResolutionMode = iota
	// Majority requires more than half to agree.
	Majority
	// Any fires if any strategy produces a signal.
	Any
	// Weighted uses weighted voting based on strategy weights.
	Weighted
)

// CompositeStrategy combines multiple sub-strategies using voting.
type CompositeStrategy struct {
	mu        sync.RWMutex
	name      string
	evaluators []weightedEvaluator
	mode      ResolutionMode
}

type weightedEvaluator struct {
	evaluator SignalEvaluator
	weight    float64
}

// NewCompositeStrategy creates a new composite strategy.
func NewCompositeStrategy(name string, mode ResolutionMode) *CompositeStrategy {
	return &CompositeStrategy{
		name: name,
		mode: mode,
	}
}

// AddStrategy adds a sub-strategy with a weight (used in Weighted mode).
func (c *CompositeStrategy) AddStrategy(evaluator SignalEvaluator, weight float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evaluators = append(c.evaluators, weightedEvaluator{
		evaluator: evaluator,
		weight:    weight,
	})
}

// Name returns the composite strategy name.
func (c *CompositeStrategy) Name() string {
	return c.name
}

// Evaluate runs all sub-strategies and resolves the composite signal.
func (c *CompositeStrategy) Evaluate(ctx context.Context) (SignalResult, error) {
	c.mu.RLock()
	evaluators := make([]weightedEvaluator, len(c.evaluators))
	copy(evaluators, c.evaluators)
	c.mu.RUnlock()

	if len(evaluators) == 0 {
		return SignalResult{Signal: SignalNone, Source: c.name}, nil
	}

	// Collect signals from all evaluators
	results := make([]SignalResult, len(evaluators))
	errs := make([]error, len(evaluators))

	for i, we := range evaluators {
		result, err := we.evaluator.Evaluate(ctx)
		results[i] = result
		errs[i] = err
	}

	// Check for critical errors
	var errCount int
	for _, e := range errs {
		if e != nil {
			errCount++
		}
	}
	if errCount == len(evaluators) {
		return SignalResult{Signal: SignalNone, Source: c.name},
			fmt.Errorf("all %d evaluators failed", errCount)
	}

	// Resolve based on mode
	switch c.mode {
	case Unanimous:
		return c.resolveUnanimous(results), nil
	case Majority:
		return c.resolveMajority(results), nil
	case Any:
		return c.resolveAny(results), nil
	case Weighted:
		return c.resolveWeighted(results, evaluators), nil
	default:
		return c.resolveMajority(results), nil
	}
}

func (c *CompositeStrategy) resolveUnanimous(results []SignalResult) SignalResult {
	if len(results) == 0 {
		return SignalResult{Signal: SignalNone, Source: c.name}
	}

	first := results[0].Signal
	if first == SignalNone {
		return SignalResult{Signal: SignalNone, Source: c.name}
	}

	for _, r := range results[1:] {
		if r.Signal != first {
			return SignalResult{Signal: SignalNone, Source: c.name}
		}
	}

	// All agree — compute average confidence
	var totalConf float64
	for _, r := range results {
		totalConf += float64(r.Confidence)
	}

	return SignalResult{
		Signal:     first,
		Confidence: Confidence(totalConf / float64(len(results))),
		Source:     c.name,
		Reason:     fmt.Sprintf("unanimous %s from %d strategies", first, len(results)),
	}
}

func (c *CompositeStrategy) resolveMajority(results []SignalResult) SignalResult {
	votes := map[Signal]int{}
	var totalConf map[Signal]float64 = make(map[Signal]float64)

	for _, r := range results {
		if r.Signal != SignalNone {
			votes[r.Signal]++
			totalConf[r.Signal] += float64(r.Confidence)
		}
	}

	var winner Signal
	var maxVotes int
	for sig, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winner = sig
		}
	}

	// Need more than half of non-NONE votes
	nonNone := 0
	for _, r := range results {
		if r.Signal != SignalNone {
			nonNone++
		}
	}

	if nonNone == 0 || maxVotes <= nonNone/2 {
		return SignalResult{Signal: SignalNone, Source: c.name}
	}

	return SignalResult{
		Signal:     winner,
		Confidence: Confidence(totalConf[winner] / float64(maxVotes)),
		Source:     c.name,
		Reason:     fmt.Sprintf("majority %s (%d/%d)", winner, maxVotes, nonNone),
	}
}

func (c *CompositeStrategy) resolveAny(results []SignalResult) SignalResult {
	var bestResult SignalResult
	bestResult.Signal = SignalNone

	for _, r := range results {
		if r.Signal != SignalNone && r.Confidence > bestResult.Confidence {
			bestResult = r
		}
	}

	if bestResult.Signal == SignalNone {
		return SignalResult{Signal: SignalNone, Source: c.name}
	}

	return SignalResult{
		Signal:     bestResult.Signal,
		Confidence: bestResult.Confidence,
		Source:     c.name,
		Reason:     fmt.Sprintf("any-mode: strongest signal from %s", bestResult.Source),
	}
}

func (c *CompositeStrategy) resolveWeighted(results []SignalResult, evaluators []weightedEvaluator) SignalResult {
	weightedScores := map[Signal]float64{}
	weightedConf := map[Signal]float64{}

	for i, r := range results {
		if r.Signal != SignalNone {
			w := evaluators[i].weight
			weightedScores[r.Signal] += w
			weightedConf[r.Signal] += float64(r.Confidence) * w
		}
	}

	var winner Signal
	var maxScore float64
	for sig, score := range weightedScores {
		if score > maxScore {
			maxScore = score
			winner = sig
		}
	}

	if winner == SignalNone {
		return SignalResult{Signal: SignalNone, Source: c.name}
	}

	totalWeight := weightedScores[winner]
	return SignalResult{
		Signal:     winner,
		Confidence: Confidence(weightedConf[winner] / totalWeight),
		Source:     c.name,
		Reason:     fmt.Sprintf("weighted %s (score=%.2f)", winner, maxScore),
	}
}

// StrategyCount returns the number of sub-strategies.
func (c *CompositeStrategy) StrategyCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.evaluators)
}
