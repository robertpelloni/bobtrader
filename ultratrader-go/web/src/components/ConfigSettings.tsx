import { useState, useEffect } from 'react';

export const ConfigSettings = () => {
  const [config, setConfig] = useState<any>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetch('/api/config').then(res => res.json()).then(setConfig);
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await fetch('/api/config/update', {
        method: 'POST',
        body: JSON.stringify(config),
      });
      alert('Settings updated successfully (Hot-reload enabled)');
    } catch (e) {
      console.error(e);
    }
    setSaving(false);
  };

  if (!config) return null;

  return (
    <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
      <h2 className="text-lg font-semibold mb-6">Execution & Risk Control</h2>

      <div className="space-y-6">
        <div>
          <label className="block text-xs font-bold text-[#5d7490] uppercase mb-2">Max Notional Per Trade</label>
          <input
            type="number"
            className="w-full bg-[#070d1a] border border-[#1e3050] rounded-lg px-4 py-2 text-[#d0dced] focus:border-[#18ffff] outline-none"
            value={config.risk.max_notional}
            onChange={(e) => setConfig({...config, risk: {...config.risk, max_notional: parseFloat(e.target.value)}})}
          />
        </div>

        <div>
          <label className="block text-xs font-bold text-[#5d7490] uppercase mb-2">Portfolio Risk %</label>
          <input
            type="number"
            className="w-full bg-[#070d1a] border border-[#1e3050] rounded-lg px-4 py-2 text-[#d0dced] focus:border-[#18ffff] outline-none"
            value={config.strategy.risk_pct}
            onChange={(e) => setConfig({...config, strategy: {...config.strategy, risk_pct: parseFloat(e.target.value)}})}
          />
        </div>

        <button
          onClick={handleSave}
          disabled={saving}
          className="w-full bg-[#18ffff] text-[#070d1a] font-bold py-3 rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50"
        >
          {saving ? 'SAVING...' : 'APPLY CONFIGURATION'}
        </button>
      </div>
    </div>
  );
};
