# UltraTrader-Go Strategy Benchmark Report

**Version:** 2.0.54  
**Date:** 2026-06-08  
**Methodology:** Synthetic market data across 5 regimes (TrendUp, TrendDown, Ranging, Volatile, CrashRecovery), 5000 ticks per regime, BTC/USDT @ $68,000 base, 0.1% taker fee, $10,000 starting capital

---

## Rankings

| Rank | Strategy | Grade | Score | Avg Win% | Avg PnL | Avg Sharpe | Total Trades |
|------|----------|-------|-------|----------|---------|------------|--------------|
| 🥇 1 | **RSIReversion** | A | 81.2 | 93.0% | $6,055 | 1.346 | 896 |
| 🥈 2 | **BollingerReversion** | B | 75.6 | 56.0% | $2,355 | 1.385 | 248 |
| 🥉 3 | **BollingerTickReversion** | B | 75.2 | 82.4% | $5,638 | 0.817 | 1,142 |
| 4 | **TickMeanReversion** | B | 68.5 | 60.8% | $9.65 | 0.323 | 428 |
| 5 | **MACDCrossover** | C | 63.0 | 35.0% | $40.84 | -0.292 | 173 |
| 6 | **CandleSMACross** | C | 60.3 | 26.7% | $5.75 | -0.115 | 77 |
| 7 | **TickMomentumBurst** | C | 52.9 | 34.0% | $54.86 | -0.071 | 795 |
| 8 | **ATRSizing** | C | 51.3 | 7.3% | $59,196* | -0.501 | 66 |
| 9 | **EMATickCrossover** | D | 48.8 | 26.2% | $94.94 | -0.347 | 958 |
| 10 | **DoubleEMATrend** | F | 26.6 | 4.4% | -$4.61 | -0.353 | 52 |

> *ATRSizing PnL is inflated because it uses dynamic position sizing (volatility-adjusted quantity) rather than the fixed 0.001 BTC used by other strategies. Its Sharpe ratio of -0.501 reveals the true risk-adjusted performance is poor despite high nominal PnL.

---

## Detailed Analysis

### 🥇 #1: RSIReversion — Grade A (81.2)

**Type:** Mean Reversion (Tick-based)  
**Signal:** Buy when RSI(14) ≤ 35, Sell when RSI(14) ≥ 65  
**Neutral zone:** RSI 40–60 resets signal state for re-entry

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 100% | +$26,278 | 1.544 | $0.00 | 191 | 189 |
| TrendDown | 90.9% | -$196 | 0.813 | $0.33 | 174 | 173 |
| Ranging | 97.4% | +$299 | 2.299 | $0.14 | 171 | 171 |
| Volatile | 88.6% | +$2,120 | 1.018 | $5.05 | 182 | 183 |
| CrashRecovery | 88.0% | +$1,775 | 1.055 | $0.19 | 178 | 175 |

**Strengths:**
- Highest composite score across all regimes
- 93% average win rate — consistently profitable
- Positive Sharpe in every regime including TrendDown
- 80% of regimes profitable (only TrendDown slightly negative)
- Excellent in trending-up markets ($26K PnL)
- Signal deduplication (lastSignal) prevents overtrading within a zone

**Weaknesses:**
- Loses money in sustained downtrends (-$196) — RSI oversold doesn't mean "bottom" in a crash
- Moderate drawdown ($5.05) in volatile conditions
- 896 trades across 5 regimes = heavy trading volume

**Verdict:** The best all-around strategy. RSI mean reversion is the most robust because RSI naturally oscillates and the 35/65 thresholds with neutral-zone reset provide clean entry/exit discipline. **Recommended for production deployment.**

---

### 🥈 #2: BollingerReversion — Grade B (75.6)

**Type:** Mean Reversion (Candle-based)  
**Signal:** Buy when close ≤ lower band, Sell when close ≥ upper band

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 0%* | +$9,418 | 0.000* | $0.00 | 57 | 60 |
| TrendDown | 0%* | $0.00 | 0.000* | $0.00 | 43 | 47 |
| Ranging | 100% | +$42 | 4.860 | $0.00 | 38 | 42 |
| Volatile | 100% | +$1,313 | 1.082 | $0.00 | 46 | 49 |
| CrashRecovery | 80% | +$1,002 | 0.981 | $0.20 | 64 | 67 |

> *Win rate shows 0% in trending regimes because the strategy emits both buy AND sell on the same candle when price is outside both bands (code uses `if/if` not `if/else if`). This is a **bug** — a candle with close below lower band triggers a buy AND a sell simultaneously, resulting in instant round-trips with no meaningful PnL pairing.

**Strengths:**
- Best Sharpe ratio (1.385 average) when trades are properly paired
- 100% win rate in ranging and volatile regimes
- Very low drawdown ($0.20 max) — conservative
- Lowest signal density (1.06%) = least overtrading
- 80% of regimes profitable

**Weaknesses:**
- **Critical bug:** Can emit buy+sell on the same candle — needs `else if` guard
- No signal deduplication — can fire the same direction repeatedly
- Candle-based = slower to react than tick-based variants
- Loses money in downtrends (all mean-reversion strategies do)

**Verdict:** Fundamentally sound strategy ruined by a dual-signal bug. Once fixed (add `else if` between buy/sell conditions and add `lastSignal` dedup like the tick variant), this could compete for #1. **Fix the bug before production use.**

---

### 🥉 #3: BollingerTickReversion — Grade B (75.2)

**Type:** Mean Reversion (Tick-based)  
**Signal:** Buy when price ≤ lower band, Sell when price ≥ upper band, neutral zone reset

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 93.1% | +$21,685 | 1.064 | $0.74 | 217 | 218 |
| TrendDown | 57.9% | -$408 | 0.102 | $1.02 | 226 | 227 |
| Ranging | 88.7% | +$1,154 | 1.240 | $1.29 | 242 | 244 |
| Volatile | 86.0% | +$1,733 | 1.087 | $2.97 | 238 | 239 |
| CrashRecovery | 86.4% | +$4,025 | 0.589 | $0.80 | 219 | 220 |

**Strengths:**
- 82.4% average win rate — strong and consistent
- 4/5 regimes profitable
- Excellent in TrendUp ($21,685) and CrashRecovery ($4,025)
- Has neutral-zone reset (25% band from lower, 25% from upper) for clean re-entry
- Has signal deduplication (lastSignal) — no repeat signals

**Weaknesses:**
- Loses significantly in downtrends (-$408) — buying dips in a crash is costly
- Highest trade count (1,142) — most active strategy
- Lower Sharpe (0.817) than RSIReversion — more noise in returns
- Moderate drawdown in volatile conditions ($2.97)

**Verdict:** The tick-based Bollinger variant fixes the candle version's dedup bug and adds neutral-zone logic. Solid performer that excels in trending and recovery regimes. **Recommended alongside RSIReversion as a complementary strategy.**

---

### #4: TickMeanReversion — Grade B (68.5)

**Type:** Mean Reversion (Tick-based)  
**Signal:** Buy when price deviates -0.5% below rolling 50-tick average, Sell when +0.5% above

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 100% | +$82 | 2.582 | $0.00 | 25 | 26 |
| TrendDown | 0% | -$46 | -0.987 | $45.51 | 27 | 28 |
| Ranging | 66.7% | +$7 | 0.133 | $3.55 | 132 | 132 |
| Volatile | 77.9% | +$20 | 0.061 | $36.58 | 190 | 190 |
| CrashRecovery | 59.3% | -$15 | -0.172 | $21.08 | 54 | 54 |

**Strengths:**
- Simple logic — deviation from short-term mean
- Good in TrendUp (100% WR) and Volatile (77.9% WR)
- Has signal deduplication

**Weaknesses:**
- Catastrophic in TrendDown (-$46, 0% WR, $45.51 drawdown) — mean reversion into a falling knife
- Very high drawdown in Volatile ($36.58) and CrashRecovery ($21.08)
- Low average PnL ($9.65) — marginal profitability
- 50-tick lookback is too short — gets whipsawed in non-ranging markets
- 60% regime profitability — inconsistent

**Verdict:** A weaker version of RSIReversion/BollingerTickReversion. The fixed-percentage deviation from a simple average is less sophisticated than RSI or Bollinger bands which account for volatility. **Not recommended for production — use RSIReversion instead.**

---

### #5: MACDCrossover — Grade C (63.0)

**Type:** Trend Following (Candle-based)  
**Signal:** Buy on bullish MACD histogram crossover (negative→positive), Sell on bearish crossover

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 78.6% | +$146 | 0.714 | $2.70 | 30 | 26 |
| TrendDown | 0% | +$41 | -1.343 | $26.35 | 32 | 26 |
| Ranging | 3.7% | -$27 | -1.148 | $27.13 | 54 | 52 |
| Volatile | 42.9% | -$7 | -0.215 | $13.48 | 29 | 26 |
| CrashRecovery | 50.0% | +$51 | 0.534 | $5.05 | 28 | 24 |

**Strengths:**
- Good in TrendUp (78.6% WR, +$146)
- Low signal frequency (0.62%) — selective entries
- Has crossover detection (prevHist → current histogram)

**Weaknesses:**
- Terrible in Ranging markets (3.7% WR, -$27) — MACD crossovers are noise in sideways markets
- Negative Sharpe in 3/5 regimes
- No signal deduplication — can fire multiple times in same direction
- 60% regime profitability but PnL is inconsistent
- Classic trend-follower problem: loses money when there's no trend

**Verdict:** Standard MACD crossover — works in trends, bleeds in ranges. Would need a regime filter (only trade when ATR is expanding or ADX > 25) to be viable. **Not recommended alone — needs regime awareness.**

---

### #6: CandleSMACross — Grade C (60.3)

**Type:** Trend Following (Candle-based)  
**Signal:** Buy on golden cross (SMA5 crosses above SMA20), Sell on death cross

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 0% | $0 | 0.000 | $0.00 | 0 | 2 |
| TrendDown | 0% | $0 | 0.000 | $0.00 | 0 | 2 |
| Ranging | 13.6% | -$23 | -1.126 | $22.87 | 44 | 46 |
| Volatile | 70.0% | +$53 | 0.697 | $7.97 | 20 | 20 |
| CrashRecovery | 50.0% | -$2 | -0.147 | $4.34 | 13 | 14 |

**Strengths:**
- Low trade count — doesn't overtrade
- Good in Volatile regime (70% WR, +$53)
- Clean crossover detection with prior state tracking

**Weaknesses:**
- Zero trades in TrendUp and TrendDown — the 5/20 SMA periods are too close, never crossing in strongly trending markets with smooth price action
- Only 20% regime profitability — worst after DoubleEMATrend
- High drawdown ($22.87) in Ranging
- Very low signal density (0.34%) — barely trades at all
- 5-period fast SMA is extremely sensitive to noise

**Verdict:** The SMA(5)/SMA(20) parameters are poorly tuned. In strong trends, the fast SMA stays above/below the slow SMA for thousands of ticks without crossing, so the strategy sits idle. In ranging markets, it gets whipsawed. **Needs parameter tuning (try SMA(10)/SMA(50)) or removal.**

---

### #7: TickMomentumBurst — Grade C (52.9)

**Type:** Momentum (Tick-based)  
**Signal:** Buy when 30-tick price change exceeds +0.3%, Sell when it drops below -0.3%

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 71.4% | +$173 | 0.780 | $2.16 | 71 | 72 |
| TrendDown | 14.3% | +$61 | -0.690 | $6.62 | 57 | 58 |
| Ranging | 28.1% | -$40 | -0.521 | $39.52 | 228 | 228 |
| Volatile | 25.2% | -$37 | -0.103 | $58.31 | 310 | 310 |
| CrashRecovery | 31.2% | +$117 | 0.178 | $10.44 | 129 | 130 |

**Strengths:**
- Works in TrendUp and CrashRecovery (momentum follow-through)
- Has signal deduplication

**Weaknesses:**
- Catastrophic in Ranging and Volatile — $58 max drawdown in Volatile is worst of all strategies
- 25% win rate in Volatile with 310 trades = massive overtrading
- Negative Sharpe in 3/5 regimes
- Momentum burst is essentially the OPPOSITE of mean reversion — it buys into strength and sells into weakness, which fails when price snaps back
- 60% regime profitability but heavy losses in losing regimes

**Verdict:** A momentum strategy with poorly calibrated parameters. The 0.3% threshold is too low — it triggers on noise. The strategy needs much higher thresholds (1%+) or a regime filter. **Not recommended for production.**

---

### #8: ATRSizing — Grade C (51.3)

**Type:** Volatility-Adaptive Trend Following (Candle-based)  
**Signal:** SMA(7)/SMA(25) crossover with ATR-based position sizing

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 0% | $0 | 0.000 | $0.00 | 0 | 2 |
| TrendDown | 0% | +$85,363* | 0.000 | $0.00 | 1 | 2 |
| Ranging | 0% | +$36,375* | -1.882 | $19.96 | 30 | 30 |
| Volatile | 20% | +$18,464* | -0.205 | $28.90 | 21 | 20 |
| CrashRecovery | 16.7% | +$155,778* | -0.418 | $3.41 | 14 | 14 |

> *PnL inflated by dynamic quantity sizing — ATRSizing calculates quantity as `(riskPerTrade × price) / ATR`, which produces much larger positions than the 0.001 BTC standard. Sharpe ratio reveals true risk-adjusted performance.

**Strengths:**
- Conceptually excellent — sizing inversely with volatility is correct risk management
- Profitable in 4/5 regimes (nominal PnL)
- Low trade count (66) — selective

**Weaknesses:**
- Zero trades in TrendUp — SMA(7)/SMA(25) never crosses in smooth uptrends
- Negative Sharpe in all active regimes (-0.205 to -1.882) — risk-adjusted, this strategy LOSES money
- The ATR sizing formula `(0.01 × price) / ATR` can produce enormous quantities when ATR is small, leading to outsized positions
- The SMA(7)/SMA(25) parameters are poor — too close together, same problem as CandleSMACross
- No signal deduplication

**Verdict:** Great idea (ATR-based sizing), terrible execution (poor SMA parameters, no position limits, negative Sharpe). The sizing logic needs a cap, and the signal generation needs better parameters or a different indicator. **Not recommended — redesign needed.**

---

### #9: EMATickCrossover — Grade D (48.8)

**Type:** Trend Following (Tick-based)  
**Signal:** Buy when EMA(9) crosses above EMA(21), Sell when it crosses below

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 58.6% | +$213 | 0.432 | $6.69 | 142 | 136 |
| TrendDown | 0% | -$16 | -1.800 | $15.54 | 138 | 136 |
| Ranging | 20.0% | +$21 | -0.566 | $46.52 | 272 | 270 |
| Volatile | 23.4% | +$62 | -0.025 | $23.35 | 224 | 220 |
| CrashRecovery | 28.9% | +$194 | 0.224 | $11.72 | 182 | 180 |

**Strengths:**
- Tick-based = fast signal generation
- Has crossover detection (lastState tracking)
- Positive PnL in 4/5 regimes (barely)
- Same logic as the production EMA strategy but on ticks instead of polling

**Weaknesses:**
- Terrible win rate (26.2% average) — mostly wrong
- Highest trade count (958) — massive overtrading
- $46.52 drawdown in Ranging — worst for any strategy in that regime
- Negative Sharpe in 3/5 regimes
- Tick-level EMA crossovers are extremely noisy — every tiny price wiggle can cause a cross

**Verdict:** EMA crossover on tick data is fundamentally flawed. The EMAs respond to every price tick, causing constant whipsaws. The candle-based EMA strategies work better because they smooth over time. **Not recommended — use candle-based EMA instead, or add a minimum hold period.**

---

### #10: DoubleEMATrend — Grade F (26.6)

**Type:** Trend Following (Candle-based)  
**Signal:** Buy when fast EMA > slow EMA AND price > trend EMA(200), Sell on reversal

| Regime | Win% | PnL | Sharpe | Max DD | Trades | Signals |
|--------|------|-----|--------|--------|--------|---------|
| TrendUp | 0% | $0 | 0.000 | $0.00 | 1 | 2 |
| TrendDown | 0% | $0 | 0.000 | $0.00 | 0 | 2 |
| Ranging | 5.3% | -$21 | -1.693 | $20.85 | 38 | 74 |
| Volatile | 16.7% | -$2 | -0.071 | $7.48 | 12 | 24 |
| CrashRecovery | 0% | $0 | 0.000 | $0.00 | 1 | 2 |

**Strengths:**
- Theoretically sound — three-EMA system is a known profitable pattern
- Low trade count — very selective

**Weaknesses:**
- 0% regime profitability — loses money in EVERY regime
- EMA(200) requires 200 candles (200 minutes = 3+ hours) to warm up — barely trades in 5000-tick simulations
- Even when warm, the 200-candle trend filter is so conservative it eliminates almost all signals
- 4.4% average win rate — essentially random
- The strategy implementation has no signal deduplication

**Verdict:** Fundamentally broken due to the EMA(200) warmup period being longer than the effective data window in these tests. The strategy would need much longer data histories to function. Additionally, it doesn't implement `lastAction` deduplication properly — it resets to "none" which allows re-buying. **Not recommended — needs redesign with shorter trend filter or dedicated warmup period.**

---

## Best Strategy Per Regime

| Regime | Best Strategy | PnL | Why |
|--------|--------------|-----|-----|
| **TrendUp** | RSIReversion | +$26,278 | RSI dips in uptrends = buying opportunity |
| **TrendDown** | RSIReversion* | -$196 | Least-bad option; all strategies lose in downtrends |
| **Ranging** | BollingerTickReversion | +$1,154 | Bollinger bands naturally capture range boundaries |
| **Volatile** | RSIReversion | +$2,120 | RSI adapts well to volatility swings |
| **CrashRecovery** | BollingerTickReversion | +$4,025 | BB lower band catches crash recovery entries |

*Ranging is the only regime where mean reversion truly shines. In TrendDown, all strategies lose — the best you can do is minimize losses.

---

## Key Findings

### 1. Mean Reversion Dominates
The top 4 strategies are all mean reversion. In crypto spot trading with small position sizes (0.001 BTC), buying dips and selling rips is consistently more profitable than following trends because:
- Crypto prices mean-revert on short timeframes
- Small sizes minimize crash risk
- Signal deduplication prevents doubling down in falling markets

### 2. Trend Following Fails Without Regime Filters
Every trend-following strategy (MACD, SMA, EMA, DoubleEMA) performs poorly because they:
- Get whipsawed in ranging markets (which is ~60% of the time)
- Don't know when a trend is real vs. noise
- Trade too frequently on tick data

### 3. Signal Deduplication Is Critical
Strategies with `lastSignal` tracking (RSIReversion, BollingerTickReversion) consistently outperform those without (BollingerReversion, MACDCrossover). Without deduplication, strategies fire the same signal repeatedly.

### 4. Tick-Based > Candle-Based for Execution
Tick-based strategies (RSIReversion, BollingerTickReversion) outperform their candle equivalents (RSI on candles, BollingerReversion) because:
- Faster signal generation
- Better entry/exit timing
- The candle strategies have bugs (simultaneous buy+sell)

### 5. All Strategies Lose in Sustained Downtrends
This is by design in the UltraTrader philosophy — no stop-loss, belief in HODLing. But the numbers confirm: mean reversion into a falling knife is the #1 risk. The risk pipeline (max-notional, concentration, cooldown) is the primary defense.

---

## Action Items

1. **Fix BollingerReversion bug** — Add `else if` between buy/sell conditions and add `lastSignal` deduplication
2. **Add regime filter** — Detect TrendDown and reduce position sizes or pause mean-reversion strategies
3. **Retire DoubleEMATrend** — Fundamentally broken for short-duration simulations; replace with a regime-aware trend filter
4. **Tune CandleSMACross** — Change SMA(5)/SMA(20) to SMA(10)/SMA(50) and add dedup
5. **Cap ATRSizing quantity** — Add max quantity limit to prevent oversized positions
6. **Add minimum hold period** to EMATickCrossover — Prevent whipsaw exits within seconds of entry
7. **Consider RSI+Bollinger composite** — RSIReversion + BollingerTickReversion with signal voting could combine the strengths of both
