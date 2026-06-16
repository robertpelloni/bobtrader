import { useState, useEffect } from 'react';

interface ArbitrageOpp {
  symbol: string;
  buy_exchange: string;
  sell_exchange: string;
  buy_price: number;
  sell_price: number;
  spread: number;
}

export const ArbitrageAlerts = () => {
  const [opps, setOpps] = useState<ArbitrageOpp[]>([]);

  useEffect(() => {
    const fetchArb = async () => {
      try {
        const response = await fetch('/api/marketdata/global-bbo?symbol=BTCUSDT');
        const data = await response.json();

        // Manual calculation for demo since /api/marketdata/global-bbo returns quotes
        const quotes = data.quotes || [];
        if (quotes.length < 2) return;

        const newOpps: ArbitrageOpp[] = [];
        for (let i = 0; i < quotes.length; i++) {
          for (let j = i + 1; j < quotes.length; j++) {
            const q1 = quotes[i];
            const q2 = quotes[j];
            const low = q1.Price < q2.Price ? q1 : q2;
            const high = q1.Price < q2.Price ? q2 : q1;
            const spread = (high.Price - low.Price) / low.Price;

            if (spread > 0.0001) { // 0.01% threshold
              newOpps.push({
                symbol: data.symbol,
                buy_exchange: low.Exchange,
                sell_exchange: high.Exchange,
                buy_price: low.Price,
                sell_price: high.Price,
                spread: spread
              });
            }
          }
        }
        setOpps(newOpps.sort((a,b) => b.spread - a.spread));
      } catch (error) {
        console.error('Failed to fetch arbitrage data:', error);
      }
    };

    fetchArb();
    const interval = setInterval(fetchArb, 5000);
    return () => clearInterval(interval);
  }, []);

  if (opps.length === 0) return null;

  return (
    <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
      <h2 className="text-lg font-semibold mb-4 text-[#18ffff] flex items-center gap-2">
        <span className="w-2 h-2 bg-[#18ffff] rounded-full animate-ping"></span>
        Arbitrage Scanner
      </h2>
      <div className="space-y-4">
        {opps.map((o, i) => (
          <div key={i} className="p-4 bg-[#121e34] rounded-lg border-l-4 border-[#18ffff]">
            <div className="flex justify-between items-start mb-2">
              <span className="font-bold">{o.symbol}</span>
              <span className="text-[#00e676] font-mono text-sm">{(o.spread * 100).toFixed(4)}%</span>
            </div>
            <div className="text-xs text-[#8ea4c2] grid grid-cols-2 gap-2">
              <div>BUY: <span className="text-[#d0dced]">{o.buy_exchange}</span> @ ${o.buy_price.toFixed(2)}</div>
              <div>SELL: <span className="text-[#d0dced]">{o.sell_exchange}</span> @ ${o.sell_price.toFixed(2)}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
