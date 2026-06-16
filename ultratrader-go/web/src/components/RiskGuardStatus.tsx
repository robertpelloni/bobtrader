import { useState, useEffect } from 'react';

interface GuardDiagnostics {
  active_guards: string[];
  metrics: {
    execution_blocked: number;
  };
}

export const RiskGuardStatus = () => {
  const [diags, setDiags] = useState<GuardDiagnostics | null>(null);

  useEffect(() => {
    const fetchGuards = async () => {
      try {
        const response = await fetch('/api/guard-diagnostics');
        const data = await response.json();
        setDiags(data);
      } catch (error) {
        console.error('Failed to fetch guard diagnostics:', error);
      }
    };

    fetchGuards();
    const interval = setInterval(fetchGuards, 5000);
    return () => clearInterval(interval);
  }, []);

  if (!diags) return <div className="p-4 text-center text-[#5d7490]">Initializing shield...</div>;

  return (
    <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
      <h2 className="text-lg font-semibold mb-4 flex items-center justify-between">
        <span className="flex items-center gap-2">
          <span className="w-2 h-2 bg-[#00e676] rounded-full animate-pulse shadow-[0_0_8px_#00e676]"></span>
          Active Risk Shield
        </span>
        <span className="text-[10px] bg-red-500/10 text-red-500 px-2 py-0.5 rounded border border-red-500/20 uppercase font-bold">
          {diags.metrics?.execution_blocked || 0} Blocked
        </span>
      </h2>
      <div className="space-y-3">
        {diags.active_guards.map(name => (
          <div key={name} className="flex justify-between items-center p-3 bg-[#121e34] rounded-lg border border-[#1e3050] hover:bg-[#1e3050]/50 transition-colors">
            <span className="text-sm font-medium">{name}</span>
            <span className="text-[10px] px-2 py-1 bg-black/20 rounded border border-[#1e3050] text-[#00e676] font-mono">
              PASS
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};
