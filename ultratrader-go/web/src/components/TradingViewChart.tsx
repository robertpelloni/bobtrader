import React, { useEffect, useRef } from 'react';
import { createChart, ColorType, type CandlestickData, type Time } from 'lightweight-charts';

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
        const response = await fetch(`/api/marketdata/candles?symbol=${symbol}&interval=${interval}&limit=500`);
        const data = await response.json();

        const formattedData: CandlestickData[] = data.map((c: any) => ({
          time: (new Date(c.timestamp).getTime() / 1000) as Time,
          open: parseFloat(c.open),
          high: parseFloat(c.high),
          low: parseFloat(c.low),
          close: parseFloat(c.close),
        }));

        candlestickSeries.setData(formattedData);
      } catch (error) {
        console.error('Failed to fetch candle data:', error);
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
