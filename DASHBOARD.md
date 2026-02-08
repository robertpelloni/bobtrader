# PowerTrader AI - Dashboard & Submodules

**Version:** 2.1.0
**Last Updated:** 2026-01-18

This document tracks all active submodules, their versions, and locations within the project structure.

---

## Project Structure Overview

```
PowerTrader_AI/
├── hub_data/                  # Runtime data (databases, logs, config)
├── powertrader-ts/            # TypeScript Web Architecture (Active Development)
│   ├── backend/               # Node.js + Express + TypeScript
│   │   ├── src/
│   │   │   ├── modules/       # Submodules & Adapters
│   │   │   │   ├── cointrade/ # External Strategy Engine
│   └── frontend/              # React + Vite
```

---

## Submodule Status Dashboard

| Submodule | Version | Location | Status | Build |
| :--- | :--- | :--- | :--- | :--- |
| **cointrade** | `0.0.1-alpha` | `powertrader-ts/backend/src/modules/cointrade` | **Placeholder** | N/A |
| **tulind** | `0.8.0` | `node_modules` (Backend) | **Integrated** | Stable |
| **genetic-js** | `0.1.14` | `node_modules` (Backend) | **Integrated** | Stable |

### Integration Details

#### Cointrade
*   **Repository:** `github.com/mnmballa2323/cointrade`
*   **Integration Type:** Adapter Pattern (`CointradeAdapter.ts`)
*   **Current State:** The folder exists, but external code fetch is pending access. The adapter mocks the signal interface to allow the rest of the system to function.

---

## Build Information

*   **Backend Build:** `npm run build` (Outputs to `dist/`)
*   **Frontend Build:** `npm run build` (Outputs to `dist/`)
*   **Docker Build:** `docker-compose build`

---

**Note:** This dashboard should be updated automatically by CI/CD pipelines in the future.
