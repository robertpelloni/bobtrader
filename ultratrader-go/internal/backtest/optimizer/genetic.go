package optimizer

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
)

// GeneticOptimizer evolves strategy parameters using a genetic algorithm.
type GeneticOptimizer struct {
	engine     *backtest.Engine
	history    backtest.HistoryProvider
	candles    backtest.CandleHistoryProvider
	builder    StrategyBuilder
	scorer     ScoringFunction
	population int
	generations int
	mutation    float64
}

func NewGeneticOptimizer(engine *backtest.Engine, h backtest.HistoryProvider, c backtest.CandleHistoryProvider, builder StrategyBuilder, scorer ScoringFunction) *GeneticOptimizer {
	if scorer == nil {
		scorer = DefaultScorer
	}
	return &GeneticOptimizer{
		engine:      engine,
		history:     h,
		candles:     c,
		builder:     builder,
		scorer:      scorer,
		population:  20,
		generations: 5,
		mutation:    0.1,
	}
}

// GeneticParamRange defines the search space for a single parameter.
type GeneticParamRange struct {
	Min  float64
	Max  float64
	Step float64
	IsInt bool
}

// GeneticOptimize runs the genetic evolution across the parameter space.
func (o *GeneticOptimizer) Optimize(ctx context.Context, ranges map[string]GeneticParamRange) ([]RunResult, error) {
	// Initialize random population
	pop := make([]ParameterMap, o.population)
	for i := 0; i < o.population; i++ {
		pop[i] = o.randomParams(ranges)
	}

	var allResults []RunResult

	for gen := 0; gen < o.generations; gen++ {
		results, err := o.evaluatePopulation(ctx, pop)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)

		// Sort by score descending
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		// Select top 50% for next generation
		elites := results[:o.population/2]

		newPop := make([]ParameterMap, 0, o.population)
		// Carry over elites
		for _, e := range elites {
			newPop = append(newPop, e.Params)
		}

		// Crossover and mutate to fill rest of population
		for len(newPop) < o.population {
			p1 := elites[rand.Intn(len(elites))].Params
			p2 := elites[rand.Intn(len(elites))].Params
			child := o.crossover(p1, p2)
			o.mutate(child, ranges)
			newPop = append(newPop, child)
		}
		pop = newPop
	}

	return allResults, nil
}

func (o *GeneticOptimizer) evaluatePopulation(ctx context.Context, pop []ParameterMap) ([]RunResult, error) {
	results := make([]RunResult, len(pop))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, params := range pop {
		wg.Add(1)
		go func(idx int, p ParameterMap) {
			defer wg.Done()

			strat, err := o.builder(p)
			if err != nil {
				mu.Lock()
				if firstErr == nil { firstErr = err }
				mu.Unlock()
				return
			}

			// Run backtest based on strategy type
			var res backtest.Result
			var bErr error

			// We need a way to run the engine with a specific strategy.
			// Since Engine has a fixed strategy from NewEngine, we'll clone it or assume it's fresh.
			// In our current Engine implementation, the strategy is fixed.
			// We'll create a temporary engine for each run.
			tempEngine := backtest.NewEngineWithOptions(strat, 10000, backtest.DefaultEmulatorOptions())

			if o.candles != nil {
				res, bErr = tempEngine.RunCandles(ctx, o.candles)
			} else if o.history != nil {
				res, bErr = tempEngine.RunTicks(ctx, o.history)
			} else {
				bErr = fmt.Errorf("no history provider")
			}

			if bErr != nil {
				mu.Lock()
				if firstErr == nil { firstErr = bErr }
				mu.Unlock()
				return
			}

			mu.Lock()
			results[idx] = RunResult{
				Params: p,
				Result: res,
				Score:  o.scorer(res),
			}
			mu.Unlock()
		}(i, params)
	}

	wg.Wait()
	return results, firstErr
}

func (o *GeneticOptimizer) randomParams(ranges map[string]GeneticParamRange) ParameterMap {
	p := make(ParameterMap)
	for name, r := range ranges {
		val := r.Min + rand.Float64()*(r.Max-r.Min)
		if r.Step > 0 {
			val = float64(int(val/r.Step)) * r.Step
		}
		if r.IsInt {
			p[name] = int(val)
		} else {
			p[name] = val
		}
	}
	return p
}

func (o *GeneticOptimizer) crossover(p1, p2 ParameterMap) ParameterMap {
	child := make(ParameterMap)
	for name, val := range p1 {
		if rand.Float64() > 0.5 {
			child[name] = val
		} else {
			child[name] = p2[name]
		}
	}
	return child
}

func (o *GeneticOptimizer) mutate(p ParameterMap, ranges map[string]GeneticParamRange) {
	for name := range p {
		if rand.Float64() < o.mutation {
			r := ranges[name]
			val := r.Min + rand.Float64()*(r.Max-r.Min)
			if r.Step > 0 {
				val = float64(int(val/r.Step)) * r.Step
			}
			if r.IsInt {
				p[name] = int(val)
			} else {
				p[name] = val
			}
		}
	}
}
