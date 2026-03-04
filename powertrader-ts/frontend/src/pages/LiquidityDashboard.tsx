import React, { useState, useEffect } from 'react';

export const LiquidityDashboard: React.FC = () => {
    const [positions, setPositions] = useState<any[]>([]);
    const [pair, setPair] = useState("WETH-USDC");
    const [amount0, setAmount0] = useState(0.1);
    const [amount1, setAmount1] = useState(300);
    const [isLoading, setIsLoading] = useState(false);

    const fetchPositions = async () => {
        try {
            const res = await fetch('http://localhost:3000/api/defi/positions');
            const data = await res.json();
            setPositions(data.positions || []);
        } catch (e) {
            console.error(e);
        }
    };

    useEffect(() => {
        fetchPositions();
    }, []);

    const addLiquidity = async () => {
        setIsLoading(true);
        try {
            const res = await fetch('http://localhost:3000/api/defi/liquidity/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ pair, amount0, amount1, rangeWidth: 2 })
            });
            const data = await res.json();
            if (data.success) {
                alert(`Liquidity Added! TX: ${data.txHash}`);
                fetchPositions();
            } else {
                alert(`Error: ${data.error}`);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    };

    const removeLiquidity = async (tokenId: number) => {
        if (!confirm("Are you sure you want to remove 100% liquidity?")) return;
        try {
            const res = await fetch('http://localhost:3000/api/defi/liquidity/remove', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ tokenId, percent: 100 })
            });
            const data = await res.json();
            if (data.success) {
                alert("Liquidity Removed");
                fetchPositions();
            }
        } catch (e) {
            console.error(e);
        }
    };

    const autoCompound = async (tokenId: number) => {
        if (!confirm("Compound unclaimed fees into position?")) return;
        setIsLoading(true);
        try {
            const res = await fetch('http://localhost:3000/api/defi/liquidity/compound', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ tokenId })
            });
            const data = await res.json();
            if (data.success) {
                alert(`Compounded Successfully! TX: ${data.txHash}`);
                fetchPositions();
            } else {
                alert("Compound failed or no fees to collect.");
            }
        } catch (e) {
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6 text-pink-600">DeFi Liquidity Manager</h1>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Control Panel */}
                <div className="bg-white p-4 sm:p-6 rounded-lg shadow order-2 lg:order-1">
                    <h2 className="text-xl font-semibold mb-4">Add Liquidity (Uniswap V3)</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Pair</label>
                            <input
                                type="text"
                                value={pair}
                                onChange={(e) => setPair(e.target.value)}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                            />
                        </div>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Amount 0 (ETH)</label>
                                <input
                                    type="number"
                                    value={amount0}
                                    onChange={(e) => setAmount0(Number(e.target.value))}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Amount 1 (USDC)</label>
                                <input
                                    type="number"
                                    value={amount1}
                                    onChange={(e) => setAmount1(Number(e.target.value))}
                                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                                />
                            </div>
                        </div>
                        <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
                            Range will be automatically calculated using Bollinger Bands (20, 2) on recent price history.
                        </div>
                        <button
                            onClick={addLiquidity}
                            disabled={isLoading}
                            className={`w-full py-2 px-4 rounded-md text-white font-bold ${isLoading ? 'bg-gray-400' : 'bg-pink-600 hover:bg-pink-700'}`}
                        >
                            {isLoading ? 'Submitting...' : 'Mint Position'}
                        </button>
                    </div>
                </div>

                {/* Positions List */}
                <div className="bg-white p-4 sm:p-6 rounded-lg shadow lg:col-span-2 order-1 lg:order-2">
                    <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-4 gap-2">
                        <h2 className="text-xl font-semibold">Your Positions</h2>
                        <button onClick={fetchPositions} className="text-sm text-blue-600 hover:underline">Refresh</button>
                    </div>

                    {positions.length === 0 ? (
                        <div className="text-center py-12 text-gray-400 border-2 border-dashed rounded">
                            No active liquidity positions found.
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {positions.map((pos) => (
                                <div key={pos.tokenId} className="border rounded-lg p-4 flex flex-col sm:flex-row justify-between items-start sm:items-center hover:bg-gray-50 gap-4">
                                    <div className="w-full sm:w-auto break-all">
                                        <div className="font-bold text-lg">Token #{pos.tokenId}</div>
                                        <div className="text-sm text-gray-600">Liquidity: <span className="font-mono text-xs">{pos.liquidity}</span></div>
                                        <div className="text-xs text-gray-400 mt-1 bg-gray-100 p-1 rounded inline-block">
                                            Range: {pos.tickLower} &lt;-&gt; {pos.tickUpper}
                                        </div>
                                    </div>
                                    <div className="text-left sm:text-right w-full sm:w-auto border-t sm:border-t-0 pt-2 sm:pt-0 mt-2 sm:mt-0">
                                        <div className="text-sm font-semibold text-green-600 mb-2">
                                            Fees: {pos.fees0} / {pos.fees1}
                                        </div>
                                        <div className="flex gap-2">
                                            <button
                                                onClick={() => autoCompound(pos.tokenId)}
                                                disabled={isLoading}
                                                className="flex-1 sm:flex-none text-purple-600 text-sm hover:text-purple-800 border border-purple-200 px-3 py-1 rounded hover:bg-purple-50 transition-colors text-center"
                                            >
                                                Compound
                                            </button>
                                            <button
                                                onClick={() => removeLiquidity(pos.tokenId)}
                                                className="flex-1 sm:flex-none text-red-600 text-sm hover:text-red-800 border border-red-200 px-3 py-1 rounded hover:bg-red-50 transition-colors text-center"
                                            >
                                                Remove
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
