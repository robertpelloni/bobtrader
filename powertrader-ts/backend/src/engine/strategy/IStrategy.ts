export interface IStrategy {
    name: string;
    interval: string;

    // Core Signals
    populateIndicators(dataframe: any): Promise<any>;
    populateBuyTrend(dataframe: any): Promise<any>;
    populateSellTrend(dataframe: any): Promise<any>;

    // Risk Management
    stopLoss?: number;
    trailingStop?: boolean;
    trailingStopPositive?: number;
    trailingStopPositiveOffset?: number;

    // Position Sizing
    positionSize?: number; // 0.0 - 1.0 (percent of wallet)
}

export interface IStrategyManager {
    loadStrategy(name: string): Promise<IStrategy>;
    getAvailableStrategies(): string[];
    validateStrategy(strategy: IStrategy): boolean;
}
