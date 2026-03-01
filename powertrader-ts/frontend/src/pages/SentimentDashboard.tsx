import React, { useState, useEffect } from 'react';

interface SentimentData {
    symbol: string;
    score: number;
    volume: number;
    trend: string;
}

export const SentimentDashboard: React.FC = () => {
    const [marketFNG, setMarketFNG] = useState<{score: number, classification: string} | null>(null);
    const [sentiments, setSentiments] = useState<SentimentData[]>([]);
    const [loading, setLoading] = useState(true);

    const coins = ["BTC", "ETH", "SOL", "DOGE"];

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        setLoading(true);
        try {
            // Fetch Global F&G
            const fngRes = await fetch('http://localhost:3000/api/sentiment/market/fng');
            const fngData = await fngRes.json();
            setMarketFNG(fngData);

            // Fetch individual coin sentiment
            const results = await Promise.all(
                coins.map(c => fetch(`http://localhost:3000/api/sentiment/${c}`).then(r => r.json()))
            );
            setSentiments(results);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const getScoreColor = (score: number) => {
        if (score < 30) return 'text-red-600';
        if (score < 45) return 'text-orange-500';
        if (score < 60) return 'text-yellow-500';
        if (score < 75) return 'text-green-400';
        return 'text-green-600';
    };

    const getGaugeFill = (score: number) => {
        return { width: `${score}%`, backgroundColor: score < 45 ? '#EF4444' : score > 55 ? '#10B981' : '#F59E0B' };
    };

    return (
        <div className="p-6 max-w-6xl mx-auto">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold text-indigo-600">The Socializer</h1>
                <button onClick={fetchData} className="px-4 py-2 bg-indigo-100 text-indigo-700 rounded hover:bg-indigo-200 text-sm font-bold">
                    Refresh Scan
                </button>
            </div>

            {loading ? (
                <div className="text-center py-12 text-gray-500 animate-pulse">Scanning Social Networks...</div>
            ) : (
                <>
                    {/* Global F&G */}
                    <div className="bg-white p-8 rounded-xl shadow-lg mb-8 text-center border-t-8 border-indigo-500">
                        <h2 className="text-xl font-medium text-gray-500 uppercase tracking-widest mb-4">Crypto Fear & Greed Index</h2>
                        <div className="flex flex-col items-center">
                            <div className={`text-6xl font-black mb-2 ${marketFNG ? getScoreColor(marketFNG.score) : 'text-gray-400'}`}>
                                {marketFNG?.score || '--'}
                            </div>
                            <div className="text-2xl font-light text-gray-600">
                                {marketFNG?.classification || 'Unknown'}
                            </div>
                            <div className="w-full max-w-md bg-gray-200 rounded-full h-4 mt-6 overflow-hidden">
                                <div className="h-full transition-all duration-1000" style={getGaugeFill(marketFNG?.score || 50)}></div>
                            </div>
                            <div className="flex justify-between w-full max-w-md text-xs text-gray-400 mt-2 font-bold uppercase">
                                <span>Extreme Fear (0)</span>
                                <span>Neutral (50)</span>
                                <span>Extreme Greed (100)</span>
                            </div>
                        </div>
                    </div>

                    {/* Individual Coins */}
                    <h2 className="text-2xl font-bold mb-4">Asset Sentiment (Twitter & Reddit)</h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        {sentiments.map(s => (
                            <div key={s.symbol} className="bg-white p-6 rounded-lg shadow-md border hover:border-indigo-300 transition-colors">
                                <div className="flex justify-between items-start mb-4">
                                    <h3 className="text-2xl font-bold">{s.symbol}</h3>
                                    <span className={`px-2 py-1 rounded text-xs font-bold uppercase ${s.trend === 'rising' ? 'bg-green-100 text-green-700' : s.trend === 'falling' ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-700'}`}>
                                        {s.trend}
                                    </span>
                                </div>
                                <div className="mb-4">
                                    <div className="text-sm text-gray-500">Sentiment Score</div>
                                    <div className={`text-4xl font-black ${getScoreColor(s.score)}`}>{s.score}</div>
                                </div>
                                <div>
                                    <div className="text-sm text-gray-500">Social Volume</div>
                                    <div className="text-lg font-medium">{s.volume.toLocaleString()} mentions</div>
                                </div>
                            </div>
                        ))}
                    </div>
                </>
            )}
        </div>
    );
};
