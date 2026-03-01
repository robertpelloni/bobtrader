import React, { useState, useEffect } from 'react';
import { useWallet } from '../context/WalletContext';

export const Dashboard: React.FC = () => {
    const [data, setData] = useState<any>(null);
    const { address, balance } = useWallet();

    useEffect(() => {
        fetchDashboard();
    }, []);

    const fetchDashboard = () => {
        fetch('http://localhost:3000/api/dashboard')
            .then(res => res.json())
            .then(setData)
            .catch(console.error);
    }

    const toggleMode = async () => {
        const newMode = data.execution_mode === 'live' ? 'paper' : 'live';
        if (newMode === 'live' && !confirm("WARNING: Switching to LIVE mode will execute real trades. Proceed?")) {
            return;
        }
        try {
            await fetch('http://localhost:3000/api/system/mode', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ mode: newMode })
            });
            fetchDashboard(); // Refresh
        } catch (e) {
            console.error(e);
        }
    };

    if (!data) return <div className="p-4">Loading...</div>;

    return (
        <div className="p-6">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold">Dashboard</h1>

                <div className="flex items-center space-x-3 bg-white px-4 py-2 rounded-full shadow-sm border border-gray-200">
                    <span className="text-sm font-medium text-gray-600 uppercase">Mode:</span>
                    <button
                        onClick={toggleMode}
                        className={`px-3 py-1 rounded-full text-xs font-bold text-white transition-colors ${data.execution_mode === 'live' ? 'bg-red-500 hover:bg-red-600' : 'bg-blue-500 hover:bg-blue-600'}`}
                    >
                        {data.execution_mode === 'live' ? '🔴 LIVE' : '🔵 PAPER'}
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                <div className="bg-white p-6 rounded-lg shadow border-l-4 border-blue-500">
                    <h3 className="text-gray-500 text-sm font-medium">Bot Balance (Exchange)</h3>
                    <p className="text-3xl font-bold">${data.account.total.toLocaleString()}</p>
                </div>
                <div className="bg-white p-6 rounded-lg shadow border-l-4 border-green-500">
                    <h3 className="text-gray-500 text-sm font-medium">Total PnL</h3>
                    <p className={`text-3xl font-bold ${data.account.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                        {data.account.pnl >= 0 ? '+' : ''}${data.account.pnl.toFixed(2)}
                    </p>
                </div>
                {/* Web3 Wallet Card */}
                <div className={`p-6 rounded-lg shadow border-l-4 ${address ? 'bg-gray-900 text-white border-purple-500' : 'bg-gray-100 border-gray-300 text-gray-400'}`}>
                    <h3 className="text-sm font-medium mb-1">Web3 Wallet (DeFi)</h3>
                    {address ? (
                        <>
                            <p className="text-2xl font-bold text-purple-400">{balance ? parseFloat(balance).toFixed(4) : '0'} ETH</p>
                            <p className="text-xs font-mono mt-2 truncate text-gray-400">{address}</p>
                        </>
                    ) : (
                        <div className="mt-2 text-sm">Not Connected. Use Sidebar to connect MetaMask.</div>
                    )}
                </div>
            </div>

            <h2 className="text-xl font-bold mb-4">Active Trades</h2>
            <div className="bg-white rounded-lg shadow overflow-hidden">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Symbol</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">PnL %</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">DCA Stage</th>
                        </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                        {data.trades.map((t: any, i: number) => (
                            <tr key={i}>
                                <td className="px-6 py-4 whitespace-nowrap font-medium">{t.symbol}</td>
                                <td className={`px-6 py-4 whitespace-nowrap font-bold ${t.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                    {t.pnl.toFixed(2)}%
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                    <span className={`px-2 py-1 rounded text-xs ${t.stage > 0 ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-800'}`}>
                                        {t.stage}/2
                                    </span>
                                </td>
                            </tr>
                        ))}
                        {data.trades.length === 0 && (
                            <tr>
                                <td colSpan={3} className="px-6 py-8 text-center text-gray-500">No active trades.</td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
