# Claude Model-Specific Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Instructions for Anthropic Claude working on PowerTrader AI and the Go ultra-project

**Parent Document:** [UNIVERSAL_LLM_INSTRUCTIONS.md](UNIVERSAL_LLM_INSTRUCTIONS.md)

> **Read `UNIVERSAL_LLM_INSTRUCTIONS.md` first.** For current project direction also review `VISION.md`, `ROADMAP.md`, `TODO.md`, `MEMORY.md`, and `HANDOFF.md`.

---

## Claude-Specific Guidelines

### 1. When to Use Claude

**Strengths of Claude:**
- Excellent at understanding complex context and requirements
- Strong reasoning and planning capabilities
- Good at code synthesis from descriptions
- Effective at technical documentation
- Strong at multi-step task coordination

**Best Use Cases for Claude:**
- Complex architectural decisions
- Code refactoring and optimization
- Documentation generation
- Feature planning and roadmap development
- Multi-file coordination and integration
- Debugging complex issues
- API integration design

**Avoid Using Claude For:**
- Simple code syntax fixes (use local tools)
- Single-line typo corrections
- Very repetitive pattern tasks (use scripts)
- Pure fact lookup (use librarian agent)

### 2. Communication Style

**For Claude:**
- Provide comprehensive context upfront
- Be explicit about requirements and constraints
- Explain reasoning when making decisions
- Provide examples when helpful
- Acknowledge uncertainties explicitly

**Claude's Communication:**
- Will ask clarifying questions when uncertain
- Will explain trade-offs and alternatives
- Will provide multiple approaches when applicable
- Will surface potential issues proactively

### 3. Task Execution for Claude

**Start with Context Gathering:**
1. Read UNIVERSAL_LLM_INSTRUCTIONS.md first
2. Read relevant files (pt_hub.py, pt_thinker.py, etc.)
3. Understand current project state
4. Identify specific task requirements

**Then Execute:**
1. Create TODO list for multi-step tasks
2. Work systematically through each TODO item
3. Update progress as you complete items
4. Validate changes with tests
5. Update documentation appropriately

**Commit Pattern:**
- Complete logical units before committing
- Use descriptive commit messages
- Reference related issues or TODOs
- Update version.md and CHANGELOG.md

### 4. PowerTrader AI Specifics for Claude

**When Working on Analytics:**
- Understand SQLite schema in pt_analytics.py
- Know the TradeJournal API surface
- Understand get_dashboard_metrics() data flow
- Consider performance implications of queries

**When Working on Notifications:**
- Review pt_notifications.py class structure
- Understand NotificationManager coordination
- Know rate limiting implementation
- Test with actual API credentials (sandbox if possible)

**When Working on Exchanges:**
- Read pt_exchanges.py ExchangeManager interface
- Understand pt_thinker_exchanges.py wrapper
- Know fallback chain order (KuCoin → Binance → Coinbase)
- Consider API rate limits and error handling

**When Working on Volume Analysis:**
- Review pt_volume.py VolumeAnalyzer class
- Understand calculation methods (SMA, EMA, VWAP)
- Know anomaly detection thresholds
- Consider integration points with pt_thinker.py

**When Working on GUI (pt_hub.py):**
- Understand the dark theme constants
- Know the WrapFrame custom widget
- Understand tab system and layout
- Consider non-blocking operations

### 5. Common Patterns for Claude

**Adding New Analytics Features:**
1. Add method to TradeJournal or PerformanceTracker
2. Add database schema changes (migration if needed)
3. Add widget to pt_analytics_dashboard.py
4. Integrate into pt_hub.py ANALYTICS tab
5. Test with sample data
6. Document in CHANGELOG.md

**Adding New Notification Channels:**
1. Create new Notifier class in pt_notifications.py
2. Implement required methods (_send_*)
3. Add to NotificationManager initialization
4. Update NotificationConfig dataclass
5. Test rate limiting
6. Add documentation to NOTIFICATIONS_README.md

**Adding New Exchange Support:**
1. Create new Exchange class in pt_exchanges.py
2. Implement required methods (get_price, get_candles)
3. Add to ExchangeManager fallback chain
4. Add API keys to configuration
5. Test with real data
6. Update MODULE_INDEX.md

**Adding New Trading Features:**
1. Identify impact on pt_trader.py logic
2. Consider pt_thinker.py integration point
3. Update analytics logging if trades affected
4. Add notification triggers if appropriate
5. Test with simulation mode first
6. Update ROADMAP.md

### 6. Testing Guidance for Claude

**Unit Testing:**
- Write tests for new public methods
- Mock external dependencies (APIs, database)
- Test edge cases (empty data, nulls, extremes)
- Test error conditions explicitly

**Integration Testing:**
- Test with actual database
- Test with exchange sandbox environments
- Test notification delivery (use test accounts)
- Test GUI interactions

**Regression Testing:**
- Run existing tests
- Test core trading flow
- Test training and prediction cycle
- Test backtester with historical data

### 7. Debugging with Claude

**Systematic Debugging:**
1. Reproduce the issue reliably
2. Isolate the problematic code
3. Add logging/diagnostics
4. Formulate hypothesis
5. Test hypothesis
6. Apply fix
7. Verify fix

**Trading System Debugging:**
- Check API responses and error codes
- Verify price data validity
- Check account balances and permissions
- Verify neural level calculations
- Check DCA and exit logic

**GUI Debugging:**
- Check main thread blocking
- Verify event handling
- Check widget state updates
- Verify threading safety

### 8. Code Review Checklist

**For Claude:**
- [ ] Code follows universal guidelines
- [ ] Uses existing patterns from codebase
- [ ] Handles errors gracefully
- [ ] Tests pass
- [ ] Documentation updated
- [ ] No hardcoded values
- [ ] No security issues
- [ ] Performance acceptable

### 9. Security Considerations

**Claude-Specific Security Focus:**
- Validate all API keys are never committed
- Check for SQL injection in database queries
- Verify input sanitization
- Check for XSS in any web components
- Validate rate limiting is implemented
- Check for sensitive data in logs

### 10. Performance Considerations

**When Optimizing:**
- Profile before optimizing
- Focus on hot paths (trading loop, prediction)
- Consider caching expensive operations
- Optimize database queries
- Minimize blocking operations in GUI

**Common Performance Patterns:**
- Batch database operations
- Cache API responses
- Use efficient data structures
- Lazy load resources
- Async I/O for network calls

---

## Example Task Execution Flow

**Task:** "Add email notification when profit exceeds 10%"

**Claude's Approach:**

1. **Understand Requirements:**
   - Read pt_notifications.py structure
   - Understand NotificationManager API
   - Identify where profit is tracked (pt_analytics.py)
   - Determine integration point

2. **Plan Implementation:**
   - Create TODO list
   - Add profit threshold configuration
   - Add notification trigger to pt_analytics.py
   - Test with sample data
   - Update documentation

3. **Execute:**
   - Modify pt_analytics.py to emit profit events
   - Hook into pt_notifications.py NotificationManager
   - Add threshold to configuration
   - Write tests
   - Verify functionality

4. **Commit:**
   - Clear commit message
   - Reference related features
   - Update CHANGELOG.md
   - Bump version if appropriate

---

## Claude's Strengths for PowerTrader AI

1. **Architecture Design:**
   - Strong at understanding system relationships
   - Good at designing clean interfaces
   - Effective at identifying dependencies

2. **Code Quality:**
   - Writes maintainable, readable code
   - Good at refactoring for clarity
   - Understands Python best practices

3. **Integration:**
   - Excellent at coordinating multi-file changes
   - Good at understanding data flow
   - Effective at avoiding breaking changes

4. **Documentation:**
   - Strong at technical writing
   - Good at comprehensive explanations
   - Understands user needs

---

## When Claude Should Delegate

**Use Explore Agent When:**
- Searching for existing code patterns
- Finding where specific functionality is implemented
- Understanding codebase structure

**Use Librarian Agent When:**
- Looking up external API documentation
- Researching third-party libraries
- Finding implementation examples from other projects

**Use Oracle Agent When:**
- Making architectural trade-offs
- Deciding between multiple approaches
- Reviewing significant code changes

---

## Quick Reference

**File Line Counts (v2.0.0):**
- pt_hub.py: 5,835 lines
- pt_thinker.py: 1,381 lines
- pt_trader.py: 2,421 lines
- pt_trainer.py: 1,625 lines
- pt_backtester.py: 876 lines
- pt_analytics.py: 770 lines
- pt_analytics_dashboard.py: 252 lines
- pt_thinker_exchanges.py: 100 lines
- pt_exchanges.py: 663 lines
- pt_notifications.py: 876 lines
- pt_volume.py: 128 lines

**Key Classes:**
- CryptoHubApp (pt_hub.py) - Main GUI
- NeuralSystem (pt_thinker.py) - Prediction AI
- CryptoAPITrading (pt_trader.py) - Trading engine
- TradeJournal (pt_analytics.py) - Analytics logging
- NotificationManager (pt_notifications.py) - Notification coordination
- ExchangeManager (pt_exchanges.py) - Exchange interface
- VolumeAnalyzer (pt_volume.py) - Volume metrics

**Entry Points:**
- `python pt_hub.py` - Main application
- `python pt_thinker.py` - Standalone predictor
- `python pt_trader.py` - Standalone trader
- `python pt_trainer.py` - Standalone trainer
- `python pt_backtester.py` - Standalone backtester

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**Parent Document:** UNIVERSAL_LLM_INSTRUCTIONS.md
