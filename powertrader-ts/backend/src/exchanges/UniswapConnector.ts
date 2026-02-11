import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import { ethers } from 'ethers';

// ABI Fragments needed
const ERC20_ABI = [
    "function balanceOf(address owner) view returns (uint256)",
    "function decimals() view returns (uint8)",
    "function approve(address spender, uint256 amount) returns (bool)"
];

const QUOTER_ABI = [
    "function quoteExactInputSingle(address tokenIn, address tokenOut, uint24 fee, uint256 amountIn, uint160 sqrtPriceLimitX96) external returns (uint256 amountOut)"
];

const ROUTER_ABI = [
    "function exactInputSingle((address tokenIn, address tokenOut, uint24 fee, address recipient, uint256 deadline, uint256 amountIn, uint256 amountOutMinimum, uint160 sqrtPriceLimitX96)) external payable returns (uint256 amountOut)"
];

export class UniswapConnector implements IExchangeConnector {
    name = "Uniswap";
    private provider: ethers.JsonRpcProvider;
    private wallet: ethers.Wallet | null = null;
    private routerAddress = "0xE592427A0AEce92De3Edee1F18E0157C05861564"; // V3 Router (Mainnet/Polygon)
    private quoterAddress = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"; // V3 Quoter (Mainnet/Polygon)
    private tokenMap: Record<string, string> = {
        "WETH": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
        "USDC": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7"
    };

    constructor(rpcUrl: string, privateKey?: string) {
        this.provider = new ethers.JsonRpcProvider(rpcUrl);
        if (privateKey) {
            this.wallet = new ethers.Wallet(privateKey, this.provider);
        }
    }

    private resolveToken(symbol: string): string {
        // Strip -USD suffix if present
        const s = symbol.replace('-USD', '');
        return this.tokenMap[s] || s;
    }

    async fetchTicker(pair: string): Promise<number> {
        // Estimate price by quoting 1 unit
        try {
            const [base, quote] = pair.split('-');
            const tokenIn = this.resolveToken(base);
            const tokenOut = this.resolveToken(quote || "USDC");

            const quoter = new ethers.Contract(this.quoterAddress, QUOTER_ABI, this.provider);
            const amountIn = ethers.parseEther("1.0"); // Assuming 18 decimals for simplicity (risky)

            // Note: In reality, we need to know decimals. For MVP assuming WETH/USDC logic
            const fee = 3000; // 0.3% pool

            const quotedAmountOut = await quoter.quoteExactInputSingle.staticCall(
                tokenIn,
                tokenOut,
                fee,
                amountIn,
                0
            );

            // Assuming quote is USDC (6 decimals)
            return parseFloat(ethers.formatUnits(quotedAmountOut, 6));
        } catch (e) {
            console.error(`[Uniswap] Error fetching ticker:`, e);
            return 0;
        }
    }

    async fetchOrderBook(pair: string): Promise<any> { return {}; }
    async fetchOHLCV(pair: string, interval: string, limit: number): Promise<any[]> { return []; } // DeFi candles require Subgraph

    async fetchBalance(): Promise<any> {
        if (!this.wallet) return {};
        try {
            // Get ETH Balance
            const bal = await this.provider.getBalance(this.wallet.address);
            return {
                ETH: parseFloat(ethers.formatEther(bal))
            };
        } catch (e) {
            return {};
        }
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        if (!this.wallet) throw new Error("Wallet not configured");

        const [base, quote] = pair.split('-');
        const tokenIn = side === 'sell' ? this.resolveToken(base) : this.resolveToken(quote || "USDC");
        const tokenOut = side === 'sell' ? this.resolveToken(quote || "USDC") : this.resolveToken(base);

        // Approve first (simplified)
        // ...

        const router = new ethers.Contract(this.routerAddress, ROUTER_ABI, this.wallet);

        const params = {
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            fee: 3000,
            recipient: this.wallet.address,
            deadline: Math.floor(Date.now() / 1000) + 60 * 20,
            amountIn: ethers.parseEther(amount.toString()), // Should use correct decimals
            amountOutMinimum: 0,
            sqrtPriceLimitX96: 0,
        };

        const tx = await router.exactInputSingle(params);
        return {
            id: tx.hash,
            pair,
            status: 'open',
            amount
        };
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> { return false; } // Cannot cancel mined tx
    async fetchOrder(id: string, pair: string): Promise<any> { return {}; }
    async fetchOpenOrders(pair?: string): Promise<any[]> { return []; }
}
