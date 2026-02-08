import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";

export class PaperExchange implements IExchangeConnector {
    name = "Paper";
    private balance: any = { USD: 10000 };
    private orders: any[] = [];

    async fetchTicker(pair: string): Promise<number> {
        return 100 + Math.random() * 10; // Mock price
    }

    async fetchOrderBook(pair: string): Promise<any> {
        return { bids: [], asks: [] };
    }

    async fetchOHLCV(pair: string, interval: string, limit?: number): Promise<any[]> {
        return [];
    }

    async fetchBalance(): Promise<any> {
        return this.balance;
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        const id = Math.random().toString(36).substring(7);
        const order = { id, pair, type, side, amount, price, status: 'open' };
        this.orders.push(order);
        console.log(`[Paper] Created order: ${side} ${amount} ${pair} @ ${price}`);
        return order;
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        this.orders = this.orders.filter(o => o.id !== id);
        return true;
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        return this.orders.find(o => o.id === id);
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        return this.orders;
    }
}
