import React, { useState } from 'react';

interface Opportunity {
    symbol: string;
    buyExchange: string;
    buyPrice: number;
    sellExchange: string;
    sellPrice: number;
    spread: number;
    spreadPct: number;
    timestamp: number;
}

export const ArbitrageDashboard: React.FC = () => {
    const [opportunities, setOpportunities] = useState<Opportunity[]>([]);
    const [coins, setCoins] = useState("BTC,ETH,SOL,MATIC,DOGE,XRP");
    const [loading, setLoading] = useState(false);

    const scan = async () => {
        setLoading(true);
        try {
            const res = await fetch(`http://localhost:3000/api/arbitrage/opportunities?coins=${coins}`);
            const data = await res.json();
            setOpportunities(data.opportunities || []);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6 text-green-500">Arbitrage Scanner</h1>

            <div className="bg-white p-6 rounded-lg shadow mb-6">
                <div className="flex gap-4 items-end">
                    <div className="flex-1">
                        <label className="block text-sm font-medium text-gray-700">Coins to Scan (Comma separated)</label>
                        <input
                            type="text"
                            value={coins}
                            onChange={(e) => setCoins(e.target.value.toUpperCase())}
                            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                        />
                    </div>
                    <button
                        onClick={scan}
                        disabled={loading}
                        className={`py-2 px-6 rounded-md text-white font-bold ${loading ? 'bg-gray-400' : 'bg-green-600 hover:bg-green-700'}`}
                    >
                        {loading ? 'Scanning...' : 'Scan Now'}
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {opportunities.length === 0 && !loading && (
                    <div className="col-span-full text-center text-gray-500 py-12 border-2 border-dashed rounded">
                        No arbitrage opportunities found (Spread {'>'} 0.5%).
                    </div>
                )}

                {opportunities.map((opp, idx) => (
                    <div key={idx} className="bg-white rounded-lg shadow border-l-4 border-green-500 p-4">
                        <div className="flex justify-between items-start mb-2">
                            <h3 className="text-2xl font-bold">{opp.symbol}</h3>
                            <span className="bg-green-100 text-green-800 text-xs font-bold px-2 py-1 rounded">
                                +{opp.spreadPct.toFixed(2)}%
                            </span>
                        </div>

                        <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                            <div>
                                <div className="text-gray-500">Buy At</div>
                                <div className="font-bold">{opp.buyExchange}</div>
                                <div className="text-green-600">${opp.buyPrice.toFixed(2)}</div>
                            </div>
                            <div className="text-right">
                                <div className="text-gray-500">Sell At</div>
                                <div className="font-bold">{opp.sellExchange}</div>
                                <div className="text-red-600">${opp.sellPrice.toFixed(2)}</div>
                            </div>
                        </div>

                        <div className="text-xs text-gray-400 text-center">
                            Spread: ${opp.spread.toFixed(2)} • {new Date(opp.timestamp).toLocaleTimeString()}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
