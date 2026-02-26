import { IStrategy } from "../IStrategy";

export class GridStrategy implements IStrategy {
    name = "GridStrategy";
    interval = "5m";

    // Grid Parameters
    private lowerPrice = 20000;
    private upperPrice = 25000;
    private gridLines = 10;
    private amountPerGrid = 100; // USD

    setParameters(params: any): void {
        if (params.lowerPrice) this.lowerPrice = params.lowerPrice;
        if (params.upperPrice) this.upperPrice = params.upperPrice;
        if (params.gridLines) this.gridLines = params.gridLines;
        if (params.amountPerGrid) this.amountPerGrid = params.amountPerGrid;
    }

    async populateIndicators(candles: any[]): Promise<any[]> {
        const gridStep = (this.upperPrice - this.lowerPrice) / this.gridLines;
        const levels: number[] = [];
        for(let i=0; i<=this.gridLines; i++) {
            levels.push(this.lowerPrice + (i * gridStep));
        }

        return candles.map(c => ({
            ...c,
            grid_levels: levels,
            grid_step: gridStep
        }));
    }

    async populateBuyTrend(candles: any[]): Promise<any[]> {
        const gridStep = (this.upperPrice - this.lowerPrice) / this.gridLines;

        // Buy if price crosses DOWN a grid line
        for (let i = 1; i < candles.length; i++) {
            const curr = candles[i].close;
            const prev = candles[i-1].close;

            let signal = false;
            // Iterate levels to check crossover
            for(let j=0; j<=this.gridLines; j++) {
                const level = this.lowerPrice + (j * gridStep);
                // Crossing Down
                if (prev > level && curr <= level) {
                    signal = true;
                    break;
                }
            }
            candles[i].buy_signal = signal;
        }
        return candles;
    }

    async populateSellTrend(candles: any[]): Promise<any[]> {
        const gridStep = (this.upperPrice - this.lowerPrice) / this.gridLines;

        // Sell if price crosses UP a grid line
        for (let i = 1; i < candles.length; i++) {
            const curr = candles[i].close;
            const prev = candles[i-1].close;

            let signal = false;
            // Iterate levels to check crossover
            for(let j=0; j<=this.gridLines; j++) {
                const level = this.lowerPrice + (j * gridStep);
                // Crossing Up
                if (prev < level && curr >= level) {
                    signal = true;
                    break;
                }
            }
            candles[i].sell_signal = signal;
        }
        return candles;
    }
}
