import { useState, useEffect } from 'react';

interface PortfolioSummary {
  open_positions: number;
  total_market_value: number;
  total_realized_pnl: number;
  total_unrealized_pnl: number;
  total_siphoned: number;
}

export const PortfolioOverview = () => {
  const [summary, setSummary] = useState<PortfolioSummary | null>(null);

  useEffect(() => {
    const fetchSummary = async () => {
      try {
        const response = await fetch('/api/portfolio-summary');
        const data = await response.json();
        setSummary(data);
      } catch (error) {
        console.error('Failed to fetch portfolio summary:', error);
      }
    };

    fetchSummary();
    const interval = setInterval(fetchSummary, 5000);
    return () => clearInterval(interval);
  }, []);

  if (!summary) return <div className="p-10 text-center text-[#5d7490]">Loading portfolio metrics...</div>;

  const fmt = (n: number) => n.toLocaleString('en-US', { style: 'currency', currency: 'USD' });

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050] hover:border-[#18ffff]/30 transition-colors">
        <h3 className="text-[#8ea4c2] text-xs uppercase tracking-wider mb-2 font-bold">Market Value</h3>
        <p className="text-2xl font-black text-[#18ffff]">{fmt(summary.total_market_value)}</p>
      </div>
      <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050] hover:border-[#00e676]/30 transition-colors">
        <h3 className="text-[#8ea4c2] text-xs uppercase tracking-wider mb-2 font-bold">Realized PnL</h3>
        <p className={`text-2xl font-black ${summary.total_realized_pnl >= 0 ? 'text-[#00e676]' : 'text-[#ff5252]'}`}>
          {fmt(summary.total_realized_pnl)}
        </p>
      </div>
      <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050] hover:border-[#b388ff]/30 transition-colors">
        <h3 className="text-[#8ea4c2] text-xs uppercase tracking-wider mb-2 font-bold">Siphoned Wealth</h3>
        <p className="text-2xl font-black text-[#b388ff]">{fmt(summary.total_siphoned)}</p>
      </div>
      <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
        <h3 className="text-[#8ea4c2] text-xs uppercase tracking-wider mb-2 font-bold">Active Exposure</h3>
        <p className="text-2xl font-black">{summary.open_positions} Positions</p>
      </div>
    </div>
  );
};
