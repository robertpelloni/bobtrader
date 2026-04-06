# Agent Instructions - PowerTrader AI

**Version:** 2.0.11
**Last Updated:** 2026-04-05
**Purpose:** Comprehensive instructions for all AI agents working on PowerTrader AI and the Go ultra-project

**Parent Document:** [UNIVERSAL_LLM_INSTRUCTIONS.md](UNIVERSAL_LLM_INSTRUCTIONS.md)

> **Always read `UNIVERSAL_LLM_INSTRUCTIONS.md` first.** Then consult `VISION.md`, `MEMORY.md`, `ROADMAP.md`, `TODO.md`, `DEPLOY.md`, and `HANDOFF.md` before major implementation sessions.

---

## Agent Overview

PowerTrader AI uses multiple specialized AI agents for different tasks. This document provides instructions for all agent types.

**Available Agents:**
- **oracle** - Senior engineering advisor for architecture decisions
- **explore** - Fast agent for exploring codebases
- **librarian** - Specialized for searching remote codebases and documentation
- **general** - General-purpose agent for multi-step tasks
- **frontend-ui-ux-engineer** - Designer-turned-developer for visual work
- **document-writer** - Technical writer for documentation
- **build** - Build specialist (manual invocation only)

---

## Universal Agent Guidelines

### 1. Agent Selection Strategy

**Choose Agent Based on Task:**

| Task Type | Best Agent | Why |
|-----------|-------------|-----|
| Architecture decisions | oracle | Senior advisor, deep reasoning |
| Code exploration | explore | Fast, contextual search |
| External research | librarian | Specialized for external sources |
| Multi-step coordination | general | Orchestrates complex tasks |
| Visual/UI changes | frontend-ui-ux-engineer | Designer expertise |
| Documentation | document-writer | Technical writing specialist |
| Build issues | build | Build system specialist |

### 2. Task Delegation Principles

**When to Delegate:**
- Task requires specialized knowledge
- Task benefits from parallel execution
- Task is too complex for direct implementation
- Task requires external research
- Task has multiple independent sub-tasks

**When NOT to Delegate:**
- Simple one-line code changes
- Obvious syntax fixes
- Direct file reads/writes
- Simple git operations

### 3. Parallel Execution Strategy

**Launch Multiple Agents When:**
- Task has multiple independent research needs
- Task requires exploring multiple code areas
- Task needs both internal and external research
- Task can be broken into independent chunks

**Example Parallel Launch:**
```
background_task(agent="explore", prompt="Find auth implementations")
background_task(agent="explore", prompt="Find error handling patterns")
background_task(agent="librarian", prompt="Find JWT best practices")
```

### 4. Background Task Management

**Task Lifecycle:**
1. Launch with `background_task()`
2. Receive `task_id` immediately
3. Continue with other work
4. Use `background_output()` when results needed
5. `background_cancel(all=true)` when done

**Best Practices:**
- Don't wait for background tasks (non-blocking)
- Use `block=true` only when absolutely necessary
- Collect multiple results with multiple `background_output()` calls
- Always cancel background tasks before final answer

### 5. Prompt Engineering

**Prompt Structure:**
1. **TASK:** Atomic, specific goal (one action per delegation)
2. **EXPECTED OUTCOME:** Concrete deliverables with success criteria
3. **REQUIRED SKILLS:** Which skill to invoke (agent-specific)
4. **REQUIRED TOOLS:** Explicit tool whitelist (prevents tool sprawl)
5. **MUST DO:** Exhaustive requirements - leave NOTHING implicit
6. **MUST NOT DO:** Forbidden actions - anticipate and block rogue behavior
7. **CONTEXT:** File paths, existing patterns, constraints

**Example Good Prompt:**
```
1. TASK: Refactor notification system to reduce code duplication
2. EXPECTED OUTCOME: Base Notifier class with common methods, EmailNotifier/DiscordNotifier/TelegramNotifier extending base
3. REQUIRED SKILLS: python, refactoring, design patterns
4. REQUIRED TOOLS: read, edit, write, lsp
5. MUST DO: Extract common methods to base class while maintaining all functionality
6. MUST NOT DO: Change public APIs, break existing functionality
7. CONTEXT: Files in pt_notifications.py, existing patterns in codebase
```

**Example Bad Prompt:**
```
Fix the notification code.
```

**Bad Prompt Issues:**
- No specific goal
- No expected outcome
- No tools specified
- No context provided
- Too vague

---

## Agent-Specific Instructions

### Oracle Agent

**When to Use:**
- Complex architecture decisions
- Multi-system trade-offs
- Designing new major subsystems
- Reviewing significant code changes
- Hard debugging after failures

**Prompt Strategy:**
- Provide full context about the decision
- Explain constraints and requirements
- Provide multiple options if applicable
- Ask for trade-off analysis

**Expected Output:**
- Detailed reasoning for decision
- Analysis of pros/cons
- Recommendation with justification
- Consideration of alternatives

### Explore Agent

**When to Use:**
- Finding existing code patterns
- Understanding codebase structure
- Searching for specific functionality
- Contextual grep operations

**Prompt Strategy:**
- Be specific about what to search for
- Provide file patterns if known
- Specify search depth (quick/medium/very thorough)
- Provide context about what you're trying to accomplish

**Expected Output:**
- File paths with matches
- Code snippets showing patterns
- Explanation of how patterns work
- Integration points and dependencies

### Librarian Agent

**When to Use:**
- Looking up external API documentation
- Researching third-party libraries
- Finding implementation examples from other projects
- Understanding library best practices

**Prompt Strategy:**
- Specify library or API name
- Specify what functionality you need
- Ask for examples and best practices
- Specify what to avoid

**Expected Output:**
- Official documentation links
- Code examples from repositories
- Best practices and common patterns
- Integration considerations

### General Agent

**When to Use:**
- Multi-step tasks with dependencies
- Complex coordination requirements
- When task requires diverse capabilities
- When no specialized agent fits perfectly

**Prompt Strategy:**
- Break down complex task into steps
- Provide clear order of operations
- Specify constraints and requirements
- Provide context about codebase

**Expected Output:**
- Systematic execution of multiple steps
- Progress updates for each step
- Final result with all steps completed
- Error handling and recovery

### Frontend UI/UX Engineer Agent

**When to Use:**
- Visual changes (colors, spacing, layout)
- UI/UX improvements
- Styling and theming
- Animation and responsive design

**Prompt Strategy:**
- Specify exact visual requirements
- Provide screenshots or descriptions if possible
- Specify which component to modify
- Keep styling separate from logic

**Expected Output:**
- Improved visual design
- Better user experience
- Consistent styling with existing design
- Responsive and accessible components

### Document-Writer Agent

**When to Use:**
- Writing README files
- Creating API documentation
- Writing technical guides
- Updating changelogs

**Prompt Strategy:**
- Specify what documentation to write
- Provide target audience information
- Specify format requirements
- Provide technical context

**Expected Output:**
- Clear, comprehensive documentation
- Proper formatting and structure
- Accurate technical information
- Examples and usage guides

---

## Common Agent Patterns for PowerTrader AI

### Pattern 1: Analytics Integration

**Oracle Prompt:**
```
Design architecture for integrating new analytics metrics into pt_analytics.py.
Consider: SQLite schema changes, impact on existing queries, performance implications.
Must maintain: All existing functionality, backward compatibility.
```

**Explore Prompt:**
```
Find all locations in pt_analytics.py where profit calculations occur.
Return: File paths with line numbers, code snippets showing calculation logic.
```

**General Prompt:**
```
1. TASK: Add profit percentage tracking to analytics system
2. EXPECTED OUTCOME: New database fields, update methods, dashboard integration
3. REQUIRED TOOLS: read, write, lsp
4. MUST DO: Add profit_pct field to trade_exits table, update PerformanceTracker, add KPI card
5. MUST NOT DO: Break existing queries, change data format
6. CONTEXT: pt_analytics.py schema, pt_analytics_dashboard.py widgets
```

### Pattern 2: Notification Integration

**Librarian Prompt:**
```
Research Discord webhook API integration best practices.
Find: Rate limiting recommendations, error handling patterns, formatting examples.
Focus on: Python discord-webhook library.
```

**Explore Prompt:**
```
Find all notification trigger points in pt_trader.py.
Return: File paths, line numbers, code snippets showing where notifications should be sent.
```

**General Prompt:**
```
1. TASK: Integrate email notifications for trade completions
2. EXPECTED OUTCOME: Email sent when trade exits with profit details
3. REQUIRED TOOLS: read, edit, write, lsp
4. MUST DO: Add notification call in pt_trader.py exit logic, configure email in pt_notifications.py
5. MUST NOT DO: Send notifications on every price update, hardcode email addresses
6. CONTEXT: pt_trader.py exit logic, pt_notifications.py EmailNotifier interface
```

### Pattern 3: Exchange Integration

**Librarian Prompt:**
```
Research Binance REST API for cryptocurrency trading.
Find: Authentication, rate limits, order placement, order history endpoints.
Focus on: Spot trading API.
```

**General Prompt:**
```
1. TASK: Add Binance as exchange option
2. EXPECTED OUTCOME: BinanceExchange class integrated, pt_exchanges.py updated
3. REQUIRED TOOLS: read, write, lsp
4. MUST DO: Implement get_price, get_candles methods, add error handling
5. MUST NOT DO: Break existing KuCoin integration, require futures trading
6. CONTEXT: pt_exchanges.py ExchangeManager interface, python-binance library docs
```

---

## Agent Task Execution Protocol

### 1. Pre-Task Planning

**Checklist:**
- [ ] Is this task suitable for agent delegation?
- [ ] Which agent is best suited?
- [ ] Can I break this into parallel tasks?
- [ ] Do I have sufficient context?
- [ ] Are requirements clear and specific?

### 2. Task Launch

**Launch Process:**
1. Formulate clear, detailed prompt
2. Choose appropriate agent
3. Use proper tool whitelist
4. Launch with `background_task()`
5. Record task_id
6. Continue with other work

**Prompt Template:**
```
1. TASK: [Specific, atomic goal]
2. EXPECTED OUTCOME: [Concrete deliverables]
3. REQUIRED SKILLS: [Agent capabilities needed]
4. REQUIRED TOOLS: [Specific tools allowed]
5. MUST DO: [Exhaustive requirements]
6. MUST NOT DO: [Forbidden actions]
7. CONTEXT: [File paths, patterns, constraints]
```

### 3. Progress Monitoring

**While Agent Works:**
- Continue with other parallel tasks
- Don't block on agent completion
- Prepare for result integration
- Plan next steps based on expected output

### 4. Result Collection

**When Agent Completes:**
1. Use `background_output(task_id="...")`
2. Review results against expected outcomes
3. Verify quality and completeness
4. Integrate into main task flow
5. Cancel agent when done

**Result Validation:**
- [ ] Does output match expected outcome?
- [ ] Is code correct and follows patterns?
- [ ] Are all requirements met?
- [ ] Is documentation updated if needed?
- [ ] Are tests passing?

### 5. Task Completion

**Before Marking Complete:**
- [ ] All deliverables received
- [ ] Quality verified
- [ ] Integrated into main flow
- [ ] No regressions introduced
- [ ] Documentation updated

**Final Steps:**
1. Cancel all background agents: `background_cancel(all=true)`
2. Provide final summary to user
3. Update todo list if applicable
4. Commit changes if appropriate

---

## Common Agent Workflows

### Workflow 1: New Feature Implementation

**Parallel Research:**
```
// Phase 1: Parallel research
background_task(agent="explore", prompt="Find similar features in codebase")
background_task(agent="librarian", prompt="Research best practices")
background_task(agent="oracle", prompt="Design architecture")
// Continue immediately, collect all results later
```

**Sequential Implementation:**
```
// Phase 2: Implement based on research
background_task(agent="general", prompt="Implement feature using architecture")
// Wait for completion
```

**Documentation:**
```
// Phase 3: Update documentation
background_task(agent="document-writer", prompt="Update README and docs")
// Wait for completion
```

### Workflow 2: Bug Investigation

**Distributed Investigation:**
```
// Parallel investigation
background_task(agent="explore", prompt="Find where bug occurs")
background_task(agent="explore", prompt="Find similar patterns in codebase")
background_task(agent="explore", prompt="Find potential causes")
// Collect all results, analyze together
```

**Solution Implementation:**
```
// Apply fix
background_task(agent="general", prompt="Implement fix based on findings")
// Wait for completion
```

### Workflow 3: Codebase Refactoring

**Pattern Discovery:**
```
// Find patterns to refactor
background_task(agent="explore", prompt="Find code duplication")
background_task(agent="explore", prompt="Find inconsistent patterns")
// Analyze results together
```

**Architecture Design:**
```
// Design refactoring approach
background_task(agent="oracle", prompt="Design refactoring plan")
// Wait for completion
```

**Implementation:**
```
// Apply refactoring
background_task(agent="general", prompt="Implement refactoring")
// Wait for completion
```

---

## Agent Communication Best Practices

### 1. Providing Context

**Always Include:**
- Project structure overview
- Relevant file paths
- Existing patterns to follow
- Constraints and requirements
- What to avoid

**Example:**
```
Context: PowerTrader AI is a crypto trading bot with these modules:
- pt_hub.py: Main GUI (5835 lines)
- pt_thinker.py: Prediction AI (1381 lines)
- pt_trader.py: Trading engine (2421 lines)

Existing pattern: All exchange integrations use pt_exchanges.py ExchangeManager class with get_price() and get_candles() methods.
```

### 2. Setting Clear Expectations

**Specify:**
- What should be delivered
- Success criteria
- What should NOT be done
- Format of output
- Deadlines or constraints

### 3. Feedback Loop

**When Results Don't Match:**
- Provide specific feedback
- Explain what's missing
- Ask for corrections
- Relaunch with refined prompt

---

## Resource Management

### 1. Background Task Limits

**Best Practices:**
- Don't launch more than 5 parallel tasks
- Cancel tasks when no longer needed
- Collect results promptly
- Monitor task status

**Cleanup:**
- Always `background_cancel(all=true)` before final answer
- This conserves resources
- Ensures clean workflow completion

### 2. Tool Usage

**Tool Whitelist:**
- Specify allowed tools to prevent sprawl
- Only necessary tools for the task
- Prevents agent from using expensive tools

**Example:**
```
REQUIRED TOOLS: read, write, edit, lsp
// Not allowed: grep, glob, bash, websearch, webfetch
```

---

## Error Handling with Agents

### 1. Agent Failures

**When Agent Fails:**
1. Review the failure reason
2. Adjust prompt or choose different agent
3. Provide more specific context
4. Simplify task if needed
5. Try alternative approach

### 2. Partial Results

**When Results Are Incomplete:**
1. Use what's available
2. Request missing information specifically
3. Fill gaps with direct tools
4. Document what was vs. wasn't delivered

### 3. Quality Issues

**When Agent Quality is Low:**
1. Provide specific feedback
2. Ask for corrections
3. Use oracle agent to review work
4. Manually fix if necessary

---

## Documentation Requirements for Agents

### 1. Agent-Specific Instructions

**Must Include:**
- When to use the agent
- Agent's strengths and weaknesses
- Best use cases
- Prompt patterns
- Expected outputs
- When not to use

### 2. Task Execution Protocols

**Must Document:**
- Pre-task checklist
- Task launch process
- Progress monitoring
- Result collection
- Task completion criteria

### 3. Common Patterns

**Must Provide:**
- Typical workflows for PowerTrader AI
- Example prompts for each agent
- Best practices for delegation
- Error handling strategies

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!

**Last Updated:** 2026-01-18
**Current Version:** 2.0.0
**Parent Document:** UNIVERSAL_LLM_INSTRUCTIONS.md
