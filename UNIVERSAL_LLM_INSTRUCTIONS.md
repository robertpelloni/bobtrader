# Universal LLM Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Universal instructions applicable to all LLM agents working on PowerTrader AI and its evolving Go ultra-project

---

## Project Overview

**PowerTrader AI** is now a dual-track project consisting of:
- the legacy Python/Tkinter PowerTrader AI runtime, and
- the emerging clean-room Go ultra-project under `ultratrader-go/`.

The legacy system is a fully automated crypto trading bot powered by:
- Custom kNN-based price prediction AI
- Structured/tiered DCA (Dollar-Cost Averaging) system
- Multi-exchange price aggregation
- Persistent trade analytics
- Multi-platform notification system
- Volume-based analysis

The Go ultra-project is intended to assimilate the strongest architecture and subsystem ideas from PowerTrader AI and the imported submodule corpus into a modular, observable, daemon-grade Go platform.

**Core Philosophy:**
- Long-term spot trading (no futures, no leverage)
- No stop-loss by design (belief in HODLing quality coins)
- Patient, disciplined approach to crypto markets
- Risk management through DCA tiers and position limits
- Data-driven decision making via neural patterns

**Architecture:**
- Modular Python system with Tkinter GUI
- Core system: pt_hub.py, pt_thinker.py, pt_trader.py, pt_trainer.py, pt_backtester.py
- Analytics: pt_analytics.py, pt_analytics_dashboard.py
- Exchanges: pt_exchanges.py, pt_thinker_exchanges.py
- Notifications: pt_notifications.py
- Volume Analysis: pt_volume.py

---

## Universal Development Guidelines

### 1. Code Quality Standards

**General Principles:**
- Write clean, readable, self-documenting code
- Use descriptive variable names (no single letters except loop variables)
- Functions should do ONE thing and do it well
- Avoid premature optimization - clarity first
- Type hints are encouraged for public interfaces
- Docstrings for public APIs only (keep code self-explanatory)

**Error Handling:**
- Never swallow exceptions silently
- Log errors with context (what, where, why)
- Provide meaningful error messages
- Use specific exception types, not generic Exception
- Graceful degradation preferred over hard failures
- In trading systems: prioritize safety over error recovery

**Comments and Documentation:**
- Comment WHY, not WHAT (code shows what)
- Comment complex algorithms and business logic
- Comment non-obvious trade-offs
- Comment workarounds for third-party limitations
- Avoid verbose docstrings for simple methods
- Write self-documenting code over excessive comments

### 2. Testing Requirements

**When Adding New Features:**
- Write unit tests for business logic
- Write integration tests for external API calls
- Test error conditions (network failures, API errors)
- Test edge cases (empty data, null values)
- Test with realistic historical data
- Ensure tests pass before committing

**Regression Prevention:**
- Run full test suite before pushing
- Test with actual exchange APIs when possible
- Validate data integrity (no nulls, no NaNs)
- Check for memory leaks in long-running processes
- Verify database integrity after operations

### 3. Version Management

**Single Source of Truth:**
- VERSION.md contains the project version number
- All references to version must read from VERSION.md
- Never hardcode version numbers in multiple places
- Update VERSION.md before committing changes
- Include version bump in commit message
- Update CHANGELOG.md with each release

**Semantic Versioning:**
- MAJOR: Breaking changes, architecture overhauls
- MINOR: New features, non-breaking changes
- PATCH: Bug fixes, small improvements

### 4. Git Workflow

**Commit Guidelines:**
- Atomic commits (one logical change per commit)
- Clear, descriptive commit messages
- Reference issue/PR numbers when applicable
- Include version bump in message if applicable
- Don't commit generated files (__pycache__, .pyc, .DS_Store)
- Don't commit sensitive data (API keys, secrets)

**Branching Strategy:**
- main: Production-ready code only
- feature/*: Feature development
- bugfix/*: Bug fixes
- docs/*: Documentation updates

### 5. Architecture Principles

**Modularity:**
- Each module should be independently testable
- Minimize coupling between modules
- Use dependency injection where appropriate
- Define clear interfaces between components

**Separation of Concerns:**
- Data layer (database, file I/O)
- Business logic (trading rules, prediction)
- Presentation layer (GUI, CLI)
- External integrations (API clients, exchanges)

### 6. Trading System Specific Guidelines

**Safety First:**
- Never execute trades without confirmation in production
- Validate all inputs before use
- Check account balances before placing orders
- Implement position limits (max per coin, max per day)
- Protect against edge cases (divide by zero, empty data)

**Data Integrity:**
- Never modify historical data
- Always validate price data ranges (no negative prices, no zeros)
- Timestamp consistency (always UTC, always epoch or ISO format)
- Handle timezone conversions explicitly
- Preserve precision for financial calculations

**API Interaction:**
- Respect rate limits
- Implement exponential backoff for retries
- Cache responses when appropriate
- Handle pagination properly
- Validate API responses before use

### 7. Database Guidelines

**Analytics Database (pt_analytics.py):**
- SQLite for simplicity and portability
- Transactions for multi-statement operations
- Index foreign keys for performance
- Periodic cleanup of old data (configurable)
- Backup strategy before schema changes

**Trade Journal:**
- Every trade must be logged
- Trade group IDs link entries/DCAs/exits
- No deletion of trade records (append-only)
- Calculated fields stored for performance

### 8. GUI Guidelines (pt_hub.py)

**Tkinter Best Practices:**
- Avoid blocking the main thread
- Use after() for deferred operations
- Don't update UI from background threads directly
- Use ttk widgets for native look
- Handle window closing gracefully

**User Experience:**
- Clear visual feedback for all actions
- Loading states for long operations
- Error messages with actionable next steps
- Keyboard shortcuts for common actions
- Responsive design (min/max window sizes)

### 9. Integration Guidelines

**Exchange Integration:**
- Unified interface for multiple exchanges
- Fallback chains for reliability
- Price aggregation strategies (median, VWAP, mean)
- Handle API version changes gracefully
- Mock exchange responses for testing

**Notification Integration:**
- Non-blocking notifications (async)
- Rate limiting per platform
- Graceful degradation if platform unavailable
- Notification levels (INFO, WARNING, ERROR, CRITICAL)
- Consolidate similar notifications to avoid spam

### 10. Documentation Requirements

**Code Documentation:**
- Public classes: docstrings with Args/Returns
- Complex algorithms: inline comments explaining logic
- Configuration files: inline comments for each setting
- External integrations: reference official docs

**Project Documentation:**
- README.md: Setup and usage
- CHANGELOG.md: Version history with descriptions
- ROADMAP.md: Current status and future plans
- MODULE_INDEX.md: Inventory of all modules
- Integration guides: Step-by-step for each system

### 11. Security Considerations

**API Keys and Secrets:**
- Never commit API keys or secrets
- Use environment variables or config files
- Separate config for dev/test/production
- Document required secrets without exposing them
- Implement key rotation strategy

**Input Validation:**
- Validate all user inputs
- Sanitize file paths (prevent directory traversal)
- Validate numeric ranges (percentages, prices, quantities)
- Reject obviously malicious inputs

### 12. Performance Optimization

**When to Optimize:**
- Profile first, optimize second
- Focus on hot paths (prediction loop, trade checks)
- Cache expensive operations (historical data, API calls)
- Use efficient data structures (sets, dicts)
- Avoid premature optimization

**Common Performance Patterns:**
- Batch database operations
- Use generators for large datasets
- Lazy loading of resources
- Async I/O where beneficial
- In-memory caching with expiration

### 13. Error Recovery Strategies

**Network Failures:**
- Exponential backoff for retries
- Circuit breaker pattern for repeated failures
- Graceful degradation when APIs unavailable
- Clear user communication of network issues

**Data Corruption:**
- Validate data integrity before use
- Checksums for large files
- Recovery mode for corrupted databases
- Fallback to defaults when config invalid

### 14. Logging Guidelines

**What to Log:**
- Trade executions (with full details)
- API errors (with response bodies)
- System start/shutdown events
- Configuration changes
- Critical failures

**What NOT to Log:**
- API keys or secrets
- PII (if applicable)
- Verbose debug logs in production
- Repetitive heartbeat messages

**Log Levels:**
- DEBUG: Detailed diagnostics (dev only)
- INFO: Normal operations
- WARNING: Non-critical issues
- ERROR: Failures needing attention
- CRITICAL: System-breaking failures

---

## Model-Specific Instructions

Each LLM model has its own instruction file that extends these universal guidelines:

- **CLAUDE.md**: Anthropic Claude-specific instructions
- **GEMINI.md**: Google Gemini-specific instructions
- **GPT.md**: OpenAI GPT-specific instructions
- **copilot-instructions.md**: GitHub Copilot-specific instructions

Always read the universal instructions first, then apply model-specific guidelines.

---

## Task Execution Protocol

When given a task to work on PowerTrader AI:

1. **Analyze Request:**
   - Understand the goal
   - Identify affected modules
   - Consider impact on existing features
   - Assess complexity and risks

2. **Plan Approach:**
   - Outline steps clearly
   - Create TODO list if multi-step task
   - Identify dependencies and blocking issues
   - Estimate time and effort

3. **Implement Incrementally:**
   - Work in small, testable chunks
   - Commit frequently with clear messages
   - Update version and CHANGELOG appropriately
   - Test after each significant change

4. **Validate:**
   - Run unit tests
   - Test with real data when possible
   - Check for regressions
   - Verify documentation is updated

5. **Document:**
   - Update relevant documentation
   - Add examples for new features
   - Update CHANGELOG with changes
   - Bump version if appropriate

---

## Common Pitfalls to Avoid

1. **Breaking Existing Functionality**
   - Always test existing features after changes
   - Maintain backward compatibility when possible
   - Document breaking changes clearly

2. **Ignoring Error Cases**
   - Network failures are common
   - API rate limits happen
   - Data may be missing or malformed
   - Always handle error cases gracefully

3. **Premature Optimization**
   - Clear code > fast code initially
   - Profile before optimizing
   - Measure impact of optimizations

4. **Hardcoding Values**
   - Use configuration files
   - Environment-specific settings
   - VERSION.md for version numbers

5. **Silent Failures**
   - Always log errors
   - Inform user of issues
   - Provide actionable error messages

6. **Over-Engineering**
   - Solve the actual problem, not theoretical ones
   - Keep it simple unless complexity is justified
   - Prefer existing solutions over reinventing

---

## Decision Framework

When faced with a design decision:

**Trade-off Analysis:**
1. **Correctness vs Performance**: Correctness first
2. **Complexity vs Power**: Simpler is usually better
3. **Immediate vs Future**: Solve current problem, design for extensibility
4. **Speed vs Reliability**: Reliability first in trading systems
5. **Features vs Stability**: Stability first, then add features

**When to Ask for Clarification:**
- Multiple valid interpretations exist
- Requirements are ambiguous
- Impact of decision is high
- Design seems to conflict with existing architecture

---

## Project-Specific Knowledge

### Trading Strategy Details
- **Entry Signal**: LONG >= 3 AND SHORT == 0
- **DCA Logic**: Neural level OR drawdown % (whichever first)
- **Max DCAs**: 2 per 24-hour rolling window
- **Exit Logic**: Trailing profit margin (5% no DCA, 2.5% with DCA)
- **Trailing Gap**: 0.5%
- **Neural Levels**: Multi-timeframe predictions (1hr to 1wk)
- **Execution**: Robinhood Crypto API (spot only)

### System Components
- **pt_hub.py**: Main GUI, orchestrates everything
- **pt_thinker.py**: kNN predictor, generates neural levels
- **pt_trader.py**: Trade execution via Robinhood
- **pt_trainer.py**: Trains AI on historical data
- **pt_analytics.py**: SQLite trade journal, performance tracking
- **pt_notifications.py**: Email/Discord/Telegram notifications
- **pt_volume.py**: Volume-based metrics
- **pt_exchanges.py**: Multi-exchange price aggregation

### File Structure
```
PowerTrader_AI/
├── pt_hub.py                    # Main GUI (5835 lines)
├── pt_thinker.py                # Prediction AI (1381 lines)
├── pt_trader.py                 # Trade execution (2421 lines)
├── pt_trainer.py                # AI training (1625 lines)
├── pt_backtester.py             # Backtesting (876 lines)
├── pt_analytics.py              # Analytics (770 lines)
├── pt_analytics_dashboard.py     # Dashboard (252 lines)
├── pt_thinker_exchanges.py      # Exchange wrapper (100 lines)
├── pt_exchanges.py             # Exchange manager (663 lines)
├── pt_notifications.py          # Notifications (876 lines)
├── pt_volume.py               # Volume analysis (128 lines)
├── VERSION.md                  # Version number
├── CHANGELOG.md               # Version history
├── ROADMAP.md                 # Feature planning
├── MODULE_INDEX.md            # Module inventory
└── requirements.txt           # Dependencies
```

### External Dependencies
- **robin_stocks**: Robinhood Crypto API
- **kucoin-python**: KuCoin API (primary price source)
- **python-binance**: Binance API (price fallback)
- **coinbase-advanced-trade-python**: Coinbase API (price fallback)
- **yagmail**: Gmail notifications
- **discord-webhook**: Discord notifications
- **python-telegram-bot**: Telegram notifications
- **pandas, numpy**: Data manipulation
- **matplotlib**: Charting
- **tkinter**: GUI framework (built-in)

---

## Continuous Improvement

**Review this document regularly:**
- Update with new patterns discovered
- Remove outdated practices
- Add lessons learned from issues
- Refine guidelines based on team feedback

**Feedback Loops:**
- Post-mortems for incidents
- Code review discussions
- User feedback integration
- Performance monitoring results

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**License:** Apache 2.0
