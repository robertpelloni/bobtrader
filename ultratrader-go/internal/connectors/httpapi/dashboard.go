package httpapi

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>UltraTrader Go</title>
<style>
:root {
  --bg: #070d1a;
  --bg2: #0b1220;
  --panel: #0e1729;
  --panel2: #121e34;
  --border: #1e3050;
  --border2: #2a4060;
  --text: #d0dced;
  --text2: #8ea4c2;
  --text3: #5d7490;
  --green: #00e676;
  --green2: #0ac28a;
  --red: #ff5252;
  --red2: #f07178;
  --orange: #ffab40;
  --blue: #448aff;
  --blue2: #42a5f5;
  --purple: #b388ff;
  --cyan: #18ffff;
  --sidebar-w: 220px;
  color-scheme: dark;
}
*, *::before, *::after { box-sizing: border-box; }
html, body { margin: 0; padding: 0; height: 100%; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Inter, Roboto, sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.5;
  display: flex;
}
a { color: var(--blue2); text-decoration: none; }
a:hover { text-decoration: underline; }

/* ── Sidebar ─────────────────────────────────────────────── */
.sidebar {
  width: var(--sidebar-w);
  min-width: var(--sidebar-w);
  background: var(--bg2);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  height: 100vh;
  position: sticky;
  top: 0;
  overflow-y: auto;
  z-index: 10;
}
.sidebar-brand {
  padding: 20px 16px 12px;
  border-bottom: 1px solid var(--border);
}
.sidebar-brand h1 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text);
}
.sidebar-brand .tag {
  display: inline-block;
  font-size: 10px;
  background: var(--green);
  color: #000;
  border-radius: 3px;
  padding: 1px 5px;
  font-weight: 700;
  margin-top: 4px;
  vertical-align: middle;
}
.sidebar-brand .tag.offline {
  background: var(--red);
  color: #fff;
}
.sidebar-nav {
  padding: 8px 0;
  flex: 1;
}
.sidebar-nav a {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  color: var(--text2);
  font-size: 13px;
  font-weight: 500;
  border-left: 3px solid transparent;
  transition: all 0.15s;
}
.sidebar-nav a:hover {
  color: var(--text);
  background: rgba(255,255,255,0.04);
  text-decoration: none;
}
.sidebar-nav a.active {
  color: var(--cyan);
  border-left-color: var(--cyan);
  background: rgba(24,255,255,0.06);
}
.sidebar-nav a .icon {
  width: 18px;
  text-align: center;
  font-size: 15px;
}
.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border);
  font-size: 11px;
  color: var(--text3);
}

/* ── Main Content ────────────────────────────────────────── */
.main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  height: 100vh;
}
.topbar {
  position: sticky;
  top: 0;
  z-index: 5;
  background: rgba(7,13,26,0.85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
  padding: 12px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.topbar-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}
.topbar-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}
.btn {
  background: var(--panel2);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.btn:hover { background: var(--border); }
.btn-primary {
  background: var(--cyan);
  color: #000;
  border-color: var(--cyan);
}
.btn-primary:hover { opacity: 0.85; }
.topbar .muted {
  font-size: 11px;
  color: var(--text3);
}
select, input[type=checkbox] {
  accent-color: var(--cyan);
}

/* ── Page sections ────────────────────────────────────────── */
.page { display: none; padding: 24px; }
.page.active { display: block; }
.page-wide { display: none; padding: 24px; max-width: 1400px; }
.page-wide.active { display: block; }

/* ── Cards / KPI ─────────────────────────────────────────── */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}
.kpi {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  position: relative;
  overflow: hidden;
}
.kpi::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  border-radius: 10px 10px 0 0;
}
.kpi.kpi-green::before { background: var(--green); }
.kpi.kpi-blue::before { background: var(--blue); }
.kpi.kpi-red::before { background: var(--red); }
.kpi.kpi-orange::before { background: var(--orange); }
.kpi.kpi-purple::before { background: var(--purple); }
.kpi.kpi-cyan::before { background: var(--cyan); }
.kpi-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text2);
  margin-bottom: 4px;
}
.kpi-value {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.2;
}
.kpi-delta {
  font-size: 11px;
  margin-top: 4px;
  font-weight: 600;
}
.kpi-delta.up { color: var(--green); }
.kpi-delta.down { color: var(--red); }
.kpi-delta.flat { color: var(--text3); }

/* ── Panels ─────────────────────────────────────────────── */
.panel {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  margin-bottom: 16px;
  overflow: hidden;
}
.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.panel-title {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}
.panel-body { padding: 16px; }
.panel-body.compact { padding: 0; }

/* ── Tables ──────────────────────────────────────────────── */
.table-wrap { overflow-x: auto; }
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
th {
  text-align: left;
  padding: 8px 12px;
  color: var(--text2);
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
  cursor: pointer;
  user-select: none;
}
th:hover { color: var(--text); }
th .sort-arrow { margin-left: 4px; opacity: 0.4; }
th.sorted .sort-arrow { opacity: 1; color: var(--cyan); }
td {
  padding: 8px 12px;
  border-bottom: 1px solid rgba(30,48,80,0.4);
  white-space: nowrap;
}
tr:hover td { background: rgba(255,255,255,0.02); }
.positive { color: var(--green); font-weight: 600; }
.negative { color: var(--red); font-weight: 600; }
.neutral { color: var(--text3); }

/* ── Badges ──────────────────────────────────────────────── */
.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.badge-buy { background: rgba(0,230,118,0.15); color: var(--green); }
.badge-sell { background: rgba(255,82,82,0.15); color: var(--red); }
.badge-ok { background: rgba(0,230,118,0.12); color: var(--green); }
.badge-warn { background: rgba(255,171,64,0.12); color: var(--orange); }
.badge-blocked { background: rgba(255,82,82,0.12); color: var(--red); }
.badge-info { background: rgba(68,138,255,0.12); color: var(--blue2); }

/* ── Guard indicators ────────────────────────────────────── */
.guard-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 8px;
}
.guard-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: var(--panel2);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 12px;
}
.guard-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.guard-dot.pass { background: var(--green); box-shadow: 0 0 6px var(--green); }
.guard-dot.block { background: var(--red); box-shadow: 0 0 6px var(--red); }
.guard-name { flex: 1; font-weight: 500; }
.guard-count {
  font-size: 11px;
  color: var(--text3);
  min-width: 24px;
  text-align: right;
}

/* ── Charts ──────────────────────────────────────────────── */
.chart-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}
.chart-box {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
}
.chart-header {
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  font-weight: 600;
}
.chart-canvas {
  width: 100%;
  height: 200px;
  position: relative;
}
.chart-canvas svg {
  width: 100%;
  height: 100%;
  display: block;
}
.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text3);
  font-size: 12px;
}

/* ── JSON viewer ─────────────────────────────────────────── */
.json-view {
  font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace;
  font-size: 11px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 400px;
  overflow-y: auto;
  padding: 12px;
  background: var(--bg2);
  border-radius: 6px;
  color: var(--text);
}

/* ── Quick start ─────────────────────────────────────────── */
.quickstart {
  background: linear-gradient(135deg, rgba(24,255,255,0.06), rgba(68,138,255,0.06));
  border: 1px solid var(--border2);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
}
.quickstart h2 {
  margin: 0 0 8px;
  font-size: 18px;
}
.quickstart p { color: var(--text2); margin: 0 0 16px; font-size: 13px; }
.steps {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}
.step {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px;
}
.step-num {
  display: inline-block;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--cyan);
  color: #000;
  text-align: center;
  line-height: 22px;
  font-size: 12px;
  font-weight: 700;
  margin-bottom: 8px;
}
.step-title { font-weight: 600; font-size: 13px; margin-bottom: 4px; }
.step-desc { font-size: 12px; color: var(--text2); }
.step code {
  background: var(--bg2);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  color: var(--cyan);
}

/* ── Config viewer ───────────────────────────────────────── */
.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
}
.config-section {
  background: var(--panel2);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px;
}
.config-section h3 {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--cyan);
}
.config-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 12px;
  border-bottom: 1px solid rgba(30,48,80,0.3);
}
.config-row:last-child { border-bottom: none; }
.config-key { color: var(--text2); }
.config-val { color: var(--text); font-weight: 500; font-family: monospace; }

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 768px) {
  .sidebar { display: none; }
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .chart-grid { grid-template-columns: 1fr; }
  .steps { grid-template-columns: 1fr; }
}

/* ── Animations ──────────────────────────────────────────── */
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
.live-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--green);
  animation: pulse 2s infinite;
  margin-right: 6px;
  vertical-align: middle;
}
.live-dot.offline { background: var(--red); animation: none; }
</style>
</head>
<body>

<!-- Sidebar Navigation -->
<nav class="sidebar">
  <div class="sidebar-brand">
    <h1>UltraTrader Go</h1>
    <span class="tag" id="status-tag">LIVE</span>
  </div>
  <div class="sidebar-nav">
    <a href="#" data-page="overview" class="active">
      <span class="icon">&#9632;</span> Overview
    </a>
    <a href="#" data-page="portfolio">
      <span class="icon">&#9670;</span> Portfolio
    </a>
    <a href="#" data-page="orders">
      <span class="icon">&#9654;</span> Orders
    </a>
    <a href="#" data-page="guards">
      <span class="icon">&#9632;</span> Risk Guards
    </a>
    <a href="#" data-page="charts">
      <span class="icon">&#8801;</span> Charts
    </a>
    <a href="#" data-page="reports">
      <span class="icon">&#9776;</span> Reports
    </a>
    <a href="#" data-page="config">
      <span class="icon">&#9881;</span> Configuration
    </a>
    <a href="#" data-page="quickstart">
      <span class="icon">&#9733;</span> Quick Start
    </a>
  </div>
  <div class="sidebar-footer">
    UltraTrader Go v2.0<br>
    <span id="uptime-footer">Loading...</span>
  </div>
</nav>

<!-- Main Content -->
<div class="main">

  <!-- Top Bar -->
  <div class="topbar">
    <span class="topbar-title" id="page-title">Overview</span>
    <div class="topbar-controls">
      <span id="ws-status" style="font-size:11px; margin-right:10px;">WS: ---</span>
      <span class="live-dot" id="live-dot"></span>
      <span class="muted" id="last-updated">Loading...</span>
      <button class="btn" id="refresh-btn">&#8635; Refresh</button>
      <label style="font-size:11px;color:var(--text2)">
        <input type="checkbox" id="auto-refresh" checked> Auto
      </label>
      <select id="refresh-interval" style="background:var(--panel2);color:var(--text);border:1px solid var(--border);border-radius:4px;padding:4px 8px;font-size:11px;">
        <option value="2000">2s</option>
        <option value="5000" selected>5s</option>
        <option value="10000">10s</option>
        <option value="30000">30s</option>
      </select>
    </div>
  </div>

  <!-- ═══════════ OVERVIEW ═══════════ -->
  <div class="page active" id="page-overview">
    <div class="kpi-grid" id="overview-kpis"></div>

    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">Portfolio at a Glance</h3>
      </div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="overview-positions">
            <thead><tr>
              <th>Symbol</th><th>Qty</th><th>Avg Entry</th><th>Market Price</th><th>Value</th><th>Unrealized PnL</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">Recent Orders</h3>
        <span class="badge badge-info" id="order-count-badge">0</span>
      </div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="overview-orders">
            <thead><tr>
              <th>ID</th><th>Symbol</th><th>Side</th><th>Qty</th><th>Price</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">Risk Status</h3>
      </div>
      <div class="panel-body">
        <div class="guard-list" id="overview-guards"></div>
      </div>
    </div>
  </div>

  <!-- ═══════════ PORTFOLIO ═══════════ -->
  <div class="page" id="page-portfolio">
    <div class="kpi-grid" id="portfolio-kpis"></div>
    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">All Positions</h3>
      </div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="portfolio-table">
            <thead><tr>
              <th>Symbol</th><th>Quantity</th><th>Avg Entry</th><th>Cost Basis</th><th>Market Price</th><th>Market Value</th><th>Unrealized PnL</th><th>Realized PnL</th><th>Concentration</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">Concentration Distribution</h3>
      </div>
      <div class="panel-body">
        <div id="concentration-bars"></div>
      </div>
    </div>
  </div>

  <!-- ═══════════ ORDERS ═══════════ -->
  <div class="page" id="page-orders">
    <div class="kpi-grid" id="orders-kpis"></div>
    <div class="panel">
      <div class="panel-header">
        <h3 class="panel-title">Order History</h3>
      </div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="orders-table">
            <thead><tr>
              <th>Order ID</th><th>Symbol</th><th>Side</th><th>Type</th><th>Qty</th><th>Price</th><th>Status</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Execution Summary</h3></div>
      <div class="panel-body">
        <div class="json-view" id="execution-detail">Loading...</div>
      </div>
    </div>
  </div>

  <!-- ═══════════ GUARDS ═══════════ -->
  <div class="page" id="page-guards">
    <div class="kpi-grid" id="guards-kpis"></div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Active Guard Pipeline</h3></div>
      <div class="panel-body">
        <div class="guard-list" id="guards-detail"></div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Block Reasons</h3></div>
      <div class="panel-body">
        <div class="chart-canvas" id="block-reasons-chart"></div>
      </div>
    </div>
  </div>

  <!-- ═══════════ CHARTS ═══════════ -->
  <div class="page" id="page-charts">
    <div class="chart-grid">
      <div class="chart-box">
        <div class="chart-header">Portfolio Value Over Time</div>
        <div class="chart-canvas" id="valuation-chart"></div>
      </div>
      <div class="chart-box">
        <div class="chart-header">Execution Success Rate</div>
        <div class="chart-canvas" id="metrics-chart"></div>
      </div>
    </div>
    <div class="chart-grid">
      <div class="chart-box">
        <div class="chart-header">Exposure Concentration</div>
        <div class="chart-canvas" id="concentration-chart"></div>
      </div>
      <div class="chart-box">
        <div class="chart-header">Blocked Count Trend</div>
        <div class="chart-canvas" id="blocked-trend-chart"></div>
      </div>
    </div>
    <div class="chart-grid">
      <div class="chart-box">
        <div class="chart-header">Siphoned Wealth Trend</div>
        <div class="chart-canvas" id="siphoned-trend-chart"></div>
      </div>
      <div class="chart-box">
        <div class="chart-header">Realized PnL Trend</div>
        <div class="chart-canvas" id="pnl-trend-chart"></div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Metrics History</h3></div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="metrics-history-table">
            <thead><tr>
              <th>Timestamp</th><th>Attempts</th><th>Success</th><th>Blocked</th><th>Success Rate</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Valuation History</h3></div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table id="valuation-history-table">
            <thead><tr>
              <th>Timestamp</th><th>Portfolio Value</th><th>Realized PnL</th><th>Unrealized PnL</th>
            </tr></thead>
            <tbody></tbody>
          </table>
        </div>
      </div>
    </div>
  </div>

  <!-- ═══════════ REPORTS ═══════════ -->
  <div class="page" id="page-reports">
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Trend Analysis</h3></div>
      <div class="panel-body">
        <div class="kpi-grid" id="trend-kpis"></div>
        <div class="json-view" id="trends-detail">Loading...</div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">Latest Reports</h3></div>
      <div class="panel-body">
        <div class="json-view" id="latest-reports">Loading...</div>
      </div>
    </div>
  </div>

  <!-- ═══════════ CONFIG ═══════════ -->
  <div class="page" id="page-config">
    <div class="config-grid" id="config-display"></div>
  </div>

  <!-- ═══════════ QUICK START ═══════════ -->
  <div class="page" id="page-quickstart">
    <div class="quickstart">
      <h2>Welcome to UltraTrader Go</h2>
      <p>
        UltraTrader Go is a clean-room crypto trading platform with policy-first risk management,
        multiple scheduling modes, and a comprehensive guard pipeline. Here's how to get started:
      </p>
      <div class="steps">
        <div class="step">
          <div class="step-num">1</div>
          <div class="step-title">Start the Server</div>
          <div class="step-desc">
            Run <code>go run ./cmd/ultratrader</code> from the project root.
            The dashboard will start on the configured port (default <code>0.0.0.0:8080</code>).
          </div>
        </div>
        <div class="step">
          <div class="step-num">2</div>
          <div class="step-title">Choose a Config</div>
          <div class="step-desc">
            Use <code>--config config/development-timer.json</code> for periodic strategy checks,
            or <code>development-stream.json</code> for tick-driven mode.
          </div>
        </div>
        <div class="step">
          <div class="step-num">3</div>
          <div class="step-title">Watch the Dashboard</div>
          <div class="step-desc">
            Navigate between Overview, Portfolio, Orders, Guards, and Charts using the sidebar.
            Data refreshes automatically every 5 seconds.
          </div>
        </div>
        <div class="step">
          <div class="step-num">4</div>
          <div class="step-title">Review Risk Guards</div>
          <div class="step-desc">
            Every order passes through 9 sequential risk guards before execution.
            Check the <strong>Risk Guards</strong> page to see the pipeline status and block reasons.
          </div>
        </div>
        <div class="step">
          <div class="step-num">5</div>
          <div class="step-title">Monitor Performance</div>
          <div class="step-desc">
            The Charts page shows portfolio value, execution success rates, and concentration trends over time.
            Use these to detect strategy drift or guard over-blocking.
          </div>
        </div>
        <div class="step">
          <div class="step-num">6</div>
          <div class="step-title">Customize Config</div>
          <div class="step-desc">
            Edit your JSON config to adjust risk limits (<code>max_notional</code>, <code>max_concentration_pct</code>),
            scheduler mode, and allowed symbols. See the <strong>Configuration</strong> page.
          </div>
        </div>
      </div>
    </div>

    <div class="panel">
      <div class="panel-header"><h3 class="panel-title">API Endpoints Reference</h3></div>
      <div class="panel-body compact">
        <div class="table-wrap">
          <table>
            <thead><tr><th>Endpoint</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td><code>/healthz</code></td><td>Health check (for load balancers)</td></tr>
              <tr><td><code>/readyz</code></td><td>Readiness check (returns 503 if not ready)</td></tr>
              <tr><td><code>/api/status</code></td><td>Runtime name, ready state, account count</td></tr>
              <tr><td><code>/api/portfolio</code></td><td>All positions with live market values</td></tr>
              <tr><td><code>/api/portfolio-summary</code></td><td>Summary: open positions, concentration, PnL</td></tr>
              <tr><td><code>/api/orders</code></td><td>Order history</td></tr>
              <tr><td><code>/api/execution-summary</code></td><td>Execution stats (total, filled, blocked)</td></tr>
              <tr><td><code>/api/execution-diagnostics</code></td><td>Execution summary + metrics snapshot</td></tr>
              <tr><td><code>/api/exposure-diagnostics</code></td><td>Concentration map, top symbol, exposure</td></tr>
              <tr><td><code>/api/guard-diagnostics</code></td><td>Guard names + per-guard trigger counts</td></tr>
              <tr><td><code>/api/metrics</code></td><td>Rolling window metrics snapshot</td></tr>
              <tr><td><code>/api/guards</code></td><td>Active guard names</td></tr>
              <tr><td><code>/api/runtime-reports/latest</code></td><td>Latest report by type</td></tr>
              <tr><td><code>/api/runtime-reports/history?type=&amp;limit=</code></td><td>Historical reports</td></tr>
              <tr><td><code>/api/runtime-reports/trends</code></td><td>Trend analysis across reports</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>

</div><!-- /main -->

<script>
// ─── State ───────────────────────────────────────────────────
let refreshTimer = null;
let currentPage = 'overview';
let appState = {};

// ─── API helpers ──────────────────────────────────────────────
async function fetchJson(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(url + ' => ' + r.status);
  return r.json();
}

function fmt(n, decimals) {
  if (n === undefined || n === null || Number.isNaN(n)) return '\u2014';
  if (typeof n !== 'number') n = Number(n);
  if (!Number.isFinite(n)) return '\u2014';
  const d = decimals !== undefined ? decimals : (Math.abs(n) >= 1000 ? 2 : 4);
  return n.toLocaleString('en-US', { minimumFractionDigits: d, maximumFractionDigits: d });
}

function fmtPct(n, decimals) {
  if (n === undefined || n === null) return '\u2014';
  return (Number(n) * 100).toFixed(decimals !== undefined ? decimals : 1) + '%';
}

function pnlClass(n) {
  if (n > 0) return 'positive';
  if (n < 0) return 'negative';
  return 'neutral';
}

function deltaHtml(current, previous, suffix) {
  if (previous === undefined || current === undefined) return '';
  const d = current - previous;
  const cls = d > 0 ? 'up' : d < 0 ? 'down' : 'flat';
  const arrow = d > 0 ? '\u2191' : d < 0 ? '\u2193' : '\u2192';
  const s = suffix || '';
  return '<div class="kpi-delta ' + cls + '">' + arrow + ' ' + fmt(Math.abs(d)) + s + '</div>';
}

// ─── Navigation ───────────────────────────────────────────────
document.querySelectorAll('.sidebar-nav a').forEach(link => {
  link.addEventListener('click', function(e) {
    e.preventDefault();
    const page = this.dataset.page;
    switchPage(page);
  });
});

function switchPage(page) {
  currentPage = page;
  document.querySelectorAll('.sidebar-nav a').forEach(a => a.classList.toggle('active', a.dataset.page === page));
  document.querySelectorAll('.page').forEach(p => p.classList.toggle('active', p.id === 'page-' + page));
  document.getElementById('page-title').textContent = {
    overview: 'Overview', portfolio: 'Portfolio', orders: 'Orders',
    guards: 'Risk Guards', charts: 'Charts', reports: 'Reports',
    config: 'Configuration', quickstart: 'Quick Start'
  }[page] || page;
}

// ─── Render: KPI Cards ───────────────────────────────────────
function kpi(label, value, cls, deltaHtml) {
  return '<div class="kpi kpi-' + (cls||'blue') + '"><div class="kpi-label">' + label +
    '</div><div class="kpi-value">' + value + '</div>' + (deltaHtml||'') + '</div>';
}

// ─── Render: Tables ──────────────────────────────────────────
function fillTable(id, rows, cols) {
  const tbody = document.querySelector('#' + id + ' tbody');
  if (!rows || rows.length === 0) {
    tbody.innerHTML = '<tr><td colspan="' + cols.length + '" class="neutral" style="text-align:center;padding:20px;">No data</td></tr>';
    return;
  }
  let html = '';
  for (const row of rows) {
    html += '<tr>' + cols.map(c => '<td class="' + (c.cls ? c.cls(row) : '') + '">' + (c.get ? c.get(row) : '') + '</td>').join('') + '</tr>';
  }
  tbody.innerHTML = html;
}

// ─── Render: Charts ──────────────────────────────────────────
function renderLineChart(elId, rows, extractor, color, unit) {
  const el = document.getElementById(elId);
  if (!el) return;
  if (!rows || rows.length === 0) { el.innerHTML = '<div class="chart-empty">No data yet</div>'; return; }
  const values = rows.map(extractor).filter(v => typeof v === 'number' && Number.isFinite(v));
  if (values.length === 0) { el.innerHTML = '<div class="chart-empty">No numeric data</div>'; return; }

  const W = 600, H = 200, P = 28;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  let points = values.map((v, i) => {
    const x = P + (i * (W - P * 2) / Math.max(1, values.length - 1));
    const y = H - P - ((v - min) / range * (H - P * 2));
    return x.toFixed(1) + ',' + y.toFixed(1);
  }).join(' ');

  // Grid lines
  let grid = '';
  for (let i = 0; i <= 4; i++) {
    const y = P + (i / 4) * (H - P * 2);
    const val = max - (i / 4) * range;
    grid += '<line x1="' + P + '" y1="' + y + '" x2="' + (W - P) + '" y2="' + y + '" stroke="#1e3050" stroke-width="1"/>';
    grid += '<text x="' + (P - 4) + '" y="' + (y + 4) + '" fill="#5d7490" font-size="9" text-anchor="end">' + val.toFixed(val >= 100 ? 0 : 2) + (unit||'') + '</text>';
  }

  el.innerHTML = '<svg viewBox="0 0 ' + W + ' ' + H + '">' +
    '<rect width="' + W + '" height="' + H + '" fill="#0b1220"/>' +
    grid +
    '<polyline fill="none" stroke="' + color + '" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round" points="' + points + '"/>' +
    '</svg>';
}

function renderBarChart(elId, entries, color) {
  const el = document.getElementById(elId);
  if (!el) return;
  if (!entries || entries.length === 0) { el.innerHTML = '<div class="chart-empty">No data</div>'; return; }

  const W = 600, H = 200, P = 28;
  const max = Math.max(...entries.map(e => e.value), 1);
  const barW = Math.max(20, Math.floor((W - P * 2) / entries.length) - 12);

  let svg = '<svg viewBox="0 0 ' + W + ' ' + H + '"><rect width="' + W + '" height="' + H + '" fill="#0b1220"/>';

  entries.forEach((entry, i) => {
    const x = P + i * ((W - P * 2) / entries.length) + 6;
    const barH = (entry.value / max) * (H - P * 2);
    const y = H - P - barH;
    svg += '<rect x="' + x.toFixed(1) + '" y="' + y.toFixed(1) + '" width="' + barW + '" height="' + barH.toFixed(1) + '" fill="' + color + '" rx="4">';
    svg += '<animate attributeName="height" from="0" to="' + barH.toFixed(1) + '" dur="0.4s" fill="freeze"/>';
    svg += '<animate attributeName="y" from="' + (H - P) + '" to="' + y.toFixed(1) + '" dur="0.4s" fill="freeze"/>';
    svg += '</rect>';
    svg += '<text x="' + (x + barW/2).toFixed(1) + '" y="' + (H - 8) + '" text-anchor="middle" fill="#8ea4c2" font-size="10">' + entry.label + '</text>';
    svg += '<text x="' + (x + barW/2).toFixed(1) + '" y="' + (y - 4).toFixed(1) + '" text-anchor="middle" fill="#d0dced" font-size="10">' + entry.display + '</text>';
  });

  svg += '</svg>';
  el.innerHTML = svg;
}

// ─── Render: Concentration Bars ───────────────────────────────
function renderConcentrationBars(concentration) {
  const el = document.getElementById('concentration-bars');
  if (!el) return;
  if (!concentration || Object.keys(concentration).length === 0) {
    el.innerHTML = '<div class="neutral" style="text-align:center;padding:16px;">No positions</div>';
    return;
  }
  let html = '';
  const entries = Object.entries(concentration).sort((a,b) => b[1] - a[1]);
  for (const [symbol, pct] of entries) {
    const pctNum = Number(pct) * 100;
    const barColor = pctNum > 50 ? 'var(--red)' : pctNum > 30 ? 'var(--orange)' : 'var(--cyan)';
    html += '<div style="display:flex;align-items:center;gap:12px;margin-bottom:8px;">' +
      '<span style="min-width:80px;font-size:13px;font-weight:600;">' + symbol + '</span>' +
      '<div style="flex:1;height:24px;background:var(--bg2);border-radius:6px;overflow:hidden;position:relative;">' +
        '<div style="width:' + pctNum.toFixed(1) + '%;height:100%;background:' + barColor + ';border-radius:6px;transition:width 0.5s;"></div>' +
      '</div>' +
      '<span style="min-width:50px;text-align:right;font-size:13px;font-weight:600;color:' + barColor + '">' + pctNum.toFixed(1) + '%</span>' +
    '</div>';
  }
  el.innerHTML = html;
}

// ─── Render: Guards ──────────────────────────────────────────
function renderGuards(guardNames, blockReasons, containerId) {
  const el = document.getElementById(containerId);
  if (!el) return;
  if (!guardNames || guardNames.length === 0) {
    el.innerHTML = '<div class="neutral">No guards configured</div>';
    return;
  }
  let html = '';
  for (const name of guardNames) {
    const count = (blockReasons && blockReasons[name]) || 0;
    const hasBlocked = count > 0;
    html += '<div class="guard-item">' +
      '<div class="guard-dot ' + (hasBlocked ? 'block' : 'pass') + '"></div>' +
      '<span class="guard-name">' + name + '</span>' +
      (hasBlocked ? '<span class="guard-count badge badge-blocked">' + count + '</span>' : '<span class="guard-count badge badge-ok">pass</span>') +
    '</div>';
  }
  el.innerHTML = html;
}

// ─── Render: Config Display ──────────────────────────────────
function renderConfig() {
  const el = document.getElementById('config-display');
  if (!appState.status) return;

  const sections = [
    { title: 'Server', rows: [
      { key: 'Status', val: appState.status.ready ? 'Running' : 'Not Ready' },
      { key: 'Accounts', val: String(appState.status.account_count || 0) },
    ]},
    { title: 'Risk Management', rows: (() => {
      const r = appState.configRisk || {};
      const rows = [];
      if (r.max_notional) rows.push({ key: 'Max Notional', val: '$' + fmt(r.max_notional, 0) });
      if (r.max_notional_per_symbol) rows.push({ key: 'Max Notional/Symbol', val: '$' + fmt(r.max_notional_per_symbol, 0) });
      if (r.max_concentration_pct) rows.push({ key: 'Max Concentration', val: fmtPct(r.max_concentration_pct / 100) });
      if (r.max_open_positions) rows.push({ key: 'Max Open Positions', val: String(r.max_open_positions) });
      if (r.cooldown_ms) rows.push({ key: 'Cooldown', val: r.cooldown_ms + 'ms' });
      if (r.allowed_symbols) rows.push({ key: 'Allowed Symbols', val: r.allowed_symbols.join(', ') });
      return rows;
    })() },
    { title: 'Execution', rows: (() => {
      const m = appState.metrics || {};
      return [
        { key: 'Total Attempts', val: String(m.execution_attempts || 0) },
        { key: 'Success Rate', val: fmtPct(m.success_rate || 0) },
        { key: 'Blocked Rate', val: fmtPct(m.blocked_rate || 0) },
      ];
    })() },
    { title: 'Scheduler', rows: (() => {
      const s = appState.configScheduler || {};
      return [
        { key: 'Mode', val: s.mode || 'timer' },
        { key: 'Interval', val: (s.interval_ms || 1000) + 'ms' },
        { key: 'Enabled', val: s.enabled ? 'Yes' : 'No' },
      ];
    })() },
  ];

  let html = '';
  for (const sec of sections) {
    html += '<div class="config-section"><h3>' + sec.title + '</h3>';
    for (const row of sec.rows) {
      html += '<div class="config-row"><span class="config-key">' + row.key + '</span><span class="config-val">' + row.val + '</span></div>';
    }
    html += '</div>';
  }
  el.innerHTML = html;
}

// ─── Main Data Fetch ──────────────────────────────────────────
async function refreshDashboard() {
  try {
    const [status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, trends, latestReports, metricsHist, valuationHist, config, wsStatus] = await Promise.all([
      fetchJson('/api/status'),
      fetchJson('/api/portfolio'),
      fetchJson('/api/portfolio-summary'),
      fetchJson('/api/orders'),
      fetchJson('/api/execution-summary'),
      fetchJson('/api/execution-diagnostics'),
      fetchJson('/api/exposure-diagnostics'),
      fetchJson('/api/guard-diagnostics'),
      fetchJson('/api/runtime-reports/trends'),
      fetchJson('/api/runtime-reports/latest'),
      fetchJson('/api/runtime-reports/history?type=metrics-snapshot&limit=20'),
      fetchJson('/api/runtime-reports/history?type=portfolio-valuation&limit=20'),
      fetchJson('/api/config'),
      fetchJson('/api/health/marketdata')
    ]);

    appState = { status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, trends, latestReports, metricsHist, valuationHist,
 configRisk: config.risk, configScheduler: config.scheduler, configStrategy: config.strategy, configMarketData: config.market_data, wsStatus
 };

    // Status indicator
    const wsEl = document.getElementById('ws-status');
    if (wsStatus && wsStatus.connected) {
      wsEl.textContent = 'WS: Connected';
      wsEl.style.color = 'var(--green)';
    } else if (wsStatus && wsStatus.source === 'rest') {
      wsEl.textContent = 'REST: Polling';
      wsEl.style.color = 'var(--blue2)';
    } else {
      wsEl.textContent = 'WS: Disconnected';
      wsEl.style.color = 'var(--red)';
    }

    const tag = document.getElementById('status-tag');
    const dot = document.getElementById('live-dot');
    if (status.ready) {
      tag.textContent = 'LIVE'; tag.className = 'tag';
      dot.className = 'live-dot';
    } else {
      tag.textContent = 'OFFLINE'; tag.className = 'tag offline';
      dot.className = 'live-dot offline';
    }

    const m = execDiag.metrics || {};
    const ps = portfolioSummary || {};
    const ed = exposureDiag || {};
    const gd = guardDiag || {};
    const t = trends || {};

    // ═══ Overview KPIs ═══
    document.getElementById('overview-kpis').innerHTML = [
      kpi('Status', status.ready ? 'Operational' : 'Down', status.ready ? 'green' : 'red'),
      kpi('Market Value', '$' + fmt(ps.total_market_value), 'cyan',
        deltaHtml(t.portfolio_value?.latest, t.portfolio_value?.previous)),
      kpi('Realized PnL', '$' + fmt(ps.total_realized_pnl), ps.total_realized_pnl >= 0 ? 'green' : 'red'),
      kpi('Unrealized PnL', '$' + fmt(ps.total_unrealized_pnl), ps.total_unrealized_pnl >= 0 ? 'green' : 'red'),
      kpi('Siphoned Wealth', '$' + fmt(ps.total_siphoned), 'purple'),
      kpi('Open Positions', String(ps.open_positions || 0), 'blue'),
      kpi('Success Rate', fmtPct(m.success_rate), m.success_rate >= 0.8 ? 'green' : m.success_rate >= 0.5 ? 'orange' : 'red'),
    ].join('');

    // Overview: Positions table
    fillTable('overview-positions', portfolio.positions || [], [
      { get: r => r.symbol },
      { get: r => fmt(r.quantity, 6) },
      { get: r => '$' + fmt(r.average_entry_price) },
      { get: r => r.market_price ? '$' + fmt(r.market_price) : '\u2014' },
      { get: r => r.market_value ? '$' + fmt(r.market_value) : '\u2014' },
      { get: r => (r.unrealized_pnl ? '$' + fmt(r.unrealized_pnl) : '\u2014'), cls: r => pnlClass(r.unrealized_pnl || 0) },
    ]);

    // Strategy Stats with Sharpe
    const stats = await fetchJson('/api/strategy-stats');
    if (stats) {
      document.getElementById('overview-kpis').innerHTML += Object.values(stats).map(s =>
        kpi(s.name + ' Sharpe', fmt(s.sharpe_ratio, 2), s.sharpe_ratio > 1 ? 'green' : 'blue')
      ).join('');
    }

    // Overview: Recent orders table
    const recentOrders = (orders || []).slice(-10).reverse();
    fillTable('overview-orders', recentOrders, [
      { get: r => '<span style="font-family:monospace;font-size:11px;">' + r.id + '</span>' },
      { get: r => r.symbol },
      { get: r => '<span class="badge badge-' + r.side + '">' + r.side + '</span>' },
      { get: r => r.quantity },
      { get: r => r.price ? '$' + fmt(Number(r.price)) : '\u2014' },
    ]);
    document.getElementById('order-count-badge').textContent = String((orders || []).length);

    // Overview: Guards
    renderGuards(gd.active_guards, m.block_reasons, 'overview-guards');

    // ═══ Portfolio Page ═══
    document.getElementById('portfolio-kpis').innerHTML = [
      kpi('Total Value', '$' + fmt(ps.total_market_value), 'cyan'),
      kpi('Realized PnL', '$' + fmt(ps.total_realized_pnl), ps.total_realized_pnl >= 0 ? 'green' : 'red'),
      kpi('Unrealized PnL', '$' + fmt(ps.total_unrealized_pnl), ps.total_unrealized_pnl >= 0 ? 'green' : 'red'),
      kpi('Positions', String(ps.open_positions || 0), 'blue'),
    ].join('');

    fillTable('portfolio-table', portfolio.positions || [], [
      { get: r => '<strong>' + r.symbol + '</strong>' },
      { get: r => fmt(r.quantity, 6) },
      { get: r => '$' + fmt(r.average_entry_price) },
      { get: r => '$' + fmt(r.cost_basis) },
      { get: r => r.market_price ? '$' + fmt(r.market_price) : '\u2014' },
      { get: r => r.market_value ? '$' + fmt(r.market_value) : '\u2014' },
      { get: r => r.unrealized_pnl != null ? '$' + fmt(r.unrealized_pnl) : '\u2014', cls: r => pnlClass(r.unrealized_pnl || 0) },
      { get: r => r.realized_pnl != null ? '$' + fmt(r.realized_pnl) : '\u2014', cls: r => pnlClass(r.realized_pnl || 0) },
      { get: r => ps.concentration ? fmtPct(ps.concentration[r.symbol] || 0) : '\u2014' },
    ]);

    renderConcentrationBars(ps.concentration || ed.concentration);

    // ═══ Orders Page ═══
    document.getElementById('orders-kpis').innerHTML = [
      kpi('Total Orders', String(execSummary.total_orders || 0), 'blue'),
      kpi('Unique Symbols', String(execSummary.unique_symbols || 0), 'purple'),
      kpi('Top Symbol', execSummary.top_symbol || '\u2014', 'cyan'),
      kpi('Success Rate', fmtPct(m.success_rate), m.success_rate >= 0.8 ? 'green' : 'orange'),
    ].join('');

    const allOrders = (orders || []).slice().reverse();
    fillTable('orders-table', allOrders, [
      { get: r => '<span style="font-family:monospace;font-size:11px;">' + r.id + '</span>' },
      { get: r => '<strong>' + r.symbol + '</strong>' },
      { get: r => '<span class="badge badge-' + r.side + '">' + r.side + '</span>' },
      { get: r => r.type || '\u2014' },
      { get: r => r.quantity },
      { get: r => r.price ? '$' + fmt(Number(r.price)) : '\u2014' },
      { get: r => r.status ? '<span class="badge badge-' + (r.status === 'filled' ? 'ok' : 'info') + '">' + r.status + '</span>' : '\u2014' },
    ]);

    document.getElementById('execution-detail').textContent = JSON.stringify(execDiag, null, 2);

    // ═══ Guards Page ═══
    document.getElementById('guards-kpis').innerHTML = [
      kpi('Active Guards', String((gd.active_guards || []).length), 'blue'),
      kpi('Total Blocked', String(m.execution_blocked || 0), 'red'),
      kpi('Block Rate', fmtPct(m.blocked_rate), m.blocked_rate > 0.5 ? 'red' : m.blocked_rate > 0.1 ? 'orange' : 'green'),
      kpi('Top Block Reason', t.latest_dominant_block_reason || '\u2014', 'orange'),
    ].join('');

    renderGuards(gd.active_guards, m.block_reasons, 'guards-detail');

    const blockEntries = Object.entries(m.block_reasons || {}).map(([label, value]) => ({
      label, value: Number(value), display: String(value)
    }));
    renderBarChart('block-reasons-chart', blockEntries, '#ff5252');

    // ═══ Charts Page ═══
    renderLineChart('valuation-chart', valuationHist, r => Number(r.payload?.portfolio_value ?? 0), '#00e676', '$');
    renderLineChart('metrics-chart', metricsHist, r => Number((r.payload?.metrics?.success_rate ?? 0) * 100), '#448aff', '%');
    renderLineChart('concentration-trend-chart', valuationHist,
      r => { const vals = Object.values(r.payload?.concentration || {}); return vals.length ? Math.max(...vals) * 100 : 0; },
      '#b388ff', '%');
    renderLineChart('blocked-trend-chart', metricsHist, r => Number(r.payload?.metrics?.execution_blocked ?? 0), '#ff5252');
    renderLineChart('siphoned-trend-chart', valuationHist, r => Number(r.payload?.total_siphoned ?? 0), '#b388ff', '$');
    renderLineChart('pnl-trend-chart', valuationHist, r => Number(r.payload?.realized_pnl ?? 0), '#00e676', '$');

    const concEntries = Object.entries(ed.concentration || {}).map(([label, value]) => ({
      label, value: Number(value), display: fmtPct(value)
    }));
    renderBarChart('concentration-chart', concEntries, '#b388ff');

    // Charts: History tables
    fillTable('metrics-history-table', metricsHist, [
      { get: r => r.timestamp ? new Date(r.timestamp).toLocaleString() : '\u2014' },
      { get: r => r.payload?.metrics?.execution_attempts },
      { get: r => r.payload?.metrics?.execution_success },
      { get: r => r.payload?.metrics?.execution_blocked },
      { get: r => fmtPct(r.payload?.metrics?.success_rate), cls: r => pnlClass(r.payload?.metrics?.success_rate - 0.5) },
    ]);
    fillTable('valuation-history-table', valuationHist, [
      { get: r => r.timestamp ? new Date(r.timestamp).toLocaleString() : '\u2014' },
      { get: r => '$' + fmt(r.payload?.portfolio_value) },
      { get: r => '$' + fmt(r.payload?.realized_pnl), cls: r => pnlClass(r.payload?.realized_pnl || 0) },
      { get: r => '$' + fmt(r.payload?.unrealized_pnl), cls: r => pnlClass(r.payload?.unrealized_pnl || 0) },
    ]);

    // ═══ Reports Page ═══
    document.getElementById('trend-kpis').innerHTML = [
      kpi('Metrics Samples', String(t.metrics_samples || 0), 'blue'),
      kpi('Valuation Samples', String(t.valuation_samples || 0), 'cyan'),
      kpi('Portfolio Delta', '$' + fmt(t.portfolio_value?.delta), t.portfolio_value?.delta >= 0 ? 'green' : 'red'),
      kpi('Success Rate Delta', fmtPct(t.success_rate?.delta), t.success_rate?.delta >= 0 ? 'green' : 'red'),
    ].join('');

    document.getElementById('trends-detail').textContent = JSON.stringify(trends, null, 2);
    document.getElementById('latest-reports').textContent = JSON.stringify(latestReports, null, 2);

    // ═══ Config Page ═══
    renderConfig();

    // Timestamp
    document.getElementById('last-updated').textContent = 'Updated ' + new Date().toLocaleTimeString();
    document.getElementById('uptime-footer').textContent = 'Last refresh: ' + new Date().toLocaleTimeString();

  } catch (error) {
    document.getElementById('last-updated').textContent = 'Error: ' + error.message;
    document.getElementById('live-dot').className = 'live-dot offline';
    document.getElementById('status-tag').textContent = 'ERROR';
    document.getElementById('status-tag').className = 'tag offline';
  }
}

// ─── Auto-refresh ─────────────────────────────────────────────
function scheduleRefresh() {
  if (refreshTimer) clearInterval(refreshTimer);
  if (document.getElementById('auto-refresh').checked) {
    refreshTimer = setInterval(refreshDashboard, Number(document.getElementById('refresh-interval').value));
  }
}
document.getElementById('refresh-btn').addEventListener('click', refreshDashboard);
document.getElementById('auto-refresh').addEventListener('change', scheduleRefresh);
document.getElementById('refresh-interval').addEventListener('change', scheduleRefresh);

// ─── Initial load ─────────────────────────────────────────────
refreshDashboard();
scheduleRefresh();
</script>
</body>
</html>
`
