export interface IExchangeConnector {
    name: string;

    // Public API
    fetchTicker(pair: string): Promise<number>;
    fetchOrderBook(pair: string): Promise<any>;
    fetchOHLCV(pair: string, interval: string, limit?: number): Promise<any[]>;

    // Private API
    fetchBalance(): Promise<any>;
    createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any>;
    cancelOrder(id: string, pair: string): Promise<boolean>;
    fetchOrder(id: string, pair: string): Promise<any>;
    fetchOpenOrders(pair?: string): Promise<any[]>;
}

export interface IConnectorManager {
    getConnector(name: string): IExchangeConnector;
    registerConnector(connector: IExchangeConnector): void;
}
