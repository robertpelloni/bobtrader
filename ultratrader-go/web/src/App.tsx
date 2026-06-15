import { useState } from 'react'
import { TradingViewChart } from './components/TradingViewChart'

function App() {
  const [symbol] = useState('BTCUSDT')

  return (
    <div className="min-h-screen bg-[#070d1a] text-[#d0dced] p-8">
      <header className="mb-8">
        <h1 className="text-2xl font-bold">UltraTrader Go Dashboard v2.5.0</h1>
        <p className="text-[#8ea4c2]">Autonomous Wealth Accumulator & Global Arbitrage</p>
      </header>

      <div className="grid grid-cols-1 gap-8">
        <div className="bg-[#0e1729] p-6 rounded-xl border border-[#1e3050]">
          <h2 className="text-lg font-semibold mb-4">Live Market Intelligence</h2>
          <TradingViewChart symbol={symbol} interval="1m" />
        </div>
      </div>
    </div>
  )
}

export default App
