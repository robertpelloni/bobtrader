import { useState } from 'react'
import { TradingViewChart } from './components/TradingViewChart'
import { PortfolioOverview } from './components/PortfolioOverview'
import { RiskGuardStatus } from './components/RiskGuardStatus'
import { PerformanceChart } from './components/PerformanceChart'
import { ArbitrageAlerts } from './components/ArbitrageAlerts'
import { ConfigSettings } from './components/ConfigSettings'
import DepthVisualizer from './components/DepthVisualizer'

function App() {
  const [symbol] = useState('BTCUSDT')

  // Mock data for demo/verification
  const mockDepth = {
    bids: [[60000, 1.5], [59950, 2.0], [59900, 5.0]] as [number, number][],
    asks: [[60100, 1.2], [60150, 2.5], [60200, 4.0]] as [number, number][]
  };

  return (
    <div className="min-h-screen bg-[#070d1a] text-[#d0dced] p-8">
      <header className="mb-10 flex justify-between items-end border-b border-[#1e3050] pb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-[#18ffff] to-[#b388ff] bg-clip-text text-transparent">UltraTrader Go</h1>
          <p className="text-[#8ea4c2] mt-1 font-medium italic">v2.8.0 &bull; Interactive Ops & Kraken Reach</p>
        </div>
        <div className="text-right">
          <div className="text-xs uppercase tracking-widest text-[#5d7490] mb-1">Execution Engine</div>
          <div className="flex items-center gap-2 text-[#00e676] font-bold">
            <span className="w-2 h-2 bg-[#00e676] rounded-full shadow-[0_0_8px_#00e676]"></span>
            OPERATIONAL
          </div>
        </div>
      </header>

      <PortfolioOverview />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-8">
          <PerformanceChart />

          <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
            <h2 className="text-lg font-semibold mb-4 flex justify-between items-center">
              Market Intelligence: {symbol}
              <span className="text-xs font-mono text-[#5d7490]">LIVE TICK FEED</span>
            </h2>
            <TradingViewChart symbol={symbol} interval="1m" />
          </div>
        </div>

        <div className="space-y-8">
          <ConfigSettings />
          <ArbitrageAlerts />
          <RiskGuardStatus />

          <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
            <h2 className="text-lg font-semibold mb-2 text-[#ffab40]">Liquidity Depth</h2>
            <DepthVisualizer symbol={symbol} data={mockDepth} />
          </div>

          <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
            <h2 className="text-lg font-semibold mb-2 text-[#ffab40]">Activity Stream</h2>
            <div className="text-sm text-[#5d7490] py-12 text-center border-2 border-dashed border-[#1e3050] rounded-lg">
              Waiting for cross-exchange signals...
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
