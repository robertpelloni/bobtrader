import React, { useState, useEffect } from 'react';

export const Settings: React.FC = () => {
    const [config, setConfig] = useState<any>({});
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await fetch('http://localhost:3000/api/settings');
            const data = await res.json();
            setConfig(data);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (section: string, key: string, value: any) => {
        setConfig((prev: any) => ({
            ...prev,
            [section]: {
                ...prev[section],
                [key]: value
            }
        }));
    };

    const handleRootChange = (key: string, value: any) => {
        setConfig((prev: any) => ({
            ...prev,
            [key]: value
        }));
    };

    const save = async () => {
        try {
            await fetch('http://localhost:3000/api/settings', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(config)
            });
            alert('Settings Saved!');
        } catch (e) {
            alert('Failed to save settings');
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div className="p-6">
            <h1 className="text-3xl font-bold mb-6">System Configuration</h1>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Trading Config */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4 border-b pb-2">Trading Engine</h2>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Active Coins</label>
                            <input
                                type="text"
                                className="mt-1 border p-2 w-full rounded"
                                value={config.coins || ""}
                                onChange={(e) => handleRootChange('coins', e.target.value)}
                            />
                            <p className="text-xs text-gray-500">Comma separated symbols (e.g. BTC,ETH)</p>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">DCA Multiplier</label>
                                <input
                                    type="number"
                                    className="mt-1 border p-2 w-full rounded"
                                    value={config.dca_multiplier || 2.0}
                                    onChange={(e) => handleRootChange('dca_multiplier', parseFloat(e.target.value))}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Max DCA Buys</label>
                                <input
                                    type="number"
                                    className="mt-1 border p-2 w-full rounded"
                                    value={config.max_dca_buys_per_24h || 2}
                                    onChange={(e) => handleRootChange('max_dca_buys_per_24h', parseInt(e.target.value))}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                {/* Notifications Config */}
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4 border-b pb-2">Notifications</h2>

                    {/* Discord */}
                    <div className="mb-6">
                        <h3 className="font-medium text-purple-600 mb-2">Discord</h3>
                        <div className="flex items-center mb-2">
                            <input
                                type="checkbox"
                                className="mr-2"
                                checked={config.notifications?.discord_enabled || false}
                                onChange={(e) => handleChange('notifications', 'discord_enabled', e.target.checked)}
                            />
                            <span className="text-sm">Enable Discord Alerts</span>
                        </div>
                        <input
                            type="text"
                            placeholder="Webhook URL"
                            className="border p-2 w-full rounded text-sm"
                            value={config.notifications?.discord_webhook_url || ""}
                            onChange={(e) => handleChange('notifications', 'discord_webhook_url', e.target.value)}
                        />
                    </div>

                    {/* Telegram */}
                    <div className="mb-6">
                        <h3 className="font-medium text-blue-500 mb-2">Telegram</h3>
                        <div className="flex items-center mb-2">
                            <input
                                type="checkbox"
                                className="mr-2"
                                checked={config.notifications?.telegram_enabled || false}
                                onChange={(e) => handleChange('notifications', 'telegram_enabled', e.target.checked)}
                            />
                            <span className="text-sm">Enable Telegram Alerts</span>
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                            <input
                                type="text"
                                placeholder="Bot Token"
                                className="border p-2 w-full rounded text-sm"
                                value={config.notifications?.telegram_bot_token || ""}
                                onChange={(e) => handleChange('notifications', 'telegram_bot_token', e.target.value)}
                            />
                            <input
                                type="text"
                                placeholder="Chat ID"
                                className="border p-2 w-full rounded text-sm"
                                value={config.notifications?.telegram_chat_id || ""}
                                onChange={(e) => handleChange('notifications', 'telegram_chat_id', e.target.value)}
                            />
                        </div>
                    </div>
                </div>
            </div>

            <button
                onClick={save}
                className="mt-6 bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-8 rounded shadow-lg"
            >
                Save Configuration
            </button>
        </div>
    );
};
