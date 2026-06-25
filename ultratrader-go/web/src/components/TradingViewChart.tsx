import React, { useEffect, useRef } from 'react';
import { createChart, ColorType, type CandlestickData, type Time, type SeriesMarker } from 'lightweight-charts';

interface ChartProps {
  symbol: string;
  interval: string;
}

export const TradingViewChart: React.FC<ChartProps> = ({ symbol, interval }) => {
  const chartContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#0e1729' },
        textColor: '#d0dced',
      },
      grid: {
        vertLines: { color: '#1e3050' },
        horzLines: { color: '#1e3050' },
      },
      width: chartContainerRef.current.clientWidth,
      height: 400,
    });

    const candlestickSeries = (chart as any).addCandlestickSeries({
      upColor: '#00e676',
      downColor: '#ff5252',
      borderVisible: false,
      wickUpColor: '#00e676',
      wickDownColor: '#ff5252',
    });

    const handleResize = () => {
      chart.applyOptions({ width: chartContainerRef.current?.clientWidth });
    };

    window.addEventListener('resize', handleResize);

    const fetchData = async () => {
      try {
        const [candleRes, signalRes, orderRes] = await Promise.all([
          fetch(`/api/marketdata/candles?symbol=${symbol}&interval=${interval}&limit=500`),
          fetch(`/api/signals`),
          fetch(`/api/orders`)
        ]);

        const candles = await candleRes.json();
        const signals = await signalRes.json();
        const orders = await orderRes.json();

        const formattedData: CandlestickData[] = candles.map((c: any) => ({
          time: (new Date(c.timestamp).getTime() / 1000) as Time,
          open: parseFloat(c.open),
          high: parseFloat(c.high),
          low: parseFloat(c.low),
          close: parseFloat(c.close),
        }));

        candlestickSeries.setData(formattedData);

        // Add Markers for Signals & Orders
        const markers: SeriesMarker<Time>[] = [];

        // Signals (v2.8.0 improvement)
        signals.filter((s: any) => s.symbol === symbol).forEach((s: any) => {
          markers.push({
            time: (new Date(s.timestamp).getTime() / 1000) as Time,
            position: s.action === 'buy' ? 'belowBar' : 'aboveBar',
            color: s.action === 'buy' ? '#18ffff' : '#ffab40',
            shape: s.action === 'buy' ? 'arrowUp' : 'arrowDown',
            text: s.action.toUpperCase(),
          });
        });

        // Orders (Confirmed trades)
        orders.filter((o: any) => o.Symbol === symbol && o.Status === 'filled').forEach((o: any) => {
           markers.push({
            time: (new Date(o.Timestamp).getTime() / 1000) as Time,
            position: o.Side === 'buy' ? 'belowBar' : 'aboveBar',
            color: o.Side === 'buy' ? '#00e676' : '#ff5252',
            shape: 'circle',
            text: 'FILL',
          });
        });

        markers.sort((a,b) => (a.time as number) - (b.time as number));
        candlestickSeries.setMarkers(markers);

      } catch (error) {
        console.error('Failed to fetch data:', error);
      }
    };

    fetchData();
    const intervalId = setInterval(fetchData, 10000);

    return () => {
      window.removeEventListener('resize', handleResize);
      clearInterval(intervalId);
      chart.remove();
    };
  }, [symbol, interval]);

  return (
    <div className="w-full h-full bg-[#0e1729] rounded-lg overflow-hidden border border-[#1e3050]">
      <div className="p-4 border-bottom border-[#1e3050] flex justify-between items-center">
        <h3 className="text-sm font-semibold">{symbol} - {interval}</h3>
      </div>
      <div ref={chartContainerRef} className="w-full h-[400px]" />
    </div>
  );
};
