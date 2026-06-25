import { useEffect, useRef } from 'react';
import { createChart, ColorType, type LineData, type Time } from 'lightweight-charts';

export const PerformanceChart = () => {
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
      height: 300,
    });

    const lineSeries = (chart as any).addLineSeries({
      color: '#00e676',
      lineWidth: 2,
    });

    const handleResize = () => {
      chart.applyOptions({ width: chartContainerRef.current?.clientWidth });
    };

    window.addEventListener('resize', handleResize);

    const fetchData = async () => {
      try {
        const response = await fetch('/api/runtime-reports/history?type=portfolio-valuation&limit=100');
        const data = await response.json();

        const formattedData: LineData[] = data.map((r: any) => ({
          time: (new Date(r.timestamp).getTime() / 1000) as Time,
          value: parseFloat(r.payload.realized_pnl || 0),
        })).sort((a: any, b: any) => (a.time as number) - (b.time as number));

        lineSeries.setData(formattedData);
        chart.timeScale().fitContent();
      } catch (error) {
        console.error('Failed to fetch performance data:', error);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 10000);

    return () => {
      window.removeEventListener('resize', handleResize);
      clearInterval(interval);
      chart.remove();
    };
  }, []);

  return (
    <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
      <h2 className="text-lg font-semibold mb-4">Realized PnL Trend</h2>
      <div ref={chartContainerRef} className="w-full h-[300px]" />
    </div>
  );
};
