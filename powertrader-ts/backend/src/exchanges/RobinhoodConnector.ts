import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import axios, { AxiosInstance } from 'axios';
import { v4 as uuidv4 } from 'uuid';
import * as nacl from 'tweetnacl';

export class RobinhoodConnector implements IExchangeConnector {
    name = "Robinhood";
    private baseUrl = "https://trading.robinhood.com";
    private apiKey: string;
    private privateKeySeed: Uint8Array;
    private axiosInstance: AxiosInstance;

    constructor(apiKey: string, privateKeyBase64: string) {
        this.apiKey = apiKey;

        // Decode base64 private key seed
        // Robinhood requires an Ed25519 keypair.
        // The 'r_secret.txt' usually contains the base64 encoded seed (32 bytes).
        const seedBuffer = Buffer.from(privateKeyBase64, 'base64');
        if (seedBuffer.length !== 32) {
            console.error(`[Robinhood] Warning: Private key seed length is ${seedBuffer.length}, expected 32 bytes.`);
        }
        this.privateKeySeed = new Uint8Array(seedBuffer);

        this.axiosInstance = axios.create({
            baseURL: this.baseUrl,
            timeout: 10000
        });
    }

    private getAuthorizationHeader(method: string, path: string, body: string, timestamp: number): any {
        const messageToSign = `${this.apiKey}${timestamp}${path}${method}${body}`;

        // Sign using tweetnacl (Ed25519)
        // We regenerate the keypair from the seed every time (or could cache the keypair)
        const keyPair = nacl.sign.keyPair.fromSeed(this.privateKeySeed);
        const messageBytes = Buffer.from(messageToSign, 'utf-8');
        const signatureBytes = nacl.sign.detached(messageBytes, keyPair.secretKey);
        const signature = Buffer.from(signatureBytes).toString('base64');

        return {
            "x-api-key": this.apiKey,
            "x-signature": signature,
            "x-timestamp": timestamp.toString(),
            "Content-Type": "application/json"
        };
    }

    private async makeRequest(method: 'GET' | 'POST', path: string, body: any = null): Promise<any> {
        const timestamp = Math.floor(Date.now() / 1000);
        const bodyStr = body ? JSON.stringify(body) : "";
        const headers = this.getAuthorizationHeader(method, path, bodyStr, timestamp);

        try {
            const response = await this.axiosInstance.request({
                method,
                url: path,
                headers,
                data: body
            });
            return response.data;
        } catch (error) {
            if (axios.isAxiosError(error) && error.response) {
                console.error(`[Robinhood] Request failed: ${error.response.status} - ${JSON.stringify(error.response.data)}`);
            } else {
                console.error(`[Robinhood] Request failed: ${error}`);
            }
            throw error;
        }
    }

    async fetchTicker(pair: string): Promise<number> {
        try {
            const res = await this.makeRequest('GET', `/api/v1/crypto/marketdata/best_bid_ask/?symbol=${pair}`);
            if (res.results && res.results.length > 0) {
                return parseFloat(res.results[0].ask_inclusive_of_buy_spread);
            }
            return 0;
        } catch (e) {
            console.error(`[Robinhood] Error fetching ticker for ${pair}:`, e);
            return 0;
        }
    }

    async fetchOrderBook(pair: string): Promise<any> {
        // Robinhood might not expose full orderbook via this API endpoint easily
        return { bids: [], asks: [] };
    }

    async fetchOHLCV(pair: string, interval: string, limit?: number): Promise<any[]> {
        // Implement historical data fetching logic
        return [];
    }

    async fetchBalance(): Promise<any> {
        try {
            const res = await this.makeRequest('GET', '/api/v1/crypto/trading/accounts/');
            // Transform RH response to standard format
            return res;
        } catch (e) {
            return {};
        }
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        const clientOrderId = uuidv4();
        const body: any = {
            client_order_id: clientOrderId,
            side: side,
            type: type,
            symbol: pair,
            market_order_config: {
                asset_quantity: amount.toFixed(8)
            }
        };

        if (type === 'limit' && price) {
            body.limit_order_config = {
                limit_price: price.toFixed(2),
                asset_quantity: amount.toFixed(8)
            };
            delete body.market_order_config;
        }

        console.log(`[Robinhood] Placing ${side} order for ${amount} ${pair}`);
        try {
            const res = await this.makeRequest('POST', '/api/v1/crypto/trading/orders/', body);
            return res;
        } catch (e) {
            console.error(`[Robinhood] Order placement failed:`, e);
            return null;
        }
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        try {
            await this.makeRequest('POST', `/api/v1/crypto/trading/orders/${id}/cancel/`);
            return true;
        } catch (e) {
            return false;
        }
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        try {
            return await this.makeRequest('GET', `/api/v1/crypto/trading/orders/${id}/`);
        } catch (e) {
            return null;
        }
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        // Implement fetching open orders
        return [];
    }
}
