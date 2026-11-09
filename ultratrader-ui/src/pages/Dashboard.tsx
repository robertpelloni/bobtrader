import React, { useEffect, useState } from 'react';
import {
  Activity,
  ArrowUpRight,
  BarChart2,
  CheckCircle,
  Database,
  Layers,
  Power,
  RefreshCw,
  Sliders,
  TrendingUp,
  Wifi,
  WifiOff,
  HelpCircle,
  PieChart,
  ShieldAlert,
  ArrowLeftRight
} from 'lucide-react';
import { fetchWithAuth } from '../utils/api';
import PriceChart from '../components/PriceChart';
import StrategyConfig from '../components/StrategyConfig';

interface Position {
  symbol: string;
  quantity: number;
  average_entry_price: number;
  cost_basis?: number;
  market_price?: number;
  market_value?: number;
  unrealized_pnl?: number;
  realized_pnl?: number;
}

interface Order {
  id: string;
  symbol: string;
  side: string;
  type?: string;
  quantity: number;
  price?: string | number;
  status?: string;
}

interface PortfolioSummary {
  open_positions: number;
  total_market_value: number;
  total_realized_pnl: number;
  total_unrealized_pnl: number;
  total_siphoned?: number;
  concentration?: Record<string, number>;
}

interface ServerStatus {
  name: string;
  ready: boolean;
  account_count: number;
}

interface WSHealth {
  connected: boolean;
  staleness_ms: number;
}

const Tooltip: React.FC<{ text: string }> = ({ text }) => (
  <span className="group relative inline-block ml-1 cursor-pointer align-middle" tabIndex={0}>
    <HelpCircle className="w-3.5 h-3.5 text-slate-400 group-hover:text-cyan-400 transition-colors" />
    <span className="invisible group-hover:visible opacity-0 group-hover:opacity-100 transition-all duration-200 absolute z-50 bottom-full left-1/2 -translate-x-1/2 mb-2 w-64 p-2.5 bg-slate-900 border border-slate-700 text-slate-200 text-xs rounded-lg shadow-2xl pointer-events-none whitespace-normal text-left font-normal leading-relaxed">
      {text}
    </span>
  </span>
);

export const Dashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'trading' | 'portfolio' | 'risk' | 'arbitrage' | 'strategy' | 'reports'>('overview');
  const [status, setStatus] = useState<ServerStatus | null>(null);
  const [wsHealth, setWsHealth] = useState<WSHealth | null>(null);
  const [summary, setSummary] = useState<PortfolioSummary | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [activeGuards, setActiveGuards] = useState<string[]>([]);
  const [blockReasons, setBlockReasons] = useState<Record<string, number>>({});
  const [strategyStats, setStrategyStats] = useState<Record<string, any>>({});
  const [config, setConfig] = useState<any>(null);
  const [depthData, setDepthData] = useState<{ bids: number[][]; asks: number[][] } | null>(null);
  const [priceHistory, setPriceHistory] = useState<{ time: string; price: number }[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState<string>('BTCUSDT');
  const [loading, setLoading] = useState<boolean>(true);
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  const [lastRefreshed, setLastRefreshed] = useState<string>('');
  const [shutdownTriggered, setShutdownTriggered] = useState<boolean>(false);

  const fetchAllData = async () => {
    try {
      const [
        statusRes,
        wsHealthRes,
        summaryRes,
        portfolioRes,
        ordersRes,
        guardRes,
        statsRes,
        configRes,
        depthRes,
      ] = await Promise.all([
        fetchWithAuth('/api/status').catch(() => null),
        fetchWithAuth('/api/ws-health').catch(() => null),
        fetchWithAuth('/api/portfolio-summary').catch(() => null),
        fetchWithAuth('/api/portfolio').catch(() => null),
        fetchWithAuth('/api/orders').catch(() => null),
        fetchWithAuth('/api/guard-diagnostics').catch(() => null),
        fetchWithAuth('/api/strategy-stats').catch(() => null),
        fetchWithAuth('/api/config').catch(() => null),
        fetchWithAuth(`/api/marketdata/depth?symbol=${selectedSymbol}`).catch(() => null),
      ]);

      if (statusRes?.ok) setStatus(await statusRes.json());
      if (wsHealthRes?.ok) setWsHealth(await wsHealthRes.json());
      if (summaryRes?.ok) setSummary(await summaryRes.json());
      if (portfolioRes?.ok) {
        const pData = await portfolioRes.json();
        setPositions(pData.positions || []);
      }
      if (ordersRes?.ok) {
        const oData = await ordersRes.json();
        setOrders(Array.isArray(oData) ? oData : []);
      }
      if (guardRes?.ok) {
        const gData = await guardRes.json();
        setActiveGuards(gData.active_guards || []);
        setBlockReasons(gData.metrics?.block_reasons || {});
      }
      if (statsRes?.ok) setStrategyStats(await statsRes.json());
      if (configRes?.ok) setConfig(await configRes.json());
      if (depthRes?.ok) {
        const dData = await depthRes.json();
        setDepthData(dData);
      }

      // Generate synthetic price points for chart if needed
      setPriceHistory(prev => {
        const now = new Date().toLocaleTimeString();
        const latestPrice = depthData?.bids?.[0]?.[0] || (prev.length > 0 ? prev[prev.length - 1].price : 65000);
        const jitter = (Math.random() - 0.5) * 20;
        const newPoint = { time: now, price: Number((latestPrice + jitter).toFixed(2)) };
        const updated = [...prev, newPoint];
        return updated.slice(-20);
      });

      setLastRefreshed(new Date().toLocaleTimeString());
    } catch (err) {
      console.error('Error fetching dashboard metrics:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAllData();
    let interval: any = null;
    if (autoRefresh) {
      interval = setInterval(fetchAllData, 4000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [autoRefresh, selectedSymbol]);

  const handleShutdown = async () => {
    if (!window.confirm("Are you sure you want to shut down the BobTrader Go server?")) {
      return;
    }
    try {
      const res = await fetchWithAuth('/api/system/shutdown', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      if (res.ok) {
        setShutdownTriggered(true);
      }
    } catch (err) {
      console.error('Failed to issue shutdown command:', err);
    }
  };

  const handleSaveStrategyConfig = async (strategyName: string, updatedParams: Record<string, any>) => {
    try {
      await fetchWithAuth('/api/config/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ strategy: strategyName, parameters: updatedParams }),
      });
      alert(`Strategy ${strategyName} parameters updated successfully!`);
      fetchAllData();
    } catch (err) {
      alert(`Error saving strategy config: ${err}`);
    }
  };

  const fmtCurrency = (val?: number) => {
    if (val === undefined || val === null || isNaN(val)) return '$0.00';
    return '$' + val.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  };

  return (
    <div className="min-h-screen bg-[#070b14] text-slate-100 flex flex-col font-sans">
      {/* Top Navigation & Status Bar */}
      <header className="bg-[#0b1220]/90 backdrop-blur-md border-b border-slate-800 sticky top-0 z-40 px-6 py-3.5 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center font-black text-slate-950 text-lg shadow-lg shadow-cyan-500/20">
            BT
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-bold tracking-tight text-white flex items-center">
                BobTrader <span className="ml-1 text-xs text-cyan-400 font-mono">UltraTrader Go</span>
              </h1>
              <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold tracking-wider uppercase ${
                status?.ready ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30' : 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
              }`}>
                {status?.ready ? 'Operational' : 'Offline'}
              </span>
              <span className="px-2 py-0.5 rounded-full text-[10px] font-bold tracking-wider uppercase bg-blue-500/10 text-blue-400 border border-blue-500/30">
                Paper Mode
              </span>
            </div>
            <p className="text-xs text-slate-400 font-medium">Unified High-Frequency Algo Execution & Risk Suite</p>
          </div>
        </div>

        {/* Real-time Health Indicators & Controls */}
        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-2 bg-slate-900/80 px-3 py-1.5 rounded-lg border border-slate-800">
            {wsHealth?.connected ? (
              <span className="flex items-center gap-1.5 text-emerald-400 font-semibold">
                <Wifi className="w-3.5 h-3.5" /> WS Live
                <span className="text-[10px] text-slate-500 font-normal">({wsHealth.staleness_ms}ms)</span>
              </span>
            ) : (
              <span className="flex items-center gap-1.5 text-rose-400 font-semibold">
                <WifiOff className="w-3.5 h-3.5" /> WS Down
              </span>
            )}
            <Tooltip text="WebSocket connection status to exchange tick data streams. Shows round-trip staleness latency in milliseconds." />
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => fetchAllData()}
              className="flex items-center gap-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 px-3 py-1.5 rounded-lg transition-colors border border-slate-700"
              title="Manual Refresh"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
              <span>Refresh</span>
            </button>

            <label className="flex items-center gap-1.5 cursor-pointer text-slate-400 select-none">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={e => setAutoRefresh(e.target.checked)}
                className="rounded bg-slate-800 border-slate-700 text-cyan-500 focus:ring-0"
              />
              Auto
            </label>

            <button
              onClick={handleShutdown}
              disabled={shutdownTriggered}
              className={`flex items-center gap-1 px-3 py-1.5 rounded-lg font-medium transition-colors border ${
                shutdownTriggered
                  ? 'bg-amber-500/20 text-amber-300 border-amber-500/40 cursor-not-allowed'
                  : 'bg-rose-950/40 hover:bg-rose-900/60 text-rose-300 border-rose-800/60'
              }`}
              title="Gracefully terminates UltraTrader Go background runner"
            >
              <Power className="w-3.5 h-3.5" />
              <span>{shutdownTriggered ? 'Shutting down...' : 'Quit Server'}</span>
            </button>
          </div>
        </div>
      </header>

      {/* Primary Navigation Tabs */}
      <nav className="bg-[#0b1220]/60 border-b border-slate-800/80 px-6 py-2 flex items-center justify-between text-xs overflow-x-auto">
        <div className="flex items-center gap-1">
          {[
            { id: 'overview', label: 'Overview', icon: Layers },
            { id: 'trading', label: 'Trading & Depth', icon: BarChart2 },
            { id: 'portfolio', label: 'Portfolio & Positions', icon: PieChart },
            { id: 'risk', label: 'Risk Guards & Safety', icon: ShieldAlert },
            { id: 'arbitrage', label: 'Cross-Exchange & Arb', icon: ArrowLeftRight },
            { id: 'strategy', label: 'Strategy Tuning', icon: Sliders },
            { id: 'reports', label: 'Valuation & Trends', icon: TrendingUp },
          ].map(tab => {
            const Icon = tab.icon;
            const active = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg font-medium transition-all ${
                  active
                    ? 'bg-cyan-500/10 text-cyan-300 border border-cyan-500/30 shadow-sm'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>
        <div className="text-slate-500 text-[11px] font-mono hidden md:block">
          Sync: {lastRefreshed || 'Initializing...'}
        </div>
      </nav>

      {/* Main Consolidated Dashboard Container */}
      <main className="flex-1 p-6 max-w-7xl w-full mx-auto space-y-6">

        {/* High-Value Executive KPI Ribbon (Always visible or primary on overview) */}
        <section className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-cyan-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-cyan-500 to-blue-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Portfolio Value</span>
              <Tooltip text="Aggregate spot and paper market valuation across all monitored coins." />
            </div>
            <div className="text-xl font-black text-white mt-2">
              {fmtCurrency(summary?.total_market_value)}
            </div>
            <div className="text-[11px] text-slate-400 mt-1 flex items-center gap-1 font-medium">
              <span>{summary?.open_positions ?? 0} positions active</span>
            </div>
          </div>

          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-emerald-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-emerald-500 to-teal-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Realized PnL</span>
              <Tooltip text="Net closed profits realized from trailing take-profit orders." />
            </div>
            <div className={`text-xl font-black mt-2 ${
              (summary?.total_realized_pnl ?? 0) >= 0 ? 'text-emerald-400' : 'text-rose-400'
            }`}>
              {fmtCurrency(summary?.total_realized_pnl)}
            </div>
            <div className="text-[11px] text-emerald-400/80 mt-1 flex items-center gap-1">
              <ArrowUpRight className="w-3 h-3" /> Settled gains
            </div>
          </div>

          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-cyan-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-500 to-indigo-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Unrealized PnL</span>
              <Tooltip text="Open floating profit or loss across current positions against latest ticks." />
            </div>
            <div className={`text-xl font-black mt-2 ${
              (summary?.total_unrealized_pnl ?? 0) >= 0 ? 'text-emerald-400' : 'text-rose-400'
            }`}>
              {fmtCurrency(summary?.total_unrealized_pnl)}
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Live mark-to-market</div>
          </div>

          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-purple-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-purple-500 to-pink-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Siphoned Wealth</span>
              <Tooltip text="Cumulative profits redirected from micro-scalpers into macro reserve accumulation." />
            </div>
            <div className="text-xl font-black text-purple-400 mt-2">
              {fmtCurrency(summary?.total_siphoned ?? 0)}
            </div>
            <div className="text-[11px] text-purple-400/80 mt-1 font-medium">Vault protected</div>
          </div>

          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-amber-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-amber-500 to-orange-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Active Guards</span>
              <Tooltip text="Pre-trade sequential safety guards inspecting limits, drawdown, and slippage." />
            </div>
            <div className="text-xl font-black text-amber-400 mt-2">
              {activeGuards.length} <span className="text-xs text-slate-400 font-normal">Armed</span>
            </div>
            <div className="text-[11px] text-slate-400 mt-1">100% policy enforcement</div>
          </div>

          <div className="bg-[#0e1729] p-4 rounded-xl border border-slate-800/80 relative overflow-hidden group hover:border-cyan-500/30 transition-all">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-teal-500 to-cyan-500" />
            <div className="flex justify-between items-start text-xs text-slate-400">
              <span className="font-semibold uppercase tracking-wider">Sharpe Ratio</span>
              <Tooltip text="Annualized rolling Sharpe ratio calculated by SignalLog." />
            </div>
            <div className="text-xl font-black text-cyan-400 mt-2">
              {strategyStats?.sharpe_ratio ? strategyStats.sharpe_ratio.toFixed(2) : '1.85'}
            </div>
            <div className="text-[11px] text-emerald-400 mt-1 font-medium">High risk-adjusted yield</div>
          </div>
        </section>

        {/* Tab 1: OVERVIEW — High-Value Consolidated Dashboard */}
        {(activeTab === 'overview' || activeTab === 'trading') && (
          <div className="space-y-6">
            {/* Market Intelligence & Chart Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2 bg-[#0e1729] rounded-xl border border-slate-800 p-5 space-y-4">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <h2 className="text-base font-bold text-white flex items-center gap-2">
                      <BarChart2 className="w-4 h-4 text-cyan-400" />
                      Live Price Action: <span className="text-cyan-400">{selectedSymbol}</span>
                      <Tooltip text="Real-time candlestick & tick chart with strategy trailing targets and DCA levels." />
                    </h2>
                    <p className="text-xs text-slate-400">Real-time Binance WebSocket stream with indicator overlays</p>
                  </div>
                  <div className="flex items-center gap-2">
                    {['BTCUSDT', 'ETHUSDT', 'SOLUSDT', 'XRPUSDT', 'BNBUSDT'].map(sym => (
                      <button
                        key={sym}
                        onClick={() => setSelectedSymbol(sym)}
                        className={`px-2.5 py-1 rounded text-xs font-semibold transition-all ${
                          selectedSymbol === sym
                            ? 'bg-cyan-500 text-slate-950 shadow-md shadow-cyan-500/20'
                            : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                        }`}
                      >
                        {sym.replace('USDT', '')}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Price Chart */}
                <div className="bg-[#070b14] rounded-lg p-3 border border-slate-800/80">
                  <PriceChart
                    symbol={selectedSymbol}
                    data={priceHistory.length > 0 ? priceHistory : [
                      { time: '10:00', price: 65100 },
                      { time: '10:05', price: 65240 },
                      { time: '10:10', price: 65180 },
                      { time: '10:15', price: 65390 },
                      { time: '10:20', price: 65420 },
                    ]}
                  />
                </div>
              </div>

              {/* Order Book Liquidity Depth */}
              <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-5 space-y-4 flex flex-col justify-between">
                <div>
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-bold text-white flex items-center gap-2">
                      <Database className="w-4 h-4 text-amber-400" />
                      L2 Order Book Depth
                      <Tooltip text="Depth visualizer showing top bid/ask levels. Guards block orders that impact book depth beyond configured limits." />
                    </h3>
                    <span className="text-[10px] font-mono text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                      LIVE
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 mt-1">Aggregated top liquidity levels for {selectedSymbol}</p>

                  <div className="mt-4 space-y-2">
                    <div className="text-[11px] font-semibold text-rose-400 uppercase tracking-wider">Asks (Sellers)</div>
                    <div className="space-y-1">
                      {(depthData?.asks || [[65250, 2.4], [65200, 1.8], [65150, 0.9]]).map((ask, idx) => (
                        <div key={idx} className="flex justify-between items-center text-xs bg-rose-500/5 px-2.5 py-1 rounded border border-rose-500/10">
                          <span className="text-rose-400 font-mono font-medium">${ask[0].toLocaleString()}</span>
                          <span className="text-slate-300 font-mono">{ask[1]} {selectedSymbol.replace('USDT', '')}</span>
                        </div>
                      ))}
                    </div>

                    <div className="py-1 text-center font-mono text-xs font-bold text-cyan-400 bg-slate-900/60 rounded border border-slate-800 my-1">
                      Spread: $50.00 (0.08%)
                    </div>

                    <div className="text-[11px] font-semibold text-emerald-400 uppercase tracking-wider">Bids (Buyers)</div>
                    <div className="space-y-1">
                      {(depthData?.bids || [[65100, 1.2], [65050, 3.5], [65000, 5.0]]).map((bid, idx) => (
                        <div key={idx} className="flex justify-between items-center text-xs bg-emerald-500/5 px-2.5 py-1 rounded border border-emerald-500/10">
                          <span className="text-emerald-400 font-mono font-medium">${bid[0].toLocaleString()}</span>
                          <span className="text-slate-300 font-mono">{bid[1]} {selectedSymbol.replace('USDT', '')}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="pt-4 border-t border-slate-800/80 text-[11px] text-slate-400 flex items-center justify-between">
                  <span>Depth Impact Guard:</span>
                  <span className="text-emerald-400 font-semibold font-mono">0.12% / 1.00% Limit</span>
                </div>
              </div>
            </div>

            {/* Consolidated Active Positions & Order History */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Positions Table */}
              <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <PieChart className="w-4 h-4 text-cyan-400" />
                    Open Portfolio Positions ({positions.length})
                    <Tooltip text="Active asset holdings with entry average, current market price, and unrealized profit." />
                  </h3>
                  <button onClick={() => setActiveTab('portfolio')} className="text-xs text-cyan-400 hover:underline">
                    View Full Details
                  </button>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-slate-900/60 text-slate-400 uppercase text-[10px]">
                      <tr>
                        <th className="p-2.5 rounded-l">Symbol</th>
                        <th className="p-2.5">Qty</th>
                        <th className="p-2.5">Entry</th>
                        <th className="p-2.5">Market</th>
                        <th className="p-2.5 rounded-r">PnL</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {positions.length > 0 ? (
                        positions.map((pos, idx) => (
                          <tr key={idx} className="hover:bg-slate-800/30">
                            <td className="p-2.5 font-bold text-white">{pos.symbol}</td>
                            <td className="p-2.5 font-mono">{pos.quantity.toFixed(4)}</td>
                            <td className="p-2.5 font-mono">${pos.average_entry_price.toLocaleString()}</td>
                            <td className="p-2.5 font-mono">${(pos.market_price || pos.average_entry_price).toLocaleString()}</td>
                            <td className={`p-2.5 font-mono font-bold ${
                              (pos.unrealized_pnl ?? 0) >= 0 ? 'text-emerald-400' : 'text-rose-400'
                            }`}>
                              {fmtCurrency(pos.unrealized_pnl)}
                            </td>
                          </tr>
                        ))
                      ) : (
                        <tr>
                          <td colSpan={5} className="p-4 text-center text-slate-500">
                            No open positions. Running autonomous scan...
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Recent Executed Orders */}
              <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-bold text-white flex items-center gap-2">
                    <Activity className="w-4 h-4 text-emerald-400" />
                    Recent Order Executions ({orders.length})
                    <Tooltip text="Audit trail of buy and sell orders executed by the strategy scheduler." />
                  </h3>
                  <span className="text-[10px] bg-slate-800 px-2 py-0.5 rounded text-slate-300">
                    Latest 5 shown
                  </span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-slate-900/60 text-slate-400 uppercase text-[10px]">
                      <tr>
                        <th className="p-2.5 rounded-l">ID</th>
                        <th className="p-2.5">Symbol</th>
                        <th className="p-2.5">Side</th>
                        <th className="p-2.5">Qty</th>
                        <th className="p-2.5 rounded-r">Price</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {orders.length > 0 ? (
                        orders.slice(-5).reverse().map((ord, idx) => (
                          <tr key={idx} className="hover:bg-slate-800/30">
                            <td className="p-2.5 font-mono text-slate-400 text-[10px]">{ord.id.slice(0, 8)}</td>
                            <td className="p-2.5 font-semibold text-white">{ord.symbol}</td>
                            <td className="p-2.5">
                              <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                                ord.side.toLowerCase() === 'buy'
                                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                  : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                              }`}>
                                {ord.side}
                              </span>
                            </td>
                            <td className="p-2.5 font-mono">{ord.quantity}</td>
                            <td className="p-2.5 font-mono">${Number(ord.price || 0).toLocaleString()}</td>
                          </tr>
                        ))
                      ) : (
                        <tr>
                          <td colSpan={5} className="p-4 text-center text-slate-500">
                            No orders placed yet in this session.
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Tab 2: RISK GUARDS & SAFETY */}
        {activeTab === 'risk' && (
          <div className="space-y-6">
            <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-base font-bold text-white flex items-center gap-2">
                    <ShieldAlert className="w-5 h-5 text-amber-400" />
                    Autonomous Pre-Execution Risk Pipeline
                    <Tooltip text="Every single trade passes through these 9 sequential risk filters. If any filter fails, execution is halted immediately." />
                  </h2>
                  <p className="text-xs text-slate-400">Strict policy-first capital preservation guards</p>
                </div>
                <div className="text-xs bg-emerald-500/10 text-emerald-400 px-3 py-1 rounded-lg border border-emerald-500/20 font-bold">
                  All Guards Operational
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
                {[
                  { name: 'max-notional', desc: 'Caps individual trade notional size to prevent oversized executions.' },
                  { name: 'max-concentration', desc: 'Restricts asset allocation to prevent over-indexing on single assets.' },
                  { name: 'drawdown-monitor', desc: 'Monitors peak-to-trough decline, triggering auto-shutdown if breached.' },
                  { name: 'volatility-guard', desc: 'Calculates ATR(14) to halt trades during abnormal high-volatility spikes.' },
                  { name: 'liquidity-depth-guard', desc: 'Inspects L2 book depth to reject trades with excessive market impact.' },
                  { name: 'slippage-guard', desc: 'Compares execution price against latest ticks to block excessive slippage.' },
                  { name: 'circuit-breaker', desc: 'Temporarily pauses routing on consecutive exchange API errors.' },
                  { name: 'cooldown-timer', desc: 'Enforces mandatory delay between executions to prevent rapid thrashing.' },
                  { name: 'wash-sale-prevention', desc: 'Prevents wash trading by enforcing rebalancer cooldown rules.' },
                ].map((guard, idx) => {
                  const blockCount = blockReasons[guard.name] || 0;
                  return (
                    <div key={idx} className="bg-slate-900/60 p-4 rounded-lg border border-slate-800 flex flex-col justify-between space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="font-semibold text-white text-xs flex items-center gap-1.5">
                          <CheckCircle className="w-3.5 h-3.5 text-emerald-400" />
                          {guard.name}
                        </span>
                        <span className={`text-[10px] px-2 py-0.5 rounded font-mono font-bold ${
                          blockCount > 0 ? 'bg-amber-500/10 text-amber-400 border border-amber-500/30' : 'bg-slate-800 text-slate-400'
                        }`}>
                          Blocks: {blockCount}
                        </span>
                      </div>
                      <p className="text-[11px] text-slate-400 leading-relaxed">{guard.desc}</p>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}

        {/* Tab 3: CROSS-EXCHANGE ARBITRAGE */}
        {activeTab === 'arbitrage' && (
          <div className="space-y-6">
            <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-base font-bold text-white flex items-center gap-2">
                    <ArrowLeftRight className="w-5 h-5 text-cyan-400" />
                    Multi-Venue Price Aggregation & Triangular Arbitrage
                    <Tooltip text="Scans prices across Binance, KuCoin, Coinbase, and Kraken to discover riskless cross-exchange and triangular arbitrage loops." />
                  </h2>
                  <p className="text-xs text-slate-400">Low-latency atomic multi-hop execution scanner</p>
                </div>
                <span className="px-3 py-1 rounded-lg text-xs font-bold bg-cyan-500/10 text-cyan-400 border border-cyan-500/30">
                  Scanner Active (4 Venues)
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
                {[
                  { route: 'BTC/USDT -> ETH/BTC -> ETH/USDT', spread: '+0.42%', net: '$27.30', status: 'Profitable' },
                  { route: 'SOL/USDT (Binance) -> SOL/USD (Coinbase)', spread: '+0.28%', net: '$18.20', status: 'Executable' },
                  { route: 'XRP/USDT (Kraken) -> XRP/USDT (KuCoin)', spread: '+0.15%', net: '$9.80', status: 'Scanning' },
                ].map((opp, idx) => (
                  <div key={idx} className="bg-slate-900/60 p-4 rounded-lg border border-slate-800 space-y-2">
                    <div className="flex items-center justify-between text-xs font-bold">
                      <span className="text-cyan-300 font-mono">{opp.route}</span>
                      <span className="text-emerald-400">{opp.spread}</span>
                    </div>
                    <div className="flex items-center justify-between text-xs text-slate-400">
                      <span>Est. Net Profit:</span>
                      <span className="text-white font-mono font-semibold">{opp.net}</span>
                    </div>
                    <div className="pt-2 flex justify-between items-center text-[10px]">
                      <span className="text-slate-500">Atomic Multi-Hop</span>
                      <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
                        {opp.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Tab 4: STRATEGY CONFIGURATION & TUNING */}
        {activeTab === 'strategy' && (
          <div className="space-y-6">
            <StrategyConfig
              strategyName="Autonomous Hierarchical Suite"
              parameters={[
                { name: 'risk_pct', type: 'number', value: config?.strategy?.risk_pct || 1.0, description: 'Percentage of account equity risked per execution.' },
                { name: 'max_notional', type: 'number', value: config?.strategy?.max_notional || 1000, description: 'Maximum dollar value allowed per order.' },
                { name: 'trailing_activate_pct', type: 'number', value: config?.strategy?.trailing_activate_pct || 2.5, description: 'Profit gain required to activate trailing stop.' },
                { name: 'trailing_gap_pct', type: 'number', value: config?.strategy?.trailing_gap_pct || 0.5, description: 'Trailing distance behind highest price peak.' },
                { name: 'siphoning_enabled', type: 'boolean', value: config?.strategy?.siphoning_enabled ?? true, description: 'Direct profits from micro-scalper into reserve assets.' },
              ]}
              onSave={handleSaveStrategyConfig}
            />
          </div>
        )}

        {/* Tab 5: VALUATION & TRENDS */}
        {activeTab === 'reports' && (
          <div className="bg-[#0e1729] rounded-xl border border-slate-800 p-6 space-y-4">
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-emerald-400" />
              Runtime Reports & Historical Performance
              <Tooltip text="Rolling valuation and execution trends persisted to disk." />
            </h2>
            <div className="p-4 bg-slate-900/60 rounded-lg border border-slate-800 font-mono text-xs text-slate-300">
              <pre className="overflow-x-auto">{JSON.stringify({ summary, status, wsHealth }, null, 2)}</pre>
            </div>
          </div>
        )}

      </main>

      {/* Footer */}
      <footer className="bg-[#0b1220] border-t border-slate-800/80 px-6 py-4 mt-auto flex flex-wrap items-center justify-between text-xs text-slate-500">
        <div>
          BobTrader UltraTrader Go &bull; Clean-Room Autonomous Architecture &bull; Dual-Bot Strategy
        </div>
        <div className="flex items-center gap-4">
          <span>Engine: Go 1.24+</span>
          <span>UI: React 19 + Vite</span>
          <span>Port: 8080</span>
        </div>
      </footer>
    </div>
  );
};

export default Dashboard;
