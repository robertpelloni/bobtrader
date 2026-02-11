# Universal AI Agent Instructions

**Version:** 3.1.0
**Last Updated:** 2026-02-10

This document serves as the **Universal Truth** for all AI agents (Claude, GPT, Gemini, Copilot, etc.) working on PowerTrader AI. It defines the core principles, coding standards, and operational protocols that **MUST** be followed.

---

## 1. Core Directives

1.  **Do Not Stop:** Continue implementing features until 100% completion. Be autonomous. If you finish a feature, proceed to the next one immediately.
2.  **Autonomous Execution:** Complete a feature, commit/push, and continue development without stopping. Only ask for confirmation if absolutely necessary.
3.  **Comprehensive Analysis:** Before starting, analyze the project history, roadmap, and current state in extreme detail. Identify missing features or unpolished code.
4.  **No Bugs:** Verify every change. Run tests. Ensure robustness.
5.  **Full Representation:** Every backend feature must have a corresponding UI element (Label, Tooltip, Control). No "hidden" functionality.
6.  **Documentation:** Every new feature must be documented in `MANUAL.md`. Update `VERSION.md`, `CHANGELOG.md`, `ROADMAP.md` and `PROJECT_STRUCTURE.md` with every major set of changes.
7.  **Versioning:** Every build/session should increment the version number. Use Semantic Versioning (Major.Minor.Patch).

---

## 2. Coding Standards

### TypeScript (Web Architecture - Primary)
*   **Strict Typing:** No `any` unless absolutely necessary.
*   **Modular:** Use `export class` and Interfaces (`IStrategy`, `IExchangeConnector`).
*   **Async/Await:** Prefer `async/await` over Promises/Callbacks.
*   **Configs:** Use `ConfigManager.getInstance()` for all settings. Hardcoding magic numbers is forbidden.
*   **Tests:** Write Jest unit tests for all core logic (`Trader`, `Thinker`, `Connectors`).

### Python (Legacy/Data Science)
*   **Type Hinting:** Use `typing` (e.g., `List[str]`, `Optional[float]`).
*   **Error Handling:** Use `try/except` blocks in loops to prevent crashes.
*   **Config:** Use `pt_config.ConfigManager` for settings.

---

## 3. Project Structure

*   `powertrader-ts/`: The active development target (Web Architecture).
*   `hub_data/`: Shared runtime data (Databases, Configs, Logs).
*   `AGENTS.md`, `CLAUDE.md`, `GPT.md`: These files simply reference **THIS** file (`UNIVERSAL_LLM_INSTRUCTIONS.md`) to maintain consistency.

---

## 4. Operational Protocols

### Feature Implementation Cycle
1.  **Analyze:** Read `VISION.md` and `ROADMAP.md` in detail.
2.  **Plan:** Check dependencies, existing code, and submodules.
3.  **Implement:** Write code (Frontend + Backend). Ensure full UI representation.
4.  **Verify:** Compile, Run Tests, Verify UI (Playwright/Screenshot).
5.  **Document:** Update `MANUAL.md` with detailed instructions.
6.  **Version:** Bump `VERSION.md` and update `CHANGELOG.md`.
7.  **Commit:** Git commit with conventional message (e.g., `feat: Add Volume Dashboard`).
8.  **Repeat:** Proceed to the next feature immediately.

### Submodule Handling
*   If a feature requires a submodule (e.g., `cointrade`), create a clean Adapter/Interface in `powertrader-ts/backend/src/modules/` so the system can work even if the submodule is missing or empty.
*   Document all submodule references in `PROJECT_STRUCTURE.md`.

---

## 5. Handoff Protocol

When finishing a session:
1.  Write `HANDOFF.md` summarizing exactly what was done, what is next, and any "memory" for the next agent.
2.  Ensure the repository is clean (no `node_modules`, `__pycache__`, build artifacts).
3.  Push all changes.

---

**YOU ARE THE BEST.** Don't ever stop. Keep on goin'. Don't ever quit.
