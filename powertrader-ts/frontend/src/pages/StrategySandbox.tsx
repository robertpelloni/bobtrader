import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

interface KPI {
    totalTrades: number;
    winRate: number;
    profitFactor: number;
    maxDrawdown: number;
    equityCurve: { time: number, value: number }[];
    chartData: any[];
}

export const StrategySandbox: React.FC = () => {
    const [results, setResults] = useState<KPI | null>(null);
    const [strategies, setStrategies] = useState<string[]>([]);
    const [selectedStrategy, setSelectedStrategy] = useState("SMAStrategy");
    const [symbol, setSymbol] = useState("BTC");
    const [isRunning, setIsRunning] = useState(false);

    // Dynamic Strategy Params
    const [params, setParams] = useState<any>({});

    useEffect(() => {
        // Fetch strategies
        fetch('http://localhost:3000/api/strategies')
            .then(res => res.json())
            .then(data => {
                if (data.strategies) {
                    setStrategies(data.strategies.map((s: any) => s.name));
                }
            })
            .catch(console.error);
    }, []);

    const runSimulation = async () => {
        setIsRunning(true);
        try {
            const res = await fetch('http://localhost:3000/api/strategy/backtest', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    strategy: selectedStrategy,
                    symbol: symbol,
                    timeframe: '1h',
                    initialBalance: 10000
                })
            });
            const data = await res.json();
            if (data.error) throw new Error(data.error);
            setResults(data);
        } catch (e) {
            console.error(e);
            alert("Backtest failed. Check console.");
        } finally {
            setIsRunning(false);
        }
    };

    // Helper for formatting date
    const formatDate = (ts: number) => new Date(ts).toLocaleDateString();

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6">Strategy Sandbox</h1>

            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                {/* Controls */}
                <div className="bg-white p-6 rounded-lg shadow lg:col-span-1">
                    <h2 className="text-xl font-semibold mb-4">Configuration</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Strategy</label>
                            <select
                                value={selectedStrategy}
                                onChange={e => setSelectedStrategy(e.target.value)}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                            >
                                {strategies.length > 0 ? (
                                    strategies.map(s => <option key={s} value={s}>{s}</option>)
                                ) : (
                                    <>
                                        <option value="SMAStrategy">SMAStrategy</option>
                                        <option value="RSIStrategy">RSIStrategy</option>
                                    </>
                                )}
                            </select>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Symbol</label>
                            <input
                                type="text"
                                value={symbol}
                                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                            />
                        </div>

                        {/* Dynamic Params Form */}
                        {selectedStrategy === "GridStrategy" && (
                            <div className="p-3 bg-gray-50 rounded border text-sm space-y-2">
                                <h3 className="font-bold text-gray-600">Grid Settings</h3>
                                <div>
                                    <label>Lower Price</label>
                                    <input type="number" className="w-full border rounded p-1"
                                        defaultValue={20000}
                                        onChange={(e) => setParams({...params, lowerPrice: Number(e.target.value)})}
                                    />
                                </div>
                                <div>
                                    <label>Upper Price</label>
                                    <input type="number" className="w-full border rounded p-1"
                                        defaultValue={25000}
                                        onChange={(e) => setParams({...params, upperPrice: Number(e.target.value)})}
                                    />
                                </div>
                                <div>
                                    <label>Grid Lines</label>
                                    <input type="number" className="w-full border rounded p-1"
                                        defaultValue={10}
                                        onChange={(e) => setParams({...params, gridLines: Number(e.target.value)})}
                                    />
                                </div>
                            </div>
                        )}

                        {selectedStrategy === "MACDStrategy" && (
                            <div className="p-3 bg-gray-50 rounded border text-sm space-y-2">
                                <h3 className="font-bold text-gray-600">MACD Settings</h3>
                                <div className="grid grid-cols-3 gap-1">
                                    <input type="number" placeholder="Fast" className="border rounded p-1"
                                        defaultValue={12} onChange={(e) => setParams({...params, fastPeriod: Number(e.target.value)})} />
                                    <input type="number" placeholder="Slow" className="border rounded p-1"
                                        defaultValue={26} onChange={(e) => setParams({...params, slowPeriod: Number(e.target.value)})} />
                                    <input type="number" placeholder="Sig" className="border rounded p-1"
                                        defaultValue={9} onChange={(e) => setParams({...params, signalPeriod: Number(e.target.value)})} />
                                </div>
                            </div>
                        )}

                        <div className="pt-4 border-t">
                            <button
                                onClick={runSimulation}
                                disabled={isRunning}
                                className={`w-full py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${isRunning ? 'bg-gray-400' : 'bg-blue-600 hover:bg-blue-700'}`}
                            >
                                {isRunning ? 'Simulating...' : 'Run Backtest'}
                            </button>
                        </div>
                    </div>

                    {results && (
                        <div className="mt-8 space-y-4">
                            <h3 className="font-semibold text-lg">Results</h3>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="bg-gray-50 p-2 rounded">
                                    <div className="text-xs text-gray-500">Total Trades</div>
                                    <div className="font-bold">{results.totalTrades}</div>
                                </div>
                                <div className="bg-gray-50 p-2 rounded">
                                    <div className="text-xs text-gray-500">Win Rate</div>
                                    <div className={`font-bold ${results.winRate > 50 ? 'text-green-600' : 'text-red-600'}`}>
                                        {results.winRate.toFixed(1)}%
                                    </div>
                                </div>
                                <div className="bg-gray-50 p-2 rounded">
                                    <div className="text-xs text-gray-500">Profit Factor</div>
                                    <div className="font-bold">{results.profitFactor.toFixed(2)}</div>
                                </div>
                                <div className="bg-gray-50 p-2 rounded">
                                    <div className="text-xs text-gray-500">Max Drawdown</div>
                                    <div className="font-bold text-red-600">{results.maxDrawdown.toFixed(2)}%</div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Charts */}
                <div className="lg:col-span-3 space-y-6">
                    {/* Equity Curve */}
                    <div className="bg-white p-6 rounded-lg shadow h-96">
                        <h2 className="text-xl font-semibold mb-4">Equity Curve</h2>
                        {results ? (
                            <ResponsiveContainer width="100%" height="100%">
                                <LineChart data={results.equityCurve}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis
                                        dataKey="time"
                                        tickFormatter={formatDate}
                                        type="number"
                                        domain={['auto', 'auto']}
                                    />
                                    <YAxis domain={['auto', 'auto']} />
                                    <Tooltip labelFormatter={(v) => new Date(v).toLocaleString()} />
                                    <Legend />
                                    <Line type="monotone" dataKey="value" stroke="#10B981" name="Equity ($)" dot={false} strokeWidth={2} />
                                </LineChart>
                            </ResponsiveContainer>
                        ) : (
                            <div className="h-full flex items-center justify-center text-gray-400">
                                Run simulation to view equity curve
                            </div>
                        )}
                    </div>

                    {/* Price & Indicators */}
                    <div className="bg-white p-6 rounded-lg shadow h-96">
                        <h2 className="text-xl font-semibold mb-4">Market Data & Indicators</h2>
                        {results && results.chartData ? (
                            <ResponsiveContainer width="100%" height="100%">
                                <LineChart data={results.chartData}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis
                                        dataKey="timestamp"
                                        tickFormatter={formatDate}
                                        type="number"
                                        domain={['auto', 'auto']}
                                    />
                                    <YAxis yAxisId="left" domain={['auto', 'auto']} />
                                    <YAxis yAxisId="right" orientation="right" />
                                    <Tooltip labelFormatter={(v) => new Date(v).toLocaleString()} />
                                    <Legend />
                                    <Line yAxisId="left" type="monotone" dataKey="close" stroke="#6366F1" name="Price" dot={false} />
                                    {/* Conditionally render indicators based on strategy */}
                                    {selectedStrategy.includes('SMA') && (
                                        <Line yAxisId="left" type="monotone" dataKey="sma" stroke="#F59E0B" name="SMA" dot={false} />
                                    )}
                                    {selectedStrategy.includes('RSI') && (
                                        <Line yAxisId="right" type="monotone" dataKey="rsi" stroke="#EC4899" name="RSI" dot={false} />
                                    )}
                                </LineChart>
                            </ResponsiveContainer>
                        ) : (
                            <div className="h-full flex items-center justify-center text-gray-400">
                                Run simulation to view indicators
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};
