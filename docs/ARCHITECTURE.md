# Mobius Architecture & Subsystems

This document details the architectural design, subsystem interaction, and dataflow of the **Mobius** AI Agent Harness.

---

## 1. Subsystem Decomposition

```text
mobius/
├── pkg/
│   ├── agent/       # State machine, step loop, budget enforcement
│   ├── agentctx/    # Token sensor, 8-section checkpoint summarizer, compactor
│   ├── artifact/    # Observation interceptor (>2k chars) & disk store
│   ├── budget/      # Live pricing, prompt/completion metering, budget ceiling
│   ├── cli/         # Interactive REPL, raw terminal menu & navigation
│   ├── events/      # Append-only JSONL EventStore telemetry
│   ├── llm/         # Multi-provider client adapters (OpenAI, DeepSeek, Claude)
│   ├── session/     # Multi-session memory management
│   ├── tools/       # Standardized tool registry (Bash, Edit, View, Write, List, Grep)
│   └── utils/       # ID generators & shared utilities
```

---

## 2. The ReAct Runtime Loop (`pkg/agent/loop.go`)

The execution loop follows the Reason-Act-Observe state machine:

```text
                    [User Task Input]
                            │
                            ▼
              ┌───────────────────────────┐
              │  1. Check Timeout/Cancel  │
              └─────────────┬─────────────┘
                            │
                            ▼
              ┌───────────────────────────┐
              │  2. Token Watermark Sensor│
              │     (Is Tokens >= 80%?)   │
              └─────────────┬─────────────┘
                     ├── YES ──► Compact Older Turns (pkg/agentctx)
                     └── NO  ──► Proceed
                            │
                            ▼
              ┌───────────────────────────┐
              │  3. Model Generation      │
              │     (provider.Generate)   │
              └─────────────┬─────────────┘
                            │
                            ▼
              ┌───────────────────────────┐
              │  4. Budget Accounting     │
              │     (tracker.Add & Check) │
              └─────────────┬─────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        [No Tool Calls]            [Tool Calls Requested]
              │                           │
              ▼                           ▼
      Return Final Answer         ┌───────────────────────────┐
                                  │  5. Tool Execution        │
                                  │     (pkg/tools/...)       │
                                  └─────────────┬─────────────┘
                                                │
                                                ▼
                                  ┌───────────────────────────┐
                                  │  6. Artifact Interceptor  │
                                  │     (pkg/artifact)        │
                                  └─────────────┬─────────────┘
                                                │
                                                ▼
                                  ┌───────────────────────────┐
                                  │  7. EventStore Telemetry  │
                                  │     (pkg/events)          │
                                  └─────────────┬─────────────┘
                                                │
                                                ▼
                                          Loop to Step N+1
```

---

## 3. Context Compactor Subsystem (`pkg/agentctx`)

### The 4-Stage Reduction Pipeline:
1. **Pre-flight Metering (`token.go`)**: Fast in-memory UTF-8 rune counter ($\approx 4$ chars/token + tool framing overhead) calculates live token consumption before sending chat requests.
2. **Watermark Trigger**: When conversation history exceeds `DefaultWatermarkRatio` (0.80) of `MaxTokens` (128,000), compaction is invoked.
3. **Tool-Pairing Integrity Guard (`findSafeCutIndex`)**: Scans backward through message history to preserve recent working context (20% tail) while verifying that assistant `tool_calls` are never detached from their corresponding `tool` results.
4. **Structured Checkpoint (`summarizer.go`)**:
   * Appends `CompactionInstruction` at the end of older messages to reuse the provider's warm KV-cache prefix.
   * Forces the model to emit a strict **8-Section Checkpoint**:
     * `## Primary Request and Intent`
     * `## Key Technical Concepts`
     * `## Files and Code`
     * `## Errors and Fixes`
     * `## Pending Jobs`
     * `## Current Work`
     * `## Next Step`
     * `## Critical Context`
5. **Shrink Check & Splicing**: Replaces old turns in memory with the `<compacted-summary>` checkpoint while retaining the system prompt and recent working tail.

---

## 4. Artifact Interception & Offloading (`pkg/artifact`)

* **Threshold Trigger**: Tool observations exceeding `SmallThreshold` (2,000 characters) are automatically stored on disk in `.mobius/artifacts/<thread_id>/`.
* **Context Preservation**: The model receives a preview (first 20 lines) and an `artifact://` reference pointer, preventing context bloat from commands like `git diff` or large file dumps.

---

## 5. Budget & Cost Enforcement (`pkg/budget`)

* Calculates exact USD and INR costs based on model-specific prompt and completion pricing per million tokens.
* Aborts immediately if cumulative spend crosses `MaxCost` configured in `agent.toml`.

---

## 6. Telemetry & Event Store (`pkg/events`)

* Durable, append-only JSONL log per thread (`.mobius/events/<thread_id>.jsonl`).
* Records all lifecycle events: `EventUserMessage`, `EventAssistantMessage`, `EventToolCall`, `EventToolResult`, with exact token counts and timestamps for offline analysis and trajectory replay.
