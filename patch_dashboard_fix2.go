package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	path := "ultratrader-go/internal/connectors/httpapi/dashboard.go"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	content := string(b)

	oldFetch := `    const [status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, trends, latestReports, metricsHist, valuationHist, config] = await Promise.all([
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
      fetchJson('/api/config')
    ]);`

	newFetch := `    const [status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, trends, latestReports, metricsHist, valuationHist, config, wsHealth] = await Promise.all([
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
      fetchJson('/api/ws-health').catch(e => ({}))
    ]);`
	content = strings.Replace(content, oldFetch, newFetch, 1)

	err = ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Rewrite dashboard done.")
}
