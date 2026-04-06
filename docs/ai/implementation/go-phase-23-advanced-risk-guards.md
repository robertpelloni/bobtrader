# Go Phase-23 Advanced Risk Guards

## Summary
This phase strengthens the Go ultra-project's risk layer with more specific protections against runaway symbol exposure and repeated same-side execution.

## Delivered
- added `max-notional-per-symbol` guard
- added `duplicate-side` guard
- extended config to support:
  - `max_notional_per_symbol`
  - `duplicate_side_window_ms`
- wired both guards into the app runtime pipeline

## Architectural significance
This phase matters because the risk layer is becoming more precise. Earlier guards covered:
- total notional,
- duplicate symbol timing,
- open position count,
- concentration.

This phase adds finer-grained protections that better match realistic trading-runtime needs.

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## Recommended next steps
1. Add max-open-position and concentration policy tuning docs/examples.
2. Add richer concentration and block-reason trend reporting.
3. Continue deeper analytics/reporting modules.
4. Continue legacy Python roadmap/module inventory reconciliation.
