import { IStrategy } from "./IStrategy";
import { SMAStrategy } from "./implementations/SMAStrategy";
import { RSIStrategy } from "./implementations/RSIStrategy";
import { MACDStrategy } from "./implementations/MACDStrategy";
import { GridStrategy } from "./implementations/GridStrategy";
import { CointradeAdapter } from "../../modules/cointrade/CointradeAdapter";

type StrategyConstructor = new () => IStrategy;

export class StrategyFactory {
    private static registry: Map<string, StrategyConstructor> = new Map();

    static {
        // Register default strategies
        this.register(SMAStrategy);
        this.register(RSIStrategy);
        this.register(MACDStrategy);
        this.register(GridStrategy);
        this.register(CointradeAdapter);
    }

    public static register(StrategyClass: StrategyConstructor): void {
        const instance = new StrategyClass();
        this.registry.set(instance.name, StrategyClass);
        // Also allow lookup by Class Name if name property differs
        if (instance.name !== StrategyClass.name) {
             this.registry.set(StrategyClass.name, StrategyClass);
        }
    }

    public static get(name: string): IStrategy | undefined {
        const StrategyClass = this.registry.get(name);
        if (StrategyClass) {
            return new StrategyClass();
        }
        return undefined;
    }

    public static getAll(): IStrategy[] {
        // Return instances for listing
        // Use a Set to avoid duplicates if registered twice
        const uniqueClasses = new Set(this.registry.values());
        return Array.from(uniqueClasses).map(C => new C());
    }
}
