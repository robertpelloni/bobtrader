import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

const data = [
  { name: 'Mon', volume: 4000 },
  { name: 'Tue', volume: 3000 },
  { name: 'Wed', volume: 2000 },
  { name: 'Thu', volume: 2780 },
  { name: 'Fri', volume: 1890 },
  { name: 'Sat', volume: 2390 },
  { name: 'Sun', volume: 3490 },
];

export const VolumeDashboard: React.FC = () => {
    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-4">Volume Analysis</h1>

            <div className="bg-white p-4 rounded shadow h-96">
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={data}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="name" />
                        <YAxis />
                        <Tooltip />
                        <Line type="monotone" dataKey="volume" stroke="#8884d8" activeDot={{ r: 8 }} />
                    </LineChart>
                </ResponsiveContainer>
            </div>

            <div className="mt-4 grid grid-cols-3 gap-4">
                <div className="bg-white p-4 rounded shadow">
                    <h3>Volume Ratio</h3>
                    <p className="text-2xl font-bold">1.2x</p>
                </div>
                <div className="bg-white p-4 rounded shadow">
                    <h3>Trend</h3>
                    <p className="text-2xl font-bold text-green-500">Increasing</p>
                </div>
                <div className="bg-white p-4 rounded shadow">
                    <h3>Anomaly Score</h3>
                    <p className="text-2xl font-bold">0.4</p>
                </div>
            </div>
        </div>
    );
};
