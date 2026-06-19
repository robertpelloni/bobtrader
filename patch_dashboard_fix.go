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

	// Fix the fetch code, making sure the old code matches
	oldFetch := `    const [status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, config, trends, latestReports, metricsHist, valuationHist] = await Promise.all([
      fetchAPI('/api/status'),
      fetchAPI('/api/portfolio'),
      fetchAPI('/api/portfolio-summary'),
      fetchAPI('/api/orders'),
      fetchAPI('/api/execution-summary'),
      fetchAPI('/api/execution-diagnostics'),
      fetchAPI('/api/exposure-diagnostics'),
      fetchAPI('/api/guard-diagnostics'),
      fetchAPI('/api/config'),
      fetchAPI('/api/runtime-reports/trends'),
      fetchAPI('/api/runtime-reports/latest'),
      fetchAPI('/api/runtime-reports/history?type=metrics_snapshot&limit=20'),
      fetchAPI('/api/runtime-reports/history?type=portfolio_snapshot&limit=20')
    ]);`

	if strings.Contains(content, oldFetch) {
		fmt.Println("Found oldFetch in dashboard.go")
	}

	newFetch := `    const [status, portfolio, portfolioSummary, orders, execSummary, execDiag, exposureDiag, guardDiag, config, trends, latestReports, metricsHist, valuationHist, wsHealth] = await Promise.all([
      fetchAPI('/api/status'),
      fetchAPI('/api/portfolio'),
      fetchAPI('/api/portfolio-summary'),
      fetchAPI('/api/orders'),
      fetchAPI('/api/execution-summary'),
      fetchAPI('/api/execution-diagnostics'),
      fetchAPI('/api/exposure-diagnostics'),
      fetchAPI('/api/guard-diagnostics'),
      fetchAPI('/api/config'),
      fetchAPI('/api/runtime-reports/trends'),
      fetchAPI('/api/runtime-reports/latest'),
      fetchAPI('/api/runtime-reports/history?type=metrics_snapshot&limit=20'),
      fetchAPI('/api/runtime-reports/history?type=portfolio_snapshot&limit=20'),
      fetchAPI('/api/ws-health')
    ]);`
	content = strings.Replace(content, oldFetch, newFetch, 1)

	// Since we likely already replaced it but forgot something, let's just make sure wsHealth is fetched
	if !strings.Contains(content, "wsHealth") {
		fmt.Println("Warning: wsHealth string is missing entirely.")
	}

	err = ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Rewrite dashboard done.")
}
