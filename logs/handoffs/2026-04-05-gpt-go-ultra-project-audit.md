# GPT Handoff Archive - 2026-04-05

## Session Summary
This session converted the imported crypto-trading search results into a structured research corpus and documented a concrete direction for a future consolidated Go trading platform.

## Delivered
- 50 git submodules organized by GitHub search page (`page-02` through `page-06`)
- `SUBMODULES.md` manifest
- Stage-1 architecture and migration docs under `docs/ai/`
- inventory JSON for automation-friendly follow-up analysis
- version bump to `2.0.1`
- changelog entry for the research/audit session
- root `HANDOFF.md`

## Core Decision
The best path is **not** a blind mega-merge.
The best path is:
- `c9s/bbgo` as the Go kernel reference,
- `TraderAlice/OpenAlice` as the system-architecture reference,
- clean-room phased reimplementation,
- selective feature assimilation from WolfBot, CCXT, PyCryptoBot, and secondary repos.

## Why This Matters
Without this framing, a direct port/merge effort would likely fail on:
- incompatible architectures,
- language/runtime sprawl,
- license contamination,
- duplicated subsystem design,
- lack of phased delivery.

## Recommended Phase-1 Build Targets
- `cmd/ultratrader`
- `internal/core`
- `internal/persistence/eventlog`
- `internal/trading/account`
- `internal/exchange`
- `internal/risk`

## Risks to Watch
- accidental inclusion of runtime-generated files in commits,
- overreliance on AGPL/GPL sources beyond architecture/behavior reference,
- trying to port UI/AI layers before the trading kernel is stable,
- mixing execution concerns directly into strategy code.

## Next Agent Instructions
- Start by reading the docs listed in `HANDOFF.md`.
- Keep the next commit focused on the new Go scaffold and supporting docs only.
- Avoid touching unrelated modified data files unless explicitly requested.
