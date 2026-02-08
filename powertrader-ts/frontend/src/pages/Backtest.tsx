import React from 'react';

export const Backtest: React.FC = () => {
    return (
        <div className="p-4">
            <h1 className="text-2xl font-bold">Backtest & Strategy</h1>
            <div className="mt-4">
                <button className="bg-green-500 text-white p-2 rounded">Run Backtest</button>
            </div>
            <div className="mt-4 h-64 bg-gray-100 border rounded flex items-center justify-center">
                <p>Backtest Results Chart Placeholder</p>
            </div>
        </div>
    );
};
