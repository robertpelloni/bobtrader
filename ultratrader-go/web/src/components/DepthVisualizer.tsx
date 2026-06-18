import React, { useEffect, useRef } from 'react';

interface DepthData {
  bids: [number, number][]; // [price, quantity]
  asks: [number, number][];
}

interface DepthVisualizerProps {
  symbol: string;
  data: DepthData;
}

const DepthVisualizer: React.FC<DepthVisualizerProps> = ({ symbol, data }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const width = canvas.width;
    const height = canvas.height;

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    if (data.bids.length === 0 && data.asks.length === 0) {
      ctx.fillStyle = '#666';
      ctx.textAlign = 'center';
      ctx.fillText('No depth data available', width / 2, height / 2);
      return;
    }

    // Combine and find price range
    const allPrices = [
      ...data.bids.map((b) => b[0]),
      ...data.asks.map((a) => a[0]),
    ];
    const minPrice = Math.min(...allPrices);
    const maxPrice = Math.max(...allPrices);
    const priceRange = (maxPrice - minPrice) || 1;

    // Calculate cumulative volume
    const bidsWithCumulative = [] as { price: number; quantity: number; cumulative: number }[];
    let bidCumulative = 0;
    [...data.bids].sort((a, b) => b[0] - a[0]).forEach((b) => {
      bidCumulative += b[1];
      bidsWithCumulative.push({ price: b[0], quantity: b[1], cumulative: bidCumulative });
    });

    const asksWithCumulative = [] as { price: number; quantity: number; cumulative: number }[];
    let askCumulative = 0;
    [...data.asks].sort((a, b) => a[0] - b[0]).forEach((a) => {
      askCumulative += a[1];
      asksWithCumulative.push({ price: a[0], quantity: a[1], cumulative: askCumulative });
    });

    const maxCumulative = Math.max(bidCumulative, askCumulative) || 1;

    const getX = (price: number) => ((price - minPrice) / priceRange) * width;
    const getY = (cumulative: number) => height - (cumulative / maxCumulative) * height;

    // Draw Bids (Buy side - Green)
    ctx.beginPath();
    ctx.fillStyle = 'rgba(0, 255, 0, 0.2)';
    ctx.strokeStyle = 'green';

    if (bidsWithCumulative.length > 0) {
        ctx.moveTo(getX(bidsWithCumulative[0].price), height);
        bidsWithCumulative.forEach(b => {
            ctx.lineTo(getX(b.price), getY(b.cumulative));
        });
        ctx.lineTo(getX(minPrice), getY(bidCumulative));
        ctx.lineTo(getX(minPrice), height);
    }
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    // Draw Asks (Sell side - Red)
    ctx.beginPath();
    ctx.fillStyle = 'rgba(255, 0, 0, 0.2)';
    ctx.strokeStyle = 'red';

    if (asksWithCumulative.length > 0) {
        ctx.moveTo(getX(asksWithCumulative[0].price), height);
        asksWithCumulative.forEach(a => {
            ctx.lineTo(getX(a.price), getY(a.cumulative));
        });
        ctx.lineTo(getX(maxPrice), getY(askCumulative));
        ctx.lineTo(getX(maxPrice), height);
    }
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    // Draw Labels
    ctx.fillStyle = 'white';
    ctx.font = '12px Arial';
    ctx.fillText(`${symbol} Depth`, 10, 20);

  }, [data, symbol]);

  return (
    <div className="bg-gray-800/50 p-4 rounded-lg">
      <canvas
        ref={canvasRef}
        width={600}
        height={300}
        className="w-full h-auto rounded"
      />
      <div className="flex justify-between mt-2 text-[10px] uppercase tracking-widest text-[#5d7490]">
        <span>Bids</span>
        <span>Asks</span>
      </div>
    </div>
  );
};

export default DepthVisualizer;
