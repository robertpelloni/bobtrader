import React, { useState, useEffect } from 'react';

export const SystemStatus: React.FC = () => {
    const [status, setStatus] = useState<any>(null);

    useEffect(() => {
        fetch('http://localhost:3000/api/system/status')
            .then(res => res.json())
            .then(setStatus)
            .catch(console.error);
    }, []);

    if (!status) return <div className="p-6">Loading System Status...</div>;

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6 text-gray-800">System Status</h1>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* General Info */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4 border-b pb-2">Core System</h2>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <span className="block text-sm text-gray-500">Version</span>
                            <span className="text-lg font-bold text-blue-600">{status.version}</span>
                        </div>
                        <div>
                            <span className="block text-sm text-gray-500">Active Exchange</span>
                            <span className="text-lg font-bold uppercase">{status.exchanges.active}</span>
                        </div>
                    </div>
                </div>

                {/* Modules */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4 border-b pb-2">Modules</h2>
                    <div className="space-y-3">
                        {Object.entries(status.modules).map(([name, info]: [string, any]) => (
                            <div key={name} className="flex justify-between items-center">
                                <span className="capitalize font-medium">{name}</span>
                                <div className="flex items-center space-x-2">
                                    <span className={`w-3 h-3 rounded-full ${info.status === 'active' ? 'bg-green-500' : 'bg-red-500'}`}></span>
                                    <span className="text-sm text-gray-600">
                                        {info.version ? `v${info.version}` : (info.engine || 'Active')}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Submodules */}
                <div className="bg-white p-6 rounded-lg shadow md:col-span-2">
                    <h2 className="text-xl font-semibold mb-4 border-b pb-2">Submodules & Extensions</h2>
                    <table className="min-w-full">
                        <thead>
                            <tr className="text-left text-sm text-gray-500">
                                <th className="pb-2">Name</th>
                                <th className="pb-2">Version</th>
                                <th className="pb-2">Location</th>
                            </tr>
                        </thead>
                        <tbody>
                            {status.submodules.map((sub: any) => (
                                <tr key={sub.name} className="border-t">
                                    <td className="py-3 font-medium capitalize">{sub.name}</td>
                                    <td className="py-3 text-gray-600">v{sub.version}</td>
                                    <td className="py-3 text-sm text-gray-400 font-mono">{sub.path}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Directory Structure */}
                <div className="bg-gray-900 text-green-400 p-6 rounded-lg shadow md:col-span-2 font-mono text-sm">
                    <h2 className="text-xl font-semibold mb-4 border-b border-gray-700 pb-2 text-white">Project Structure</h2>
                    <div className="space-y-1">
                        <div>/ (Root)</div>
                        {Object.entries(status.project_structure).map(([key, path]) => (
                            <div key={key} className="pl-4">
                                ├── {key}: <span className="text-gray-400">{String(path)}</span>
                            </div>
                        ))}
                    </div>
                    <div className="mt-4 pt-4 border-t border-gray-700 text-xs text-gray-500">
                        * Virtual Submodules (Cointrade, HyperOpt) are integrated directly into the TypeScript build.
                    </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
