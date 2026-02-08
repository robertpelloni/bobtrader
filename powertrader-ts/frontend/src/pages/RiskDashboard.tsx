import React from 'react';

export const RiskDashboard: React.FC = () => {
    // Mock correlation matrix
    const matrix = [
        [1.0, 0.8, 0.2],
        [0.8, 1.0, 0.3],
        [0.2, 0.3, 1.0]
    ];
    const coins = ['BTC', 'ETH', 'SOL'];

    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-4">Risk Management</h1>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-white p-6 rounded shadow">
                    <h2 className="text-xl font-semibold mb-4">Correlation Matrix</h2>
                    <div className="grid grid-cols-4 gap-1 text-center text-sm">
                        <div className="font-bold"></div>
                        {coins.map(c => <div key={c} className="font-bold">{c}</div>)}

                        {matrix.map((row, i) => (
                            <React.Fragment key={i}>
                                <div className="font-bold">{coins[i]}</div>
                                {row.map((val, j) => (
                                    <div key={j}
                                         className={`p-2 rounded ${val > 0.7 && i!==j ? 'bg-red-200' : 'bg-green-100'}`}>
                                        {val.toFixed(2)}
                                    </div>
                                ))}
                            </React.Fragment>
                        ))}
                    </div>
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
        </div>
    );
};
