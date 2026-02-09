import { IStrategy } from "../IStrategy";
import { SMAStrategy } from "./implementations/SMAStrategy";
import { CointradeAdapter } from "../../modules/cointrade/CointradeAdapter";

export class StrategyFactory {
    private static strategies: Map<string, IStrategy> = new Map();

    static {
        // Register default strategies
        this.register(new SMAStrategy());
        this.register(new CointradeAdapter());
    }

    public static register(strategy: IStrategy): void {
        this.strategies.set(strategy.name, strategy);
    }

    public static get(name: string): IStrategy | undefined {
        return this.strategies.get(name);
    }

    public static getAll(): IStrategy[] {
        return Array.from(this.strategies.values());
    }
}
