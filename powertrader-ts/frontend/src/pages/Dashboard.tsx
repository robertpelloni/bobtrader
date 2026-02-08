import React, { useEffect, useState } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

interface Trade {
    symbol: string;
    pnl: number;
    stage: number;
}

export const Dashboard: React.FC = () => {
    const [trades, setTrades] = useState<Trade[]>([]);
    const [account, setAccount] = useState({ total: 0, pnl: 0 });
    const { isConnected, lastMessage } = useWebSocket('ws://localhost:3000');

    // Handle real-time updates
    useEffect(() => {
        if (lastMessage) {
            if (lastMessage.type === 'TRADE_UPDATE') {
                // Update specific trade in list
                const update = lastMessage.payload;
                setTrades(prev => {
                    const idx = prev.findIndex(t => t.symbol === update.symbol);
                    if (idx >= 0) {
                        const next = [...prev];
                        next[idx] = update;
                        return next;
                    } else {
                        return [...prev, update];
                    }
                });
            } else if (lastMessage.type === 'ACCOUNT_UPDATE') {
                setAccount(lastMessage.payload);
            }
        }
    }, [lastMessage]);

    // Initial fetch (fallback/hydration)
    useEffect(() => {
        const fetchData = async () => {
            try {
                const res = await fetch('http://localhost:3000/api/dashboard');
                const data = await res.json();
                setAccount(data.account);
                setTrades(data.trades);
            } catch (e) {
                console.error(e);
            }
        };
        fetchData();
    }, []);

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold">PowerTrader Dashboard</h1>
                <span className={`px-3 py-1 rounded text-sm ${isConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                    {isConnected ? 'Live Connected' : 'Disconnected'}
                </span>
            </div>

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
