import React, { useEffect, useState } from 'react';

interface Trade {
    symbol: string;
    pnl: number;
    stage: number;
}

export const Dashboard: React.FC = () => {
    const [trades, setTrades] = useState<Trade[]>([]);
    const [account, setAccount] = useState({ total: 0, pnl: 0 });

    useEffect(() => {
        // Connect to Backend API
        const fetchData = async () => {
            try {
                // Mock fetch
                // const res = await fetch('/api/dashboard');
                // const data = await res.json();

                // Using mock data for scaffolding verification
                setAccount({ total: 12500, pnl: 450 });
                setTrades([
                    { symbol: 'BTC', pnl: 2.5, stage: 1 },
                    { symbol: 'ETH', pnl: -1.2, stage: 0 }
                ]);
            } catch (e) {
                console.error(e);
            }
        };
        fetchData();
        const interval = setInterval(fetchData, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6">PowerTrader Dashboard</h1>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Account Card */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4">Account Overview</h2>
                    <div className="space-y-2">
                        <div className="flex justify-between">
                            <span className="text-gray-600">Total Value</span>
                            <span className="font-bold text-lg">${account.total.toLocaleString()}</span>
                        </div>
                        <div className="flex justify-between">
                            <span className="text-gray-600">PnL (Realized)</span>
                            <span className={`font-bold ${account.pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                                {account.pnl >= 0 ? '+' : ''}${account.pnl.toLocaleString()}
                            </span>
                        </div>
                    </div>
                </div>

                {/* Active Trades Card */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4">Active Trades</h2>
                    <div className="overflow-x-auto">
                        <table className="min-w-full">
                            <thead>
                                <tr className="border-b">
                                    <th className="text-left py-2">Coin</th>
                                    <th className="text-right py-2">PnL %</th>
                                    <th className="text-right py-2">DCA Stage</th>
                                </tr>
                            </thead>
                            <tbody>
                                {trades.map(t => (
                                    <tr key={t.symbol} className="border-b last:border-0">
                                        <td className="py-2">{t.symbol}</td>
                                        <td className={`text-right py-2 ${t.pnl >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                                            {t.pnl.toFixed(2)}%
                                        </td>
                                        <td className="text-right py-2">{t.stage}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};
