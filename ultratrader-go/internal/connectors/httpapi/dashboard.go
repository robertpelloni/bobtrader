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
    header { padding: 20px; border-bottom: 1px solid #243044; background: #111a2d; }
    h1 { margin: 0 0 6px; font-size: 24px; }
    p { margin: 0; color: #9fb0c8; }
    main { padding: 20px; display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; }
    section { background: #121c2f; border: 1px solid #243044; border-radius: 10px; padding: 16px; }
    h2 { margin-top: 0; font-size: 18px; }
    pre { white-space: pre-wrap; word-break: break-word; font-size: 12px; color: #d7e2f0; }
    .muted { color: #9fb0c8; }
  </style>
</head>
<body>
  <header>
    <h1>UltraTrader Go Dashboard</h1>
    <p class="muted">Bootstrap operator dashboard for the clean-room Go ultra-project.</p>
  </header>
  <main>
    <section><h2>Status</h2><pre id="status">loading...</pre></section>
    <section><h2>Portfolio Summary</h2><pre id="portfolio-summary">loading...</pre></section>
    <section><h2>Execution Diagnostics</h2><pre id="execution-diagnostics">loading...</pre></section>
    <section><h2>Exposure Diagnostics</h2><pre id="exposure-diagnostics">loading...</pre></section>
    <section><h2>Metrics</h2><pre id="metrics">loading...</pre></section>
    <section><h2>Guards</h2><pre id="guards">loading...</pre></section>
    <section><h2>Report Trends</h2><pre id="report-trends">loading...</pre></section>
    <section><h2>Latest Reports</h2><pre id="runtime-reports">loading...</pre></section>
  </main>
  <script>
    async function load(id, url) {
      const el = document.getElementById(id);
      try {
        const response = await fetch(url);
        const data = await response.json();
        el.textContent = JSON.stringify(data, null, 2);
      } catch (error) {
        el.textContent = 'error: ' + error;
      }
    }
    load('status', '/api/status');
    load('portfolio-summary', '/api/portfolio-summary');
    load('execution-diagnostics', '/api/execution-diagnostics');
    load('exposure-diagnostics', '/api/exposure-diagnostics');
    load('metrics', '/api/metrics');
    load('guards', '/api/guard-diagnostics');
    load('report-trends', '/api/runtime-reports/trends');
    load('runtime-reports', '/api/runtime-reports/latest');
  </script>
</body>
</html>
`
