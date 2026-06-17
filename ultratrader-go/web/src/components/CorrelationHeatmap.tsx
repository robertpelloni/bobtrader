import { useState, useEffect } from 'react';

interface CorrelationCell {
  symbol1: string;
  symbol2: string;
  correlation: number;
}

export const CorrelationHeatmap = () => {
  const [data, setData] = useState<{heatmap: CorrelationCell[], symbols: string[]} | null>(null);

  useEffect(() => {
    const fetchCorr = async () => {
      try {
        const response = await fetch('/api/analytics/correlation');
        const json = await response.json();
        setData(json);
      } catch (e) {
        console.error(e);
      }
    };
    fetchCorr();
    const interval = setInterval(fetchCorr, 15000);
    return () => clearInterval(interval);
  }, []);

  if (!data || !data.symbols || data.symbols.length < 2) return null;

  const { symbols, heatmap } = data;

  const getCorr = (s1: string, s2: string) => {
    if (s1 === s2) return 1.0;
    const cell = heatmap.find(c => (c.symbol1 === s1 && c.symbol2 === s2) || (c.symbol1 === s2 && c.symbol2 === s1));
    return cell ? cell.correlation : 0;
  };

  const getColor = (val: number) => {
    const intensity = Math.abs(val);
    if (val > 0.5) return `rgba(0, 230, 118, ${intensity})`; // Green for positive
    if (val < -0.5) return `rgba(255, 82, 82, ${intensity})`; // Red for negative
    return `rgba(30, 48, 80, ${intensity})`; // Muted for low
  };

  return (
    <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
      <h2 className="text-lg font-semibold mb-6 flex justify-between items-center">
        Portfolio Correlation Matrix
        <span className="text-[10px] text-[#5d7490] font-mono uppercase tracking-tighter">Rolling 100 Ticks</span>
      </h2>

      <div className="overflow-x-auto">
        <table className="border-collapse">
          <thead>
            <tr>
              <th className="p-2"></th>
              {symbols.map(s => (
                <th key={s} className="p-2 text-[10px] font-bold text-[#8ea4c2] transform -rotate-45 h-12 origin-bottom-left">{s}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {symbols.map(s1 => (
              <tr key={s1}>
                <td className="p-2 text-[10px] font-bold text-[#8ea4c2] text-right pr-4 border-r border-[#1e3050]">{s1}</td>
                {symbols.map(s2 => {
                  const corr = getCorr(s1, s2);
                  return (
                    <td
                      key={s2}
                      className="p-3 border border-[#1e3050] text-center text-[10px] font-mono transition-all hover:scale-110"
                      style={{ backgroundColor: getColor(corr), color: Math.abs(corr) > 0.6 ? '#000' : '#d0dced' }}
                      title={`${s1} vs ${s2}: ${corr.toFixed(4)}`}
                    >
                      {corr.toFixed(2)}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
