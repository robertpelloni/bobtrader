import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { Settings } from './pages/Settings';
import { StrategySandbox } from './pages/StrategySandbox';
import { RiskDashboard } from './pages/RiskDashboard';
import { VolumeDashboard } from './pages/VolumeDashboard';
import { AILab } from './pages/AILab';
import { SystemStatus } from './pages/SystemStatus';
import { LiquidityDashboard } from './pages/LiquidityDashboard';
import { ArbitrageDashboard } from './pages/ArbitrageDashboard';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-100 flex">
        {/* Sidebar */}
        <div className="w-64 bg-gray-900 text-white p-4">
          <h1 className="text-2xl font-bold mb-8 text-blue-400">PowerTrader AI</h1>
          <nav className="space-y-2">
            <Link to="/" className="block py-2 px-4 rounded hover:bg-gray-800">Dashboard</Link>
            <Link to="/ai-lab" className="block py-2 px-4 rounded hover:bg-gray-800 text-purple-300">AI Lab</Link>
            <Link to="/defi/liquidity" className="block py-2 px-4 rounded hover:bg-gray-800 text-pink-300">Liquidity</Link>
            <Link to="/strategy-sandbox" className="block py-2 px-4 rounded hover:bg-gray-800">Backtest</Link>
            <Link to="/risk" className="block py-2 px-4 rounded hover:bg-gray-800">Risk</Link>
            <Link to="/volume" className="block py-2 px-4 rounded hover:bg-gray-800">Volume</Link>
            <Link to="/arbitrage" className="block py-2 px-4 rounded hover:bg-gray-800 text-green-300">Arbitrage</Link>
            <Link to="/status" className="block py-2 px-4 rounded hover:bg-gray-800">System Status</Link>
            <Link to="/settings" className="block py-2 px-4 rounded hover:bg-gray-800">Settings</Link>
          </nav>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/ai-lab" element={<AILab />} />
            <Route path="/defi/liquidity" element={<LiquidityDashboard />} />
            <Route path="/strategy-sandbox" element={<StrategySandbox />} />
            <Route path="/risk" element={<RiskDashboard />} />
            <Route path="/volume" element={<VolumeDashboard />} />
            <Route path="/arbitrage" element={<ArbitrageDashboard />} />
            <Route path="/status" element={<SystemStatus />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </div>
      </div>
    </Router>
  );
}

export default App;
