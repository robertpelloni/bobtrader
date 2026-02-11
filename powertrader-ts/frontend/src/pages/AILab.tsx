import React, { useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export const AILab: React.FC = () => {
    const [training, setTraining] = useState(false);
    const [history, setHistory] = useState<any>(null);
    const [prediction, setPrediction] = useState<any>(null);
    const [symbol, setSymbol] = useState("BTC");
    const [epochs] = useState(20);

    const startTraining = async () => {
        setTraining(true);
        try {
            const res = await fetch('http://localhost:3000/api/ai/train', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ symbol, epochs })
            });
            const data = await res.json();
            if (data.history) {
                // Format history for chart (loss per epoch)
                const chartData = data.history.loss.map((loss: number, i: number) => ({
                    epoch: i + 1,
                    loss
                }));
                setHistory(chartData);
            }
        } catch (e) {
            console.error(e);
            alert("Training failed");
        } finally {
            setTraining(false);
        }
    };

    const runInference = async () => {
        try {
            const res = await fetch('http://localhost:3000/api/ai/predict', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ symbol })
            });
            const data = await res.json();
            setPrediction(data);
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6 text-purple-600">DeepThinker AI Lab</h1>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Control Panel */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4">Model Training</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Target Asset</label>
                            <input
                                type="text"
                                value={symbol}
                                onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm p-2 border"
                            />
                        </div>
                        <button
                            onClick={startTraining}
                            disabled={training}
                            className={`w-full py-2 px-4 rounded-md text-white font-bold ${training ? 'bg-gray-400' : 'bg-purple-600 hover:bg-purple-700'}`}
                        >
                            {training ? 'Training Model (This may take a while)...' : 'Start Training Loop'}
                        </button>
                    </div>

                    {prediction && (
                        <div className="mt-8 p-4 bg-gray-50 rounded border border-gray-200">
                            <h3 className="font-bold text-lg mb-2">Live Inference</h3>
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <div className="text-sm text-gray-500">Predicted Close</div>
                                    <div className="text-xl font-bold">${prediction.prediction.toFixed(2)}</div>
                                </div>
                                <div>
                                    <div className="text-sm text-gray-500">Direction</div>
                                    <div className={`text-xl font-bold ${prediction.direction === 'UP' ? 'text-green-600' : 'text-red-600'}`}>
                                        {prediction.direction}
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={runInference}
                                className="mt-4 w-full py-1 px-3 bg-blue-100 text-blue-800 rounded hover:bg-blue-200"
                            >
                                Refresh Prediction
                            </button>
                        </div>
                    )}
                    {!prediction && (
                         <button
                            onClick={runInference}
                            className="mt-8 w-full py-2 px-4 border border-gray-300 rounded-md hover:bg-gray-50"
                        >
                            Test Inference
                        </button>
                    )}
                </div>

                {/* Training Visualization */}
                <div className="bg-white p-6 rounded-lg shadow h-96">
                    <h2 className="text-xl font-semibold mb-4">Training Loss Curve</h2>
                    {history ? (
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={history}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="epoch" label={{ value: 'Epoch', position: 'insideBottom', offset: -5 }} />
                                <YAxis label={{ value: 'Loss (MSE)', angle: -90, position: 'insideLeft' }} />
                                <Tooltip />
                                <Line type="monotone" dataKey="loss" stroke="#8884d8" strokeWidth={2} dot={false} />
                            </LineChart>
                        </ResponsiveContainer>
                    ) : (
                        <div className="h-full flex items-center justify-center text-gray-400 border-2 border-dashed rounded">
                            Train the model to see performance metrics
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
