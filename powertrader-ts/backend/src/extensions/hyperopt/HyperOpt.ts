import { BacktestEngine } from '../../engine/backtest/BacktestEngine';
import { IBacktestConfig } from '../../engine/backtest/IBacktestEngine';
import { StrategyFactory } from '../../engine/strategy/StrategyFactory';
import { HistoricalData } from '../../engine/backtest/HistoricalData';

export interface HyperOptConfig {
    strategyName: string;
    pair: string;
    timeframe: string;
    startDate: number;
    endDate: number;
    populationSize: number;
    generations: number;
    parameterSpace: Record<string, { min: number, max: number, step?: number }>;
}

export class HyperOpt {
    public async optimize(config: HyperOptConfig): Promise<any> {
        console.log(`Starting HyperOpt for ${config.strategyName} on ${config.pair}...`);

        const strategyTemplate = StrategyFactory.get(config.strategyName);
        if (!strategyTemplate || !strategyTemplate.setParameters) {
            throw new Error(`Strategy ${config.strategyName} does not support parameter optimization.`);
        }

        // 1. Pre-fetch Data
        console.log(`[HyperOpt] Pre-fetching data...`);
        const candles = await HistoricalData.fetch(config.pair, config.timeframe, config.startDate, config.endDate);
        if (candles.length === 0) throw new Error("No data for optimization");
        console.log(`[HyperOpt] Loaded ${candles.length} candles.`);

        // 2. Initialize Population
        let population = this.generatePopulation(config.populationSize, config.parameterSpace);

        const engine = new BacktestEngine();
        const backtestConfig: IBacktestConfig = {
            strategy: config.strategyName,
            pair: config.pair,
            startDate: config.startDate,
            endDate: config.endDate,
            initialBalance: 10000,
            timeframe: config.timeframe,
            data: candles
        };

        let bestResult = null;
        let bestParams = null;

        // Track stats
        const evolution: any[] = [];

        for (let gen = 0; gen < config.generations; gen++) {
            // console.log(`Generation ${gen + 1}/${config.generations}`);
            const results: { params: any, score: number }[] = [];

            // 3. Evaluate Fitness
            for (const params of population) {
                // Reuse strategy instance
                strategyTemplate.setParameters!(params);

                try {
                    const result = await engine.run(backtestConfig, strategyTemplate);

                    // Fitness Function: Total Profit (if trades > 5)
                    let score = -10000;
                    const finalEquity = result.equityCurve[result.equityCurve.length-1].value;
                    const profit = finalEquity - 10000;

                    if (result.totalTrades > 5) {
                        score = profit;
                    } else {
                        // Penalty for too few trades
                        score = -5000;
                    }

                    results.push({ params, score });

                    if (!bestResult || score > bestResult.score) {
                        bestResult = { score, ...result };
                        bestParams = params;
                    }
                } catch (e) {
                    results.push({ params, score: -10000 });
                }
            }

            // 4. Selection & Stats
            results.sort((a, b) => b.score - a.score);
            const bestScore = results[0].score;
            const avgScore = results.reduce((a, b) => a + b.score, 0) / results.length;

            console.log(`[HyperOpt] Generation ${gen + 1}: Best Score = ${bestScore.toFixed(2)}, Avg = ${avgScore.toFixed(2)}`);
            evolution.push({ generation: gen + 1, bestScore, avgScore });

            // Elitism: Keep top 20%
            const survivors = results.slice(0, Math.floor(config.populationSize * 0.2)).map(r => r.params);

            // Crossover / Mutation
            const offspring: any[] = [];
            while (offspring.length < config.populationSize - survivors.length) {
                const parent1 = survivors[Math.floor(Math.random() * survivors.length)];
                const parent2 = survivors[Math.floor(Math.random() * survivors.length)];
                offspring.push(this.crossover(parent1, parent2, config.parameterSpace));
            }

            population = [...survivors, ...offspring];

            // Mutate (10% rate)
            population = population.map(p => this.mutate(p, config.parameterSpace, 0.1));
        }

        return {
            bestParams,
            bestScore: bestResult?.score,
            evolution,
            bestResult: {
                totalTrades: bestResult?.totalTrades,
                winRate: bestResult?.winRate,
                profitFactor: bestResult?.profitFactor,
                maxDrawdown: bestResult?.maxDrawdown,
                equityCurve: bestResult?.equityCurve // Include equity curve for visualization
            }
        };
    }

    private generatePopulation(size: number, space: any): any[] {
        const pop = [];
        for (let i = 0; i < size; i++) {
            const ind: any = {};
            for (const key in space) {
                const range = space[key];
                const val = range.min + Math.random() * (range.max - range.min);
                ind[key] = range.step ? Math.round(val / range.step) * range.step : Math.round(val);
            }
            pop.push(ind);
        }
        return pop;
    }

    private crossover(p1: any, p2: any, space: any): any {
        const child: any = {};
        for (const key in space) {
            child[key] = Math.random() > 0.5 ? p1[key] : p2[key];
        }
        return child;
    }

    private mutate(ind: any, space: any, rate: number): any {
        if (Math.random() > rate) return ind;
        const mutated = { ...ind };
        // Pick one gene to mutate
        const keys = Object.keys(space);
        const key = keys[Math.floor(Math.random() * keys.length)];
        const range = space[key];
        const val = range.min + Math.random() * (range.max - range.min);
        mutated[key] = range.step ? Math.round(val / range.step) * range.step : Math.round(val);
        return mutated;
    }
}
