import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";
import { KuCoinConnector } from "../../exchanges/KuCoinConnector";

export class PaperExchange implements IExchangeConnector {
    name = "Paper";
    private balance: Map<string, number> = new Map([['USD', 10000]]);
    private orders: any[] = [];
    private dataConnector: IExchangeConnector;

    constructor(initialBalance: number = 10000) {
        this.balance.set('USD', initialBalance);
        this.dataConnector = new KuCoinConnector(); // Use real data for paper trading
    }

    async fetchTicker(pair: string): Promise<number> {
        return this.dataConnector.fetchTicker(pair);
    }

    async fetchOrderBook(pair: string): Promise<any> {
        return this.dataConnector.fetchOrderBook(pair);
    }

    async fetchOHLCV(pair: string, interval: string, limit?: number): Promise<any[]> {
        return this.dataConnector.fetchOHLCV(pair, interval, limit);
    }

    async fetchBalance(): Promise<any> {
        const result: any = {};
        for (const [currency, amount] of this.balance.entries()) {
            result[currency] = amount;
        }
        return result;
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        const currentPrice = price || await this.fetchTicker(pair);
        // Normalize pair if needed, assuming XXX-USD
        const parts = pair.split('-');
        const base = parts[0];
        const quote = parts[1] || 'USD';

        // Simple fee simulation (0.1%)
        const feeRate = 0.001;

        if (side === 'buy') {
            const cost = amount * currentPrice;
            const fees = cost * feeRate;
            const totalCost = cost + fees;

            const currentQuote = this.balance.get(quote) || 0;
            if (currentQuote < totalCost) {
                // For paper trading, maybe log warning but allow partial if configured? No, strict checks.
                throw new Error(`[Paper] Insufficient funds: needed ${totalCost.toFixed(2)} ${quote}, have ${currentQuote.toFixed(2)}`);
            }

            this.balance.set(quote, currentQuote - totalCost);
            this.balance.set(base, (this.balance.get(base) || 0) + amount);
        } else {
            const currentBase = this.balance.get(base) || 0;
            if (currentBase < amount) {
                throw new Error(`[Paper] Insufficient funds: needed ${amount} ${base}, have ${currentBase}`);
            }

            const proceeds = amount * currentPrice;
            const fees = proceeds * feeRate;
            const totalProceeds = proceeds - fees;

            this.balance.set(base, currentBase - amount);
            this.balance.set(quote, (this.balance.get(quote) || 0) + totalProceeds);
        }

        const id = Math.random().toString(36).substring(7);
        const order = {
            id,
            pair,
            type,
            side,
            amount,
            price: currentPrice,
            status: 'closed', // Instant fill
            timestamp: Date.now()
        };
        this.orders.push(order);

        console.log(`[Paper] Executed: ${side.toUpperCase()} ${amount} ${base} @ ${currentPrice} (Fee: ${(amount * currentPrice * feeRate).toFixed(2)})`);
        return order;
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        // Since orders are instantly filled, cannot cancel.
        return false;
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        return this.orders.find(o => o.id === id);
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        return this.orders.filter(o => o.status === 'open');
    }
}
