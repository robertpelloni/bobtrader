# UltraTrader-Go Risk Management Guide

The `ultratrader-go` engine implements a strict pipeline of risk management "guards" that intercept order intent before execution. All guards run concurrently or sequentially depending on the pipeline configuration.

## 1. Max Open Position Policy

The Max Open Position guard prevents the bot from opening too many concurrent trades, ensuring you always have reserve capital and are not over-exposed across the entire market.

**Configuration Example:**
```go
import "github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"

policy := risk.MaxPositionsPolicy{
    MaxConcurrent: 5,
}

guard := risk.NewMaxPositionsGuard(portfolioTracker, policy)
pipeline.AddGuard(guard)
```

**How it works:**
If the portfolio already has 5 active positions (assets with `quantity > 0`), any `Buy` intent for a *new* symbol will be blocked with a `GuardError`. Selling an existing position or buying more of an already open position is permitted.

## 2. Max Concentration (Exposure) Policy

The Concentration Guard ensures that no single asset makes up more than a specified percentage of your total portfolio *Market Value*.

**Configuration Example:**
```go
import "github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"

policy := risk.MaxConcentrationPolicy{
    MaxPercentPerSymbol: 0.25, // No single coin can exceed 25% of total portfolio value
}

guard := risk.NewMaxConcentrationGuard(portfolioTracker, marketFeed, policy)
pipeline.AddGuard(guard)
```

**How it works:**
When evaluating a `Buy` order for $1,000 of BTC:
1. The guard calculates current Total Portfolio Value (e.g., $10,000).
2. It calculates the *projected* value of BTC if the order succeeds (e.g., currently holding $2,000 + $1,000 new = $3,000).
3. It checks the projected concentration: $3,000 / $11,000 = 27%.
4. Since 27% > 25%, the order is blocked.

*Note: The calculation utilizes live market prices from the `marketdata.Feed`, falling back to `CostBasis` if live data is unavailable.*

## 3. Duplicate Side Suppression

Prevents runaway algorithmic bugs from spamming the same side of an order book for a given symbol.

**Configuration Example:**
```go
policy := risk.DuplicateSidePolicy{
    CooldownDuration: 5 * time.Minute, // Cannot place two BUYs for the same symbol within 5 minutes
}
```

## 4. Portfolio Rebalancing

While not strictly a pre-execution guard, the Rebalancer manages passive risk by curing drift.

**Configuration Example:**
```go
import "github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"

config := portfolio.RebalanceConfig{
    Enabled:           true,
    TargetAllocations: map[string]float64{"BTC": 50.0, "ETH": 30.0},
    DriftThresholdPct: 5.0,  // Only rebalance if an asset drifts by +/- 5%
    AvoidWashSales:    true, // Prevent selling at a loss if bought within 30 days
}
rebalancer := portfolio.NewRebalancer(config, logger)
```

**Wash Sale Prevention:**
If `AvoidWashSales` is true, the rebalancer scans order history. If an asset has drifted high and needs to be sold, but its current market price is lower than your average entry cost AND it was purchased in the last 30 days, the rebalancer will suppress the sell order to avoid realizing a wash sale.
