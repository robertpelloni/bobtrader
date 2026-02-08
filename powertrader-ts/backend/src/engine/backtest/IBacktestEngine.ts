import { IStrategy } from "../strategy/IStrategy";

export interface IBacktestConfig {
    strategy: string;
    pair: string;
    startDate: number;
    endDate: number;
    initialBalance: number;
    timeframe: string;
}

export interface IBacktestResult {
    totalTrades: number;
    winRate: number;
    profitFactor: number;
    maxDrawdown: number;
    sharpeRatio: number;
    trades: any[];
    equityCurve: { time: number, value: number }[];
}

export interface IBacktestEngine {
    run(config: IBacktestConfig, strategy: IStrategy): Promise<IBacktestResult>;
}
