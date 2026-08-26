# `mobius` — Development & Learning Roadmap

This document tracks our progressive build of the **Go Agent Harness** and the Go concepts mastered at each step.

---

## 🎯 Milestone Tracker

- [x] **Milestone 1: Tool Execution Engine (The Agent's "Hands")**
  - [x] Tool interface & Registry (`pkg/tools/types.go`, `pkg/tools/registry.go`)
  - [x] Command Runner tool with timeout (`pkg/tools/bash.go`)
  - [x] File Viewer with line slicing (`pkg/tools/view.go`)
  - [x] File Editor with chunk replacement (`pkg/tools/edit.go`)
  - [x] File Writer with safe overwrite (`pkg/tools/write.go`)
  - [x] Directory Lister & Grep Search (`pkg/tools/list.go`, `pkg/tools/grep.go`)
  - *Go Concepts Learned: Packages, Structs, Interfaces, Error Handling, Pointer Receivers, `os/exec`, `io`, `bufio`*

- [x] **Milestone 2: LLM Provider Client (The Agent's "Brain")**
  - [x] Provider interface & message schema (`pkg/llm/types.go`)
  - [x] OpenAI / DeepSeek / Ollama compatible adapter (`pkg/llm/openai.go`)
  - [x] Provider registry & factory (`pkg/llm/registry.go`, `pkg/llm/factory.go`)
  - [x] Dynamic model switcher & interactive REPL menu (`pkg/cli/menu.go`, `/models`)
  - *Go Concepts Learned: `net/http`, JSON tags & marshaling, Type switches, Environment variables, Raw terminal I/O*

- [x] **Milestone 3: Agent Runtime Loop & Context Engine**
  - [x] Multi-turn ReAct execution state machine (`pkg/agent/loop.go`)
  - [x] Step budgets, timeout guards & interrupt handling (`pkg/agent/loop.go`)
  - [x] Token estimation & 80% watermark trigger (`pkg/agentctx/token.go`)
  - [x] 8-section checkpoint summarizer & KV-cache reuse (`pkg/agentctx/summarizer.go`)
  - [x] Safe cut-point & tool-pairing integrity compactor (`pkg/agentctx/compactor.go`)
  - [x] Observation interceptor & artifact store (`pkg/artifact/`)
  - [x] Dynamic pricing & live USD/INR budget tracker (`pkg/budget/tracker.go`)
  - [x] Structured append-only JSONL EventStore (`pkg/events/store.go`)
  - *Go Concepts Learned: `context.Context` (timeouts & cancellation), Slices & memory layout, String/Rune metrics, State machines*

- [ ] **Milestone 4: Evaluation Benchmark Harness & Concurrency**
  - [ ] Dataset loader for SWE-bench & custom JSONL (`pkg/eval/dataset.go`)
  - [ ] Git worktree / workspace provisioner (`pkg/eval/provisioner.go`)
  - [ ] Test oracle & patch verifier (`pkg/eval/oracle.go`)
  - [ ] Concurrent matrix runner with worker pool (`pkg/eval/runner.go`)
  - [ ] Scorecard & Pass@k metrics calculator (`pkg/eval/metrics.go`)
  - *Go Concepts to Learn: Goroutines, Channels (`chan`), `sync.WaitGroup`, `sync.Mutex`, Worker pools, Race condition prevention*

- [ ] **Milestone 5: CLI & Trajectory Inspector**
  - [ ] CLI subcommands (`cmd/mobius/main.go` - `run`, `eval`, `view`)
  - [ ] Interactive terminal trajectory visualizer (`pkg/trace/viewer.go`)
  - *Go Concepts to Learn: CLI argument parsing, TUI styling, Stream processing*
