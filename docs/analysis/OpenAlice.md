# Architectural Analysis: TraderAlice/OpenAlice

## Overview
OpenAlice is an architectural reference that emphasizes **modularity, event-driven state, and tool-centric design**. It is built with TypeScript and follows a clean separation of concerns.

## Key Architectural Patterns

### 1. Event-Driven Core (`EventBus`)
- **Facade Pattern:** The `EventBus` provides a clean facade over a persistent `EventLog`.
- **In-Process Communication:** Allows plugins and custom tools to emit events without needing to understand the underlying transport or storage (JSONL logs).
- **Fan-out:** Events are appended to a log and then distributed through a `ListenerRegistry`.
- **Assimilation Strategy for Go:** Strengthen the existing `internal/core/eventlog` with a more ergonomic `Bus` interface and ensure all major state transitions in `ultratrader-go` are emitted as events.

### 2. Unified Tool Registry (`ToolCenter`)
- **Centralized Definition:** All capabilities (tools) are registered in a single `ToolCenter` during bootstrap.
- **Provider-Agnostic:** The registry can export tools in different formats (Vercel AI SDK, MCP, etc.) depending on the consumer.
- **Enabled/Disabled Logic:** Integrated configuration check to enable or disable tools dynamically.
- **Assimilation Strategy for Go:** Implement a `Registry` pattern for execution strategies and risk guards that supports metadata extraction for operator dashboards.

### 3. Modular Service Architecture
- **Domain Separation:** Services are clearly separated into `auth`, `uta-client`, `uta-supervisor`, etc.
- **Domain-Driven Design:** The `src/domain` directory likely contains the core entities and logic, kept separate from the infrastructure/server layer.

## Implementation Takeaways for UltraTrader Go
1. **Execution Management:** OpenAlice's tool-centric approach suggests that execution should be handled by a coordinator that can dispatch to specialized execution "tools" or strategies.
2. **Event Sourcing:** All trading actions (Order Placed, Order Filled, Guard Blocked) should be first-class events in the `EventLog`.
3. **Registry Pattern:** Move from hardcoded strategy lists to a dynamic `Registry` that the `ExecutionManager` uses to resolve and execute trading intent.
