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

const POSITION_MANAGER_ABI = [
    "function mint((address token0, address token1, uint24 fee, int24 tickLower, int24 tickUpper, uint256 amount0Desired, uint256 amount1Desired, uint256 amount0Min, uint256 amount1Min, address recipient, uint256 deadline)) external payable returns (uint256 tokenId, uint128 liquidity, uint256 amount0, uint256 amount1)",
    "function increaseLiquidity((uint256 tokenId, uint256 amount0Desired, uint256 amount1Desired, uint256 amount0Min, uint256 amount1Min, uint256 deadline)) external payable returns (uint128 liquidity, uint256 amount0, uint256 amount1)",
    "function decreaseLiquidity((uint256 tokenId, uint128 liquidity, uint256 amount0Min, uint256 amount1Min, uint256 deadline)) external payable returns (uint256 amount0, uint256 amount1)",
    "function collect((uint256 tokenId, address recipient, uint128 amount0Max, uint128 amount1Max)) external payable returns (uint256 amount0, uint256 amount1)",
    "function positions(uint256 tokenId) external view returns (uint96 nonce, address operator, address token0, address token1, uint24 fee, int24 tickLower, int24 tickUpper, uint128 liquidity, uint256 feeGrowthInside0LastX128, uint256 feeGrowthInside1LastX128, uint128 tokensOwed0, uint128 tokensOwed1)",
    "function balanceOf(address owner) view returns (uint256)",
    "function tokenOfOwnerByIndex(address owner, uint256 index) view returns (uint256)"
];

export class UniswapConnector implements IExchangeConnector {
    name = "Uniswap";
    private provider: ethers.JsonRpcProvider;
    private wallet: ethers.Wallet | null = null;
    private routerAddress = "0xE592427A0AEce92De3Edee1F18E0157C05861564"; // V3 Router (Mainnet/Polygon)
    private quoterAddress = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"; // V3 Quoter (Mainnet/Polygon)
    private positionManagerAddress = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"; // NonfungiblePositionManager (Mainnet/Polygon)
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

    // --- Liquidity Provisioning ---

    async addLiquidity(pair: string, amount0: number, amount1: number, tickLower: number, tickUpper: number): Promise<any> {
        if (!this.wallet) throw new Error("Wallet not configured");

        const [base, quote] = pair.split('-');
        const token0 = this.resolveToken(base);
        const token1 = this.resolveToken(quote || "USDC");

        // Ensure token0 < token1 address sort order (Uniswap requirement)
        // If not, swap amounts and tokens. (Simplified for this snippet, assumes caller handles or simple order)

        const pm = new ethers.Contract(this.positionManagerAddress, POSITION_MANAGER_ABI, this.wallet);

        // Approve tokens first... (omitted for brevity, assume approved)

        const params = {
            token0: token0,
            token1: token1,
            fee: 3000,
            tickLower: tickLower,
            tickUpper: tickUpper,
            amount0Desired: ethers.parseEther(amount0.toString()), // Assume 18 decimals
            amount1Desired: ethers.parseUnits(amount1.toString(), 6), // Assume USDC 6 decimals
            amount0Min: 0,
            amount1Min: 0,
            recipient: this.wallet.address,
            deadline: Math.floor(Date.now() / 1000) + 60 * 20
        };

        const tx = await pm.mint(params);
        return tx.hash;
    }

    async removeLiquidity(tokenId: number, percent: number = 100): Promise<any> {
        if (!this.wallet) throw new Error("Wallet not configured");
        const pm = new ethers.Contract(this.positionManagerAddress, POSITION_MANAGER_ABI, this.wallet);

        const pos = await pm.positions(tokenId);
        const liquidity = pos.liquidity;
        const liquidityToRemove = (liquidity * BigInt(percent)) / BigInt(100);

        const params = {
            tokenId: tokenId,
            liquidity: liquidityToRemove,
            amount0Min: 0,
            amount1Min: 0,
            deadline: Math.floor(Date.now() / 1000) + 60 * 20
        };

        const tx = await pm.decreaseLiquidity(params);
        return tx.hash;
    }

    async increaseLiquidity(tokenId: number, amount0: number, amount1: number): Promise<any> {
        if (!this.wallet) throw new Error("Wallet not configured");
        const pm = new ethers.Contract(this.positionManagerAddress, POSITION_MANAGER_ABI, this.wallet);

        // Approve tokens first (Assuming already approved for MVP simplicity, or re-approve here)
        // In prod, check allowance.

        const params = {
            tokenId: tokenId,
            amount0Desired: ethers.parseEther(amount0.toString()), // Assume 18 decimals (WETH)
            amount1Desired: ethers.parseUnits(amount1.toString(), 6), // Assume 6 decimals (USDC)
            amount0Min: 0,
            amount1Min: 0,
            deadline: Math.floor(Date.now() / 1000) + 60 * 20
        };

        const tx = await pm.increaseLiquidity(params);
        return tx.hash;
    }

    async collectFees(tokenId: number): Promise<any> {
        if (!this.wallet) throw new Error("Wallet not configured");
        const pm = new ethers.Contract(this.positionManagerAddress, POSITION_MANAGER_ABI, this.wallet);

        const params = {
            tokenId: tokenId,
            recipient: this.wallet.address,
            amount0Max: BigInt("340282366920938463463374607431768211455"), // MaxUint128
            amount1Max: BigInt("340282366920938463463374607431768211455")  // MaxUint128
        };

        const tx = await pm.collect(params);
        return tx.hash;
    }

    async fetchPositions(): Promise<any[]> {
        if (!this.wallet) return [];
        const pm = new ethers.Contract(this.positionManagerAddress, POSITION_MANAGER_ABI, this.wallet); // Using wallet for provider

        const balance = await pm.balanceOf(this.wallet.address);
        const positions: any[] = [];

        for (let i = 0; i < Number(balance); i++) {
            const tokenId = await pm.tokenOfOwnerByIndex(this.wallet.address, i);
            const pos = await pm.positions(tokenId);
            positions.push({
                tokenId: tokenId.toString(),
                liquidity: pos.liquidity.toString(),
                token0: pos.token0,
                token1: pos.token1,
                tickLower: pos.tickLower,
                tickUpper: pos.tickUpper,
                fees0: pos.tokensOwed0.toString(),
                fees1: pos.tokensOwed1.toString()
            });
        }
        return positions;
    }
}
