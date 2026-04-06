# Gemini Model-Specific Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Instructions for Google Gemini working on PowerTrader AI and the Go ultra-project

**Parent Document:** [UNIVERSAL_LLM_INSTRUCTIONS.md](UNIVERSAL_LLM_INSTRUCTIONS.md)

> **Read `UNIVERSAL_LLM_INSTRUCTIONS.md` first.** For current strategic direction also review `VISION.md`, `ROADMAP.md`, `TODO.md`, `MEMORY.md`, and `HANDOFF.md`.

---

## Gemini-Specific Guidelines

### 1. When to Use Gemini

**Strengths of Gemini:**
- Excellent at understanding complex multi-modal inputs
- Strong reasoning and analytical capabilities
- Good at code generation and refactoring
- Effective at understanding context and relationships
- Strong at cross-language translation

**Best Use Cases for Gemini:**
- Complex code refactoring and optimization
- Multi-step task coordination with dependencies
- Understanding and documenting existing codebases
- Cross-language code generation
- Complex analytical problem-solving

**Avoid Using Gemini For:**
- Simple one-line changes (use direct tools)
- Very repetitive tasks without reasoning needs
- Simple factual lookups (use librarian agent)

### 2. Communication Style

**For Gemini:**
- Provide clear, structured requirements
- Be explicit about dependencies and relationships
- Provide context when asking about existing code
- Explain the "why" behind requests when possible

**Gemini's Communication:**
- Will analyze patterns and relationships in code
- Will suggest optimizations based on best practices
- Will ask clarifying questions when needed
- Will provide multiple approaches when applicable

### 3. Task Execution for Gemini

**Start with Analysis:**
1. Read UNIVERSAL_LLM_INSTRUCTIONS.md first
2. Identify all relevant files and their relationships
3. Understand the complete context before coding
4. Plan the full solution, not just immediate step

**Then Execute:**
1. Create comprehensive TODO for multi-step tasks
2. Execute systematically, updating progress
3. Validate each step before proceeding
4. Test thoroughly
5. Document all changes

**Commit Pattern:**
- Atomic commits for logical units
- Clear commit messages with full context
- Reference parent issues or related changes
- Update version and CHANGELOG appropriately

### 4. PowerTrader AI Specifics for Gemini

**When Working on Analytics:**
- Understand the full SQLite schema
- Analyze data flow between components
- Identify optimization opportunities in queries
- Consider performance implications of changes

**When Working on Notifications:**
- Review all notification platforms together
- Identify common patterns for DRY code
- Ensure rate limiting is consistent
- Consider error handling across all platforms

**When Working on Exchanges:**
- Analyze the unified interface design
- Understand fallback chains and edge cases
- Consider API version compatibility
- Identify opportunities for error recovery

**When Working on GUI (pt_hub.py):**
- Understand the full widget hierarchy
- Analyze the event flow and threading
- Identify opportunities for non-blocking operations
- Consider user experience across all interactions

### 5. Common Patterns for Gemini

**Code Generation:**
- Generate complete, working code
- Include necessary imports and error handling
- Follow existing code style and patterns
- Add type hints where appropriate
- Ensure code is testable

**Refactoring:**
- Preserve all existing functionality
- Improve code organization and clarity
- Follow Python best practices (PEP 8)
- Add helpful comments only when necessary
- Ensure no regressions

**Feature Addition:**
- Identify minimal set of changes needed
- Follow existing architectural patterns
- Add comprehensive error handling
- Consider all edge cases
- Update documentation appropriately

### 6. Testing Guidance for Gemini

**Write Tests That:**
- Cover all new code paths
- Test edge cases and error conditions
- Mock external dependencies appropriately
- Are maintainable and clear
- Provide confidence in code correctness

**Test Quality:**
- Test meaningful scenarios, not just code coverage
- Test error conditions explicitly
- Verify no regressions in existing functionality
- Test with realistic data when possible

### 7. Debugging with Gemini

**Approach:**
- Analyze the complete system state
- Identify root causes, not just symptoms
- Consider all interactions between components
- Propose solutions that address underlying issues
- Verify fixes don't cause regressions

**Trading System Debugging:**
- Understand the full trading flow
- Check API responses and data validity
- Verify calculations (neural levels, DCA, profit margins)
- Check database integrity
- Verify notification triggers

### 8. Code Review Checklist

**For Gemini:**
- [ ] Code is clean and well-organized
- [ ] Follows existing patterns in codebase
- [ ] Error handling is comprehensive
- [ ] Performance is acceptable
- [ ] Tests are thorough
- [ ] Documentation is updated
- [ ] No hardcoded configuration
- [ ] Security is maintained
- [ ] Code is maintainable

### 9. Security Considerations

**Gemini-Specific Focus:**
- Identify potential security issues proactively
- Validate all external inputs
- Check for SQL injection vulnerabilities
- Verify secrets are never exposed or committed
- Consider rate limiting and abuse prevention

### 10. Performance Considerations

**When Optimizing:**
- Analyze full context before optimizing
- Focus on bottlenecks in hot paths
- Consider caching strategies
- Evaluate trade-offs between memory and CPU
- Profile to confirm improvements

**Common Optimizations:**
- Database query optimization
- Efficient data structures
- Async I/O for network operations
- Caching frequently accessed data
- Minimize blocking operations in GUI

---

## Example Task Execution Flow

**Task:** "Refactor notification system to reduce code duplication"

**Gemini's Approach:**

1. **Analyze Existing Code:**
   - Read all notifier classes (Email, Discord, Telegram)
   - Identify common patterns and duplication
   - Understand NotificationManager interface
   - Analyze error handling across platforms

2. **Plan Refactoring:**
   - Extract common interface methods to base class
   - Create platform-specific implementations
   - Consolidate error handling
   - Ensure rate limiting is unified
   - Plan backward compatibility

3. **Execute:**
   - Create base Notifier class with common methods
   - Refactor EmailNotifier to extend base
   - Refactor DiscordNotifier to extend base
   - Refactor TelegramNotifier to extend base
   - Update NotificationManager to use base class
   - Update all imports
   - Run all tests
   - Verify no regressions

4. **Commit:**
   - Clear commit message explaining refactoring
   - Reference any related issues
   - Update CHANGELOG.md
   - Bump version if appropriate

---

## Gemini's Strengths for PowerTrader AI

1. **Code Analysis:**
   - Excellent at understanding existing code patterns
   - Good at identifying refactoring opportunities
   - Effective at understanding code relationships

2. **Complex Coordination:**
   - Strong at managing multi-step tasks
   - Good at identifying dependencies
   - Effective at ensuring all pieces work together

3. **Optimization:**
   - Good at identifying performance opportunities
   - Strong at suggesting efficient algorithms
   - Effective at trade-off analysis

4. **Documentation:**
   - Strong at comprehensive documentation
   - Good at explaining complex concepts
   - Effective at creating clear examples

---

## When Gemini Should Delegate

**Use Explore Agent When:**
- Need to find specific code patterns
- Need to understand codebase structure
- Need to search for existing functionality

**Use Librarian Agent When:**
- Need to research external APIs or libraries
- Need to find implementation examples
- Need to understand third-party integration patterns

**Use Oracle Agent When:**
- Making architectural decisions
- Evaluating trade-offs between approaches
- Reviewing complex code changes

---

## Quick Reference

**Project Version:** 2.0.0
**Key Files:**
- pt_hub.py (5,835 lines) - Main GUI
- pt_thinker.py (1,381 lines) - Prediction AI
- pt_trader.py (2,421 lines) - Trade execution
- pt_analytics.py (770 lines) - Analytics
- pt_notifications.py (876 lines) - Notifications
- pt_exchanges.py (663 lines) - Multi-exchange
- pt_volume.py (128 lines) - Volume analysis

**Common Tasks:**
- Add new analytics metrics
- Add notification platform
- Add exchange integration
- Refactor existing code
- Optimize performance
- Fix bugs and issues

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**Parent Document:** UNIVERSAL_LLM_INSTRUCTIONS.md
