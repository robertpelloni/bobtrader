# Go Ultra-Project Requirements

## Purpose
Define the requirements for consolidating the current PowerTrader AI project and the 50 imported crypto-trading submodules into a single long-horizon Go platform.

## Problem Statement
The repository now contains:
- the existing PowerTrader AI Python application,
- 50 external crypto-trading submodules spanning Go, TypeScript, Python, Java, C#, Rust, and C++,
- overlapping capabilities across execution, backtesting, research, UI, notifications, arbitrage, market data, and AI-assisted trading.

The current ecosystem is valuable but fragmented. The goal is to create a unified Go system that preserves the strongest ideas and capabilities while removing duplicated architecture, runtime sprawl, inconsistent quality, and integration debt.

## Primary Goal
Build a single, modular, production-oriented Go trading platform that:
1. uses a coherent core architecture,
2. clean-room reimplements the best ideas from the audited submodules,
3. can eventually subsume the existing Python application,
4. maintains strong documentation, testability, and operational safety.

## Stage-1 Requirement
Before any broad porting starts, the project must complete a structured audit and target-architecture definition.

This phase must answer:
- Which submodule has the best architecture?
- Which submodule is the best written system overall?
- Which projects are reusable only as references?
- Which projects are blocked by license, maintenance, or quality risk?
- What should the Go target architecture be?
- In what order should features be assimilated?

## Success Criteria
A valid Stage-1 outcome must produce:
- a full submodule inventory,
- a comparative architecture audit,
- a legal/licensing risk analysis,
- a target Go architecture,
- a phased migration plan,
- clear recommendation for the base kernel and the architectural reference system.

## Functional Requirements for the Target Go Platform
The eventual Go ultra-project should support:

### Core Trading
- multi-exchange spot trading,
- optional derivatives support behind clear capability flags,
- unified order, position, balance, and portfolio models,
- real-time market data via websocket and REST fallbacks,
- account/session separation,
- high-precision financial arithmetic.

### Strategy System
- built-in strategies,
- pluggable strategy engine,
- reusable indicators,
- multi-timeframe signal composition,
- event-driven strategy execution,
- paper trading and simulation modes.

### Research and Analytics
- historical data ingestion,
- backtesting,
- optimization,
- trade journaling,
- equity curve and PnL reporting,
- market/asset screening,
- technical, statistical, and optional AI-assisted analysis.

### Risk and Controls
- guard pipeline before execution,
- max position sizing,
- cooldown rules,
- symbol whitelists/blacklists,
- exposure/concentration limits,
- per-strategy and per-account controls,
- auditable decision path.

### Arbitrage and Advanced Modes
- cross-exchange arbitrage,
- triangular arbitrage,
- optional statistical arbitrage,
- optional lending/yield/funding modules,
- modular support for DeFi or prediction-market adapters.

### Operator Experience
- CLI,
- web dashboard,
- notifications,
- configuration management,
- structured logs,
- event log / append-only audit trail,
- clean deployment story for local, server, and container environments.

## Non-Functional Requirements
- Go-first architecture,
- composable modules,
- deterministic tests where feasible,
- explicit interfaces between subsystems,
- low operational complexity,
- no mandatory heavyweight database for core local usage,
- production-safe logging and error propagation,
- supportable licensing posture.

## Constraints

### Licensing Constraint
The target system must not blindly port or merge code from incompatible or unclear licenses.

Observed risks in imported submodules include:
- AGPL-3.0,
- GPL,
- CC BY-NC 4.0,
- Shareware,
- no license present.

Therefore the target system must prefer:
- clean-room reimplementation,
- architecture inspiration rather than direct code copying,
- permissive-source borrowing only where legally appropriate.

### Scope Constraint
Porting all 50 projects feature-for-feature in a single iteration is not realistic.
The effort must be phased.

### Quality Constraint
The project must prefer:
- architecture quality,
- maintainability,
- testability,
- operational coherence,

over naive feature count.

## Key Decisions from Stage-1 Audit
- **Best architecture:** `TraderAlice/OpenAlice`
- **Best written / best target kernel:** `c9s/bbgo`
- **Best exchange abstraction reference:** `ccxt/ccxt`
- **Richest feature monolith to mine for ideas:** `Ekliptor/WolfBot`

## Recommended Delivery Strategy
1. Audit and document.
2. Design target Go architecture.
3. Scaffold a new Go project around the chosen kernel approach.
4. Assimilate features subsystem-by-subsystem.
5. Use parity tests and acceptance criteria for each assimilation wave.

## Out of Scope for Stage-1
- complete feature port of all submodules,
- wholesale code transplantation,
- replacing the existing PowerTrader Python system immediately,
- irreversible removal of source submodules.
