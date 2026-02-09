import React, { useEffect, useState } from 'react';

interface Strategy {
    name: string;
    interval: string;
}

export const StrategyManager: React.FC = () => {
    const [strategies, setStrategies] = useState<Strategy[]>([]);
    const [activeStrategy, setActiveStrategy] = useState("");

    useEffect(() => {
        const fetchStrategies = async () => {
            try {
                const res = await fetch('http://localhost:3000/api/strategies');
                const data = await res.json();
                setStrategies(data.strategies);
                setActiveStrategy(data.active);
            } catch (e) {
                console.error(e);
            }
        };
        fetchStrategies();
    }, []);

    const handleActivate = async (name: string) => {
        try {
            await fetch('http://localhost:3000/api/strategies/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ strategy: name })
            });
            setActiveStrategy(name);
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6">Strategy Management</h1>

            <div className="bg-white rounded-lg shadow overflow-hidden">
                <table className="min-w-full">
                    <thead className="bg-gray-50">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Interval</th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                        </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                        {strategies.map((s) => (
                            <tr key={s.name}>
                                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{s.name}</td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{s.interval}</td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                    {activeStrategy === s.name ? (
                                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                                            Active
                                        </span>
                                    ) : (
                                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">
                                            Inactive
                                        </span>
                                    )}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    {activeStrategy !== s.name && (
                                        <button
                                            onClick={() => handleActivate(s.name)}
                                            className="text-indigo-600 hover:text-indigo-900"
                                        >
                                            Activate
                                        </button>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
