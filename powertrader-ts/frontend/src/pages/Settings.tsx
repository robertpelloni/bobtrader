import React from 'react';

export const Settings: React.FC = () => {
    return (
        <div className="p-4">
            <h1 className="text-2xl font-bold">Settings</h1>
            <form className="mt-4 space-y-4">
                <div>
                    <label className="block">Coins (comma separated)</label>
                    <input type="text" className="border p-2 w-full" defaultValue="BTC, ETH, XRP" />
                </div>
                <div>
                    <label className="block">DCA Multiplier</label>
                    <input type="number" className="border p-2 w-full" defaultValue="2.0" />
                </div>
                <button className="bg-blue-500 text-white p-2 rounded">Save</button>
            </form>
        </div>
    );
};
