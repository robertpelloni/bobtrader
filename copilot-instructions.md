# GitHub Copilot Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Instructions for GitHub Copilot working on PowerTrader AI and the Go ultra-project

**Parent Document:** [UNIVERSAL_LLM_INSTRUCTIONS.md](UNIVERSAL_LLM_INSTRUCTIONS.md)

> **Read `UNIVERSAL_LLM_INSTRUCTIONS.md` first.** For current implementation direction also review `VISION.md`, `ROADMAP.md`, `TODO.md`, and `HANDOFF.md`.

---

## Copilot-Specific Guidelines

### 1. When to Use Copilot

**Strengths of Copilot:**
- Excellent at following existing code patterns
- Strong at suggesting completions based on context
- Good at writing boilerplate and repetitive code
- Effective at maintaining code consistency
- Good at detecting common patterns

**Best Use Cases for Copilot:**
- Writing repetitive code (similar patterns)
- Generating boilerplate code
- Following established patterns
- Writing test cases
- Implementing standard APIs
- Completing common code structures

**Avoid Using Copilot For:**
- Complex architectural decisions (use oracle agent)
- Understanding project architecture (use explore agent)
- Researching external libraries (use librarian agent)
- Novel implementations not based on existing patterns

### 2. Communication Style

**For Copilot:**
- Provide clear context about what you want
- Let Copilot suggest based on existing code
- Review suggestions carefully before accepting
- Be specific about what patterns to follow

**Copilot's Behavior:**
- Will suggest based on surrounding code
- Will follow existing naming conventions
- Will suggest common patterns
- Will adapt to your coding style

### 3. Working with Copilot in PowerTrader AI

**Context Awareness:**
- Copilot works best with good context
- Keep relevant files open while coding
- Let Copilot see the patterns you want to follow
- Review suggestions before accepting

**Pattern Following:**
- Let Copilot suggest based on existing code
- Use similar structures to existing implementations
- Maintain naming consistency
- Follow established error handling patterns

**Code Review:**
- Always review Copilot suggestions
- Verify suggestions meet requirements
- Check for security issues
- Ensure no regressions

### 4. Copilot-Specific Patterns for PowerTrader AI

**Common Patterns in Codebase:**

**Tkinter GUI Patterns (pt_hub.py):**
```python
# Frame creation
frame = ttk.Frame(parent)
frame.pack(fill="both", expand=True, padx=10, pady=10)

# Label creation
label = ttk.Label(frame, text="Label Text")
label.pack(anchor="center", padx=5, pady=5)

# Button creation
button = ttk.Button(frame, text="Button", command=callback)
button.pack(anchor="center", padx=5, pady=5)
```

**Database Patterns (pt_analytics.py):**
```python
# SQLite connection
conn = sqlite3.connect(database_path)
cursor = conn.cursor()

# Query execution
cursor.execute("SELECT * FROM table WHERE condition = ?", (value,))
results = cursor.fetchall()

# Transaction handling
conn.commit()
conn.close()
```

**Error Handling Patterns:**
```python
try:
    # Operation
except Exception as e:
    print(f"Error: {e}")
    # Handle gracefully
```

**Exchange API Patterns:**
```python
# API call with error handling
try:
    response = api.call()
    return response.data
except Exception as e:
    print(f"API Error: {e}")
    return None
```

### 5. Best Practices with Copilot

**Accept Suggestions When:**
- They follow existing patterns
- They use correct imports
- They maintain code style
- They don't introduce security issues
- They meet the requirements

**Review Suggestions Before Accepting:**
- Check for correctness
- Verify they follow patterns
- Check for potential bugs
- Consider performance implications
- Ensure they're necessary

**Edit Suggestions Carefully:**
- Copilot may not understand full context
- Suggestions may need adjustments
- Always verify logic is correct
- Add necessary error handling
- Ensure type hints are correct

### 6. Copilot Limitations

**What Copilot May Not Do Well:**
- Complex architectural decisions
- Understanding business logic not visible in code
- Novel solutions not based on patterns
- Security considerations beyond code patterns
- Performance optimization without profiling

**When to Seek Help:**
- Architectural decisions → Use oracle agent
- Understanding codebase → Use explore agent
- External library research → Use librarian agent
- Complex debugging → Use Claude or GPT

### 7. Code Style Consistency

**Naming Conventions:**
- Follow existing variable names in the file
- Use descriptive names (no single letters except loops)
- Match existing class/method naming
- Use underscores for private methods

**Import Organization:**
- Keep imports at top of file
- Group standard library, third-party, local imports
- Follow existing import order
- Use absolute imports for local modules

**Comment Style:**
- Follow existing comment style in file
- Comment WHY, not WHAT (unless complex)
- Keep comments brief and clear
- Update comments when code changes

### 8. Testing with Copilot

**Test Generation:**
- Let Copilot generate test boilerplate
- Adjust test cases to be meaningful
- Add edge cases manually
- Verify tests actually test requirements
- Run tests to ensure they pass

**Test Patterns:**
```python
# Unit test pattern
def test_function():
    # Setup
    result = function_to_test(input)
    # Assert
    assert result == expected

# Integration test pattern
def test_integration():
    # Setup
    # Execute
    # Verify
```

### 9. Common Copilot Patterns for PowerTrader AI

**Database Operations:**
- Use parameterized queries (prevent SQL injection)
- Use transactions for multi-step operations
- Close connections properly
- Handle database errors gracefully

**API Calls:**
- Add timeout to requests
- Implement retry logic for transient failures
- Validate responses before use
- Handle rate limits

**GUI Updates:**
- Don't update GUI from background threads
- Use `after()` for deferred updates
- Handle widget lifecycle properly
- Manage event loops correctly

### 10. Copilot and Version Management

**When Updating Version:**
- Update VERSION.md with new version
- Update CHANGELOG.md with changes
- Ensure all version references are updated
- Commit version bump with clear message

**Version Display:**
- Read version from VERSION.md (don't hardcode)
- Display in GUI (already implemented in pt_hub.py)
- Use in documentation (read from VERSION.md)

---

## Copilot Strengths for PowerTrader AI

1. **Pattern Recognition:**
   - Excellent at following existing patterns
   - Good at maintaining consistency
   - Effective at writing similar code structures

2. **Boilerplate Generation:**
   - Good at generating repetitive code
   - Fast at creating standard structures
   - Effective at test case generation

3. **Context Awareness:**
   - Good at understanding immediate context
   - Adapts to coding style
   - Learns from patterns in codebase

---

## When Copilot Should Not Be Used

**Avoid For:**
- Complex architectural decisions (use oracle agent)
- Understanding full codebase structure (use explore agent)
- External library research (use librarian agent)
- Novel implementation without existing patterns

**Use For:**
- Writing repetitive code
- Following established patterns
- Generating boilerplate
- Implementing standard APIs
- Test case generation

---

## Quick Reference

**PowerTrader AI Constants (from pt_hub.py):**
- Dark theme colors: DARK_BG, DARK_FG, DARK_ACCENT, etc.
- Widget patterns: ttk.Frame, ttk.Label, ttk.Button
- Database patterns: SQLite with parameterized queries
- Error handling: try/except with logging

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**Parent Document:** UNIVERSAL_LLM_INSTRUCTIONS.md
