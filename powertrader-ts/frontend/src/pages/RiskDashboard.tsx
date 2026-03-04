import React, { useState, useEffect } from 'react';

export const RiskDashboard: React.FC = () => {
    const [matrix, setMatrix] = useState<any>({});
    const [coins, setCoins] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        fetchCorrelation();
    }, []);

    const fetchCorrelation = async () => {
        setLoading(true);
        try {
            const res = await fetch('http://localhost:3000/api/risk/correlation?coins=BTC,ETH,SOL,MATIC,BNB');
            const data = await res.json();
            if (data.matrix) {
                setCoins(data.coins);
                setMatrix(data.matrix);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-4 text-red-600">Risk Management</h1>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-white p-6 rounded shadow">
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="text-xl font-semibold">Correlation Matrix</h2>
                        <button onClick={fetchCorrelation} disabled={loading} className="text-sm text-blue-600">
                            {loading ? 'Calculating...' : 'Refresh'}
                        </button>
                    </div>

                    {coins.length > 0 ? (
                        <div className="overflow-x-auto">
                            <table className="min-w-full text-center text-sm">
                                <thead>
                                    <tr>
                                        <th></th>
                                        {coins.map(c => <th key={c} className="p-2">{c}</th>)}
                                    </tr>
                                </thead>
                                <tbody>
                                    {coins.map(coinA => (
                                        <tr key={coinA}>
                                            <td className="font-bold p-2">{coinA}</td>
                                            {coins.map(coinB => {
                                                const val = matrix[coinA]?.[coinB] || 0;
                                                const bgClass = val > 0.8 && coinA !== coinB ? 'bg-red-200'
                                                              : val > 0.5 && coinA !== coinB ? 'bg-yellow-100'
                                                              : val < 0 && coinA !== coinB ? 'bg-green-200' // Negative correlation is good hedge
                                                              : 'bg-gray-50';
                                                return (
                                                    <td key={coinB} className={`p-2 rounded border border-white ${bgClass}`}>
                                                        {val.toFixed(2)}
                                                    </td>
                                                );
                                            })}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <div className="text-gray-400 text-center py-8">No data available</div>
                    )}
                </div>

                <div className="bg-white p-6 rounded shadow">
                    <h2 className="text-xl font-semibold mb-4">Position Sizing</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm text-gray-600">Account Risk %</label>
                            <input type="number" className="border rounded p-2 w-full" defaultValue="2.0" />
                        </div>
                        <div className="p-4 bg-gray-50 rounded">
                            <p className="text-sm">Recommended Position Size (BTC)</p>
                            <p className="text-2xl font-bold">$1,250.00</p>
                            <p className="text-xs text-gray-500">Based on ATR volatility</p>
                        </div>
                    </div>
                </div>
            </div>

            <WhaleWatcherWidget />
        </div>
    );
};

const WhaleWatcherWidget: React.FC = () => {
    const [events, setEvents] = useState<any[]>([]);

    useEffect(() => {
        fetch('http://localhost:3000/api/risk/whales')
            .then(res => res.json())
            .then(data => setEvents(data.events || []))
            .catch(console.error);
    }, []);

    return (
        <div className="mt-8 bg-white p-6 rounded shadow">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold">On-Chain Whale Watcher</h2>
                <span className="text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded-full font-bold uppercase">Stablecoins</span>
            </div>
            <div className="overflow-x-auto">
                <table className="min-w-full text-sm text-left">
                    <thead className="bg-gray-50 text-gray-500">
                        <tr>
                            <th className="p-3 font-medium">Time</th>
                            <th className="p-3 font-medium">Type</th>
                            <th className="p-3 font-medium">Amount</th>
                            <th className="p-3 font-medium">From / To</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                        {events.map(ev => (
                            <tr key={ev.id} className="hover:bg-gray-50 transition-colors">
                                <td className="p-3 whitespace-nowrap text-gray-500">
                                    {new Date(ev.timestamp).toLocaleTimeString()}
                                </td>
                                <td className="p-3">
                                    <span className={`px-2 py-1 rounded text-xs font-bold uppercase ${ev.type === 'deposit' ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'}`}>
                                        {ev.type === 'deposit' ? 'Exchange Deposit' : 'Exchange Withdrawal'}
                                    </span>
                                </td>
                                <td className="p-3 font-bold whitespace-nowrap">
                                    ${(ev.amountUsd / 1000000).toFixed(2)}M {ev.token}
                                </td>
                                <td className="p-3 font-mono text-xs text-gray-400">
                                    {ev.from.substring(0,6)}... -> {ev.to.substring(0,6)}...
                                </td>
                            </tr>
                        ))}
                        {events.length === 0 && (
                            <tr><td colSpan={4} className="p-6 text-center text-gray-400">No recent whale activity.</td></tr>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
