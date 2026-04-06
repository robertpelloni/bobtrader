package httpapi

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>UltraTrader Go Dashboard</title>
  <style>
    :root { color-scheme: dark; }
    body { font-family: Arial, sans-serif; margin: 0; background: #0b1220; color: #d7e2f0; }
    header { padding: 20px; border-bottom: 1px solid #243044; background: #111a2d; position: sticky; top: 0; z-index: 1; }
    h1 { margin: 0 0 6px; font-size: 24px; }
    p { margin: 0; color: #9fb0c8; }
    .toolbar { margin-top: 14px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
    button, select { background: #17233a; color: #d7e2f0; border: 1px solid #2f3f5d; border-radius: 8px; padding: 8px 12px; }
    label { color: #c7d2e4; font-size: 14px; }
    main { padding: 20px; display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 16px; }
    section { background: #121c2f; border: 1px solid #243044; border-radius: 10px; padding: 16px; }
    h2 { margin-top: 0; font-size: 18px; }
    pre { white-space: pre-wrap; word-break: break-word; font-size: 12px; color: #d7e2f0; }
    .muted { color: #9fb0c8; }
    .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; }
    .card { background: #0f1729; border: 1px solid #22324e; border-radius: 8px; padding: 12px; }
    .card .label { color: #8ea4c2; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
    .card .value { font-size: 22px; margin-top: 6px; font-weight: bold; }
    table { width: 100%; border-collapse: collapse; font-size: 12px; }
    th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid #243044; }
    th { color: #9fb0c8; }
    .ok { color: #59d089; }
    .warn { color: #f0c36d; }
    .error { color: #f07178; }
  </style>
</head>
<body>
  <header>
    <h1>UltraTrader Go Dashboard</h1>
    <p class="muted">Operator dashboard for the clean-room Go ultra-project.</p>
    <div class="toolbar">
      <button id="refresh-btn">Refresh now</button>
      <label><input type="checkbox" id="auto-refresh" checked> auto refresh</label>
      <label>interval
        <select id="refresh-interval">
          <option value="2000">2s</option>
          <option value="5000" selected>5s</option>
          <option value="10000">10s</option>
        </select>
      </label>
      <span id="last-updated" class="muted">never</span>
    </div>
  </header>
  <main>
    <section style="grid-column: 1 / -1;">
      <h2>Top-Level Summary</h2>
      <div id="cards" class="cards"></div>
    </section>
    <section><h2>Status</h2><pre id="status">loading...</pre></section>
    <section><h2>Execution Diagnostics</h2><pre id="execution-diagnostics">loading...</pre></section>
    <section><h2>Exposure Diagnostics</h2><pre id="exposure-diagnostics">loading...</pre></section>
    <section><h2>Guard Diagnostics</h2><pre id="guards">loading...</pre></section>
    <section><h2>Report Trends</h2><pre id="report-trends">loading...</pre></section>
    <section><h2>Latest Reports</h2><pre id="runtime-reports">loading...</pre></section>
    <section style="grid-column: 1 / -1;"><h2>Metrics History</h2><div id="metrics-history">loading...</div></section>
    <section style="grid-column: 1 / -1;"><h2>Valuation History</h2><div id="valuation-history">loading...</div></section>
  </main>
  <script>
    let refreshTimer = null;

    async function fetchJson(url) {
      const response = await fetch(url);
      if (!response.ok) throw new Error(url + ' => ' + response.status);
      return response.json();
    }

    function setJson(id, data) {
      document.getElementById(id).textContent = JSON.stringify(data, null, 2);
    }

    function card(label, value, cls='') {
      return '<div class="card"><div class="label">' + label + '</div><div class="value ' + cls + '">' + value + '</div></div>';
    }

    function renderHistoryTable(elId, rows, columns) {
      const el = document.getElementById(elId);
      if (!rows || rows.length === 0) {
        el.innerHTML = '<div class="muted">No data</div>';
        return;
      }
      let html = '<table><thead><tr>' + columns.map(c => '<th>' + c.label + '</th>').join('') + '</tr></thead><tbody>';
      for (const row of rows) {
        html += '<tr>' + columns.map(c => '<td>' + (c.get(row) ?? '') + '</td>').join('') + '</tr>';
      }
      html += '</tbody></table>';
      el.innerHTML = html;
    }

    async function refreshDashboard() {
      try {
        const [status, portfolioSummary, executionDiagnostics, exposureDiagnostics, guardDiagnostics, trends, latestReports, metricsHistory, valuationHistory] = await Promise.all([
          fetchJson('/api/status'),
          fetchJson('/api/portfolio-summary'),
          fetchJson('/api/execution-diagnostics'),
          fetchJson('/api/exposure-diagnostics'),
          fetchJson('/api/guard-diagnostics'),
          fetchJson('/api/runtime-reports/trends'),
          fetchJson('/api/runtime-reports/latest'),
          fetchJson('/api/runtime-reports/history?type=metrics-snapshot&limit=10'),
          fetchJson('/api/runtime-reports/history?type=portfolio-valuation&limit=10')
        ]);

        document.getElementById('cards').innerHTML = [
          card('Ready', status.ready ? 'Yes' : 'No', status.ready ? 'ok' : 'error'),
          card('Accounts', status.account_count ?? 0),
          card('Market Value', portfolioSummary.total_market_value ?? 0),
          card('Realized PnL', portfolioSummary.total_realized_pnl ?? 0),
          card('Unrealized PnL', portfolioSummary.total_unrealized_pnl ?? 0),
          card('Open Positions', portfolioSummary.open_positions ?? 0),
          card('Exec Attempts', executionDiagnostics.metrics?.execution_attempts ?? 0),
          card('Exec Success Rate', ((executionDiagnostics.metrics?.success_rate ?? 0) * 100).toFixed(1) + '%'),
          card('Exec Blocked Rate', ((executionDiagnostics.metrics?.blocked_rate ?? 0) * 100).toFixed(1) + '%')
        ].join('');

        setJson('status', status);
        setJson('execution-diagnostics', executionDiagnostics);
        setJson('exposure-diagnostics', exposureDiagnostics);
        setJson('guards', guardDiagnostics);
        setJson('report-trends', trends);
        setJson('runtime-reports', latestReports);

        renderHistoryTable('metrics-history', metricsHistory, [
          { label: 'Timestamp', get: r => r.timestamp },
          { label: 'Attempts', get: r => r.payload?.metrics?.execution_attempts },
          { label: 'Success', get: r => r.payload?.metrics?.execution_success },
          { label: 'Blocked', get: r => r.payload?.metrics?.execution_blocked },
          { label: 'Success Rate', get: r => ((r.payload?.metrics?.success_rate ?? 0) * 100).toFixed(1) + '%' }
        ]);

        renderHistoryTable('valuation-history', valuationHistory, [
          { label: 'Timestamp', get: r => r.timestamp },
          { label: 'Portfolio Value', get: r => r.payload?.portfolio_value },
          { label: 'Realized PnL', get: r => r.payload?.realized_pnl },
          { label: 'Unrealized PnL', get: r => r.payload?.unrealized_pnl }
        ]);

        document.getElementById('last-updated').textContent = 'last updated ' + new Date().toLocaleTimeString();
      } catch (error) {
        document.getElementById('last-updated').textContent = 'refresh failed: ' + error;
      }
    }

    function scheduleRefresh() {
      if (refreshTimer) clearInterval(refreshTimer);
      if (document.getElementById('auto-refresh').checked) {
        refreshTimer = setInterval(refreshDashboard, Number(document.getElementById('refresh-interval').value));
      }
    }

    document.getElementById('refresh-btn').addEventListener('click', refreshDashboard);
    document.getElementById('auto-refresh').addEventListener('change', scheduleRefresh);
    document.getElementById('refresh-interval').addEventListener('change', scheduleRefresh);

    refreshDashboard();
    scheduleRefresh();
  </script>
</body>
</html>
`
