# GPT Model-Specific Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Instructions for OpenAI GPT working on PowerTrader AI and the Go ultra-project

**Parent Document:** [UNIVERSAL_LLM_INSTRUCTIONS.md](UNIVERSAL_LLM_INSTRUCTIONS.md)

> **Read `UNIVERSAL_LLM_INSTRUCTIONS.md` first.** For current strategic direction also review `VISION.md`, `ROADMAP.md`, `TODO.md`, `MEMORY.md`, and `HANDOFF.md`.

---

## GPT-Specific Guidelines

### 1. When to Use GPT

**Strengths of GPT:**
- Excellent at following complex instructions
- Strong at code generation from descriptions
- Good at understanding context and requirements
- Effective at debugging and troubleshooting
- Strong at architectural decision making

**Best Use Cases for GPT:**
- Code generation from detailed specifications
- Debugging complex issues
- Refactoring with specific requirements
- Writing comprehensive documentation
- API integration and implementation
- Testing and validation

**Avoid Using GPT For:**
- Very simple syntax fixes (use direct tools)
- Understanding project structure (use explore agent)
- Looking up external documentation (use librarian agent)
- Repetitive pattern tasks without variation

### 2. Communication Style

**For GPT:**
- Be precise and specific in requirements
- Provide clear context about project structure
- Specify exact file paths and locations
- Provide examples when helpful
- Be explicit about what to avoid

**GPT's Communication:**
- Will follow instructions precisely
- Will ask for clarification when ambiguous
- Will provide detailed code with comments
- Will explain reasoning when helpful

### 3. Task Execution for GPT

**Start with Requirements Analysis:**
1. Read UNIVERSAL_LLM_INSTRUCTIONS.md first
2. Identify specific files and their exact locations
3. Understand project structure and patterns
4. Clarify any ambiguous requirements

**Then Execute:**
1. Break down complex tasks into smaller steps
2. Implement each step systematically
3. Test as you go
4. Handle errors gracefully
5. Document all changes

**Commit Pattern:**
- Commit frequently after logical units
- Use descriptive commit messages
- Reference related TODOs or issues
- Update version and CHANGELOG appropriately

### 4. PowerTrader AI Specifics for GPT

**When Working on Analytics:**
- Follow exact database schema in pt_analytics.py
- Use correct SQL queries with proper escaping
- Handle SQLite-specific constraints
- Consider query performance

**When Working on Notifications:**
- Follow exact API requirements for each platform
- Implement rate limiting correctly
- Handle platform-specific errors
- Test with mock data when possible

**When Working on Exchanges:**
- Follow exact API specifications
- Implement proper error handling and retries
- Handle rate limits correctly
- Validate all responses before use

**When Working on GUI (pt_hub.py):**
- Follow exact Tkinter patterns used in existing code
- Use correct theme constants (DARK_BG, DARK_FG, etc.)
- Handle threading correctly (no GUI updates from background)
- Follow exact widget patterns (WrapFrame, NeuralSignalTile)

### 5. Common Patterns for GPT

**Adding New Feature:**
1. Identify exact file(s) to modify
2. Read existing code to understand patterns
3. Implement new functionality following patterns
4. Add tests if applicable
5. Update documentation
6. Test thoroughly

**Bug Fixing:**
1. Reproduce issue reliably
2. Identify root cause
3. Implement minimal fix
4. Verify fix works
5. Check for regressions
6. Update tests if needed

**Refactoring:**
1. Understand existing functionality completely
2. Plan refactoring approach
3. Make incremental changes
4. Test after each change
5. Ensure no functionality is broken

### 6. Testing Guidance for GPT

**Unit Tests:**
- Test individual functions in isolation
- Mock external dependencies
- Test edge cases and error conditions
- Use descriptive test names
- Keep tests focused

**Integration Tests:**
- Test interactions between components
- Test with real or mock databases
- Test API integrations
- Test end-to-end flows

### 7. Code Quality for GPT

**When Writing Code:**
- Follow PEP 8 style guidelines
- Use descriptive variable names
- Add type hints for public functions
- Include docstrings for public APIs
- Write self-documenting code

**Error Handling:**
- Catch specific exceptions
- Provide meaningful error messages
- Log errors with context
- Handle edge cases
- Fail gracefully

### 8. GPT-Specific Strengths

**For PowerTrader AI:**
- Good at implementing specific requirements
- Good at following existing patterns
- Good at writing clean, working code
- Good at debugging complex issues
- Good at API integration

### 9. Quick Reference

**Important Constants (from pt_hub.py):**
```python
DARK_BG = "#070B10"
DARK_BG2 = "#0B1220"
DARK_PANEL = "#0E1626"
DARK_PANEL2 = "#121C2F"
DARK_BORDER = "#243044"
DARK_FG = "#C7D1DB"
DARK_MUTED = "#8B949E"
DARK_ACCENT = "#00FF66"
DARK_ACCENT2 = "#00E5FF"
DARK_SELECT_BG = "#17324A"
DARK_SELECT_FG = "#00FF66"
```

**Key Module Imports:**
```python
# For notifications
from pt_notifications import NotificationManager, NotificationConfig

# For analytics
from pt_analytics import TradeJournal, get_dashboard_metrics

# For exchanges
from pt_thinker_exchanges import get_aggregated_current_price

# For volume
from pt_volume import VolumeAnalyzer, VolumeMetrics
```

---

## When GPT Should Delegate

**Use Explore Agent When:**
- Need to find existing code patterns
- Need to understand file structure
- Need to search for specific functionality

**Use Librarian Agent When:**
- Need to research external APIs
- Need to find implementation examples
- Need to understand third-party libraries

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**Parent Document:** UNIVERSAL_LLM_INSTRUCTIONS.md
