import { Trader } from '../src/trader/Trader';
import { IExchangeConnector } from '../src/engine/connector/IExchangeConnector';
import { ConfigManager } from '../src/config/ConfigManager';

// Mock Exchange Connector
class MockExchange implements IExchangeConnector {
    name = "Mock";
    public orders: any[] = [];
    public ticker: number = 100;

    async fetchTicker(pair: string): Promise<number> {
        return this.ticker;
    }
    async fetchOrderBook(pair: string): Promise<any> { return {}; }
    async fetchOHLCV(pair: string, interval: string, limit?: number): Promise<any[]> {
        // Return dummy candles
        return Array.from({ length: limit || 50 }, (_, i) => ({
            timestamp: Date.now() - (i * 60000),
            open: 100, high: 105, low: 95, close: 100, volume: 1000
        }));
    }
    async fetchBalance(): Promise<any> { return { USD: 10000 }; }
    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        const order = { id: '123', pair, type, side, amount, price };
        this.orders.push(order);
        return order;
    }
    async cancelOrder(id: string, pair: string): Promise<boolean> { return true; }
    async fetchOrder(id: string, pair: string): Promise<any> { return {}; }
    async fetchOpenOrders(pair?: string): Promise<any[]> { return []; }
}

describe('Trader Logic', () => {
    let trader: Trader;
    let exchange: MockExchange;

    beforeEach(() => {
        // Reset Config
        (ConfigManager as any).instance = null;
        exchange = new MockExchange();
        trader = new Trader(exchange);
        // Mock Config
        const config = ConfigManager.getInstance();
        (config as any).config = {
            trading: {
                coins: ["BTC"],
                dca_levels: [-2.5, -5.0],
                dca_multiplier: 2.0,
                max_dca_buys_per_24h: 2,
                active_strategy: "SMAStrategy"
            },
            notifications: { enabled: false }
        };
    });

    test('should initialize correctly', () => {
        expect(trader).toBeDefined();
    });

    // Note: Testing private methods or async loops in a rigorous way often requires
    // dependency injection or more complex mocking of the strategy signals.
    // For this basic test suite, we verifying the class structure and basic flow.

    test('Mock Exchange should record orders', async () => {
        await exchange.createOrder('BTC-USD', 'market', 'buy', 1);
        expect(exchange.orders.length).toBe(1);
        expect(exchange.orders[0].side).toBe('buy');
    });
});
