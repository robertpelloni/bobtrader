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
import { WalletProvider, useWallet } from './context/WalletContext';

const WalletButton = () => {
    const { address, connect, disconnect, isConnecting, balance } = useWallet();

    if (address) {
        return (
            <div className="bg-gray-800 p-3 rounded-lg mb-6">
                <div className="text-xs text-gray-400 mb-1">Connected Wallet</div>
                <div className="font-mono text-sm truncate text-green-400" title={address}>{address.substring(0,6)}...{address.substring(38)}</div>
                <div className="text-xs text-gray-300 mt-1">{balance ? parseFloat(balance).toFixed(4) : '0'} ETH</div>
                <button onClick={disconnect} className="text-xs text-red-400 hover:text-red-300 mt-2">Disconnect</button>
            </div>
        );
    }

    return (
        <button
            onClick={connect}
            disabled={isConnecting}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2 px-4 rounded mb-6 text-sm font-bold"
        >
            {isConnecting ? 'Connecting...' : 'Connect Wallet'}
        </button>
    );
};

function App() {
  return (
    <WalletProvider>
    <Router>
      <div className="min-h-screen bg-gray-100 flex">
        {/* Sidebar */}
        <div className="w-64 bg-gray-900 text-white p-4">
          <h1 className="text-2xl font-bold mb-8 text-blue-400">PowerTrader AI</h1>

          <WalletButton />

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
    </WalletProvider>
  );
}

export default App;
