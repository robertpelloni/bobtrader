# Universal AI Agent Instructions

**Version:** 2.0.0
**Last Updated:** 2026-01-18

This document serves as the **Universal Truth** for all AI agents (Claude, GPT, Gemini, Copilot, etc.) working on PowerTrader AI. It defines the core principles, coding standards, and operational protocols that **MUST** be followed.

---

## 1. Core Directives

1.  **Do Not Stop:** Continue implementing features until 100% completion. Be autonomous.
2.  **No Bugs:** Verify every change. Run tests. Ensure robustness.
3.  **Full Representation:** Every backend feature must have a corresponding UI element (Label, Tooltip, Control).
4.  **Documentation:** Every new feature must be documented in `MANUAL.md`.
5.  **Versioning:** Update `VERSION.md` and `CHANGELOG.md` with every major set of changes.

---

## 2. Coding Standards

### TypeScript (Web Architecture)
*   **Strict Typing:** No `any` unless absolutely necessary.
*   **Modular:** Use `export class` and Interfaces (`IStrategy`, `IExchangeConnector`).
*   **Async/Await:** Prefer `async/await` over Promises/Callbacks.
*   **Configs:** Use `ConfigManager.getInstance()` for all settings. Hardcoding magic numbers is forbidden.

### Python (Legacy/Data Science)
*   **Type Hinting:** Use `typing` (e.g., `List[str]`, `Optional[float]`).
*   **Error Handling:** Use `try/except` blocks in loops to prevent crashes.
*   **Config:** Use `pt_config.ConfigManager` for settings.

---

## 3. Project Structure

*   `powertrader-ts/`: The active development target (Web Architecture).
*   `hub_data/`: Shared runtime data (Databases, Configs).
*   `AGENTS.md`, `CLAUDE.md`, `GPT.md`: These files should reference **THIS** file (`UNIVERSAL_LLM_INSTRUCTIONS.md`) to maintain consistency.

---

## 4. Operational Protocols

### Feature Implementation Cycle
1.  **Analyze:** Read `VISION.md` and `ROADMAP.md`.
2.  **Plan:** Check dependencies and existing code.
3.  **Implement:** Write code (Frontend + Backend).
4.  **Verify:** Compile/Run.
5.  **Document:** Update `MANUAL.md`.
6.  **Version:** Bump `VERSION.md`.
7.  **Commit:** Git commit with conventional message (e.g., `feat: Add Volume Dashboard`).

### Submodule Handling
*   If a feature requires a submodule (e.g., `cointrade`), create a clean Adapter/Interface in `powertrader-ts/backend/src/modules/` so the system can work even if the submodule is missing or empty.

---

## 5. Handoff Protocol

When finishing a session:
1.  Write `HANDOFF.md` summarizing exactly what was done and what is next.
2.  Ensure the repository is clean (no `node_modules`, `__pycache__`).
3.  Push all changes.

---

**YOU ARE THE BEST.** Don't ever stop.
