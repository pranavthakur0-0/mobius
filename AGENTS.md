# Project Context: `mobius` — High-Performance AI Agent Harness in Go

## 1. Project Mission & Identity
`mobius` is a high-performance, modular **AI Agent Harness** and **Evaluation Benchmark Engine** written in idiomatic Go (Golang).
It is also an interactive learning codebase designed to teach Go fundamentals and systems programming concepts step-by-step.

## 2. Core Architecture
The system consists of two primary engines:

### A. The Runtime / Operational Agent Harness (Execution Engine)
* **Agent Loop**: Multi-turn ReAct (Reason-Act-Observe) state machine with step budgets and timeouts.
* **LLM Adapters**: Native Go clients for OpenAI, Anthropic Claude, Google Gemini, and Ollama/vLLM.
* **Tool Registry & MCP**: Standardized tools (`view_file`, `edit_file`, `write_file`, `list_dir`, `run_command`, `grep_search`) and Model Context Protocol support.
* **Context Manager**: Token estimation, sliding window history, and dynamic tool result compaction.
* **Sandbox & Isolation**: Process-group isolation (`setpgid`) and ephemeral Docker container execution.
* **Telemetry**: Structured JSONL trajectory recording with token/cost accounting.

### B. The Evaluation & Benchmark Harness (Test Bench)
* **Dataset Ingestion**: SWE-bench format and custom JSONL/YAML task suites.
* **Workspace Provisioner**: Clean git worktrees / ephemeral containers per task instance.
* **Evaluation Oracle**: Verification patch applier, test runner (`go test`, `pytest`), and `FAIL_TO_PASS` / `PASS_TO_PASS` validation.
* **Parallel Matrix Runner**: Concurrent task worker pool with rate-limiting and isolation.
* **Metrics & Leaderboard**: Pass@1, Pass@k, latency, token efficiency, and terminal scorecard.

## 3. Directory Layout
```
mobius/
├── cmd/
│   └── mobius/
│       └── main.go              # CLI Entrypoint (run, eval, view)
├── pkg/
│   ├── agent/                   # Runtime Agent Loop (Observe -> Reason -> Act)
│   ├── llm/                     # LLM Provider Adapters (Claude, OpenAI, Gemini)
│   ├── tools/                   # Tool Registry & Implementations (Bash, Edit, Search)
│   ├── context/                 # Context Compactor & Token Budgeter
│   ├── sandbox/                 # Subprocess & Docker Isolation
│   ├── eval/                    # Benchmark Loader, Provisioner, Oracle & Metrics
│   └── trace/                   # JSONL Trajectory Logger & Viewer
├── docs/                        # Architecture & Go Learning Guides
├── benchmarks/                  # Sample Benchmark Datasets
├── AGENTS.md                    # This persistent project context file
├── go.mod                       # Go module definition (module mobius)
└── go.sum
```

## 4. Progressive Milestones & Go Learning Roadmap
* **Milestone 1**: Tool Execution Engine (Go Structs, Interfaces, `os/exec`, File I/O, Error Handling).
* **Milestone 2**: LLM Provider Adapters (Go `net/http`, JSON Tags, Marshaling, Enums/Types).
* **Milestone 3**: Agent Runtime Loop (`context.Context`, Slices/Maps, ReAct State Machine, Token Compaction).
* **Milestone 4**: Evaluation Test Bench & Concurrency (Goroutines, Channels, `sync.WaitGroup`, Worker Pools, Git Operations).
* **Milestone 5**: CLI & Trajectory Inspector (CLI Flags, JSONL Streaming, Colored Terminal UX).

## 5. Coding & Teaching Guidelines for the Agent (STRICT TUTOR MODE)
* **User Writes the Code**: Do NOT modify or create code files directly on behalf of the user. Guide and teach from chat.
* **Explain Concepts First**: Explain *what* we are building, *why* Go works that way, and provide code examples/explanations in chat for the user to write.
* **Review & Debug**: Review what the user writes, explain compiler errors, and answer questions.
* **Idiomatic Go**: Teach clean, standard-library-first Go practices.
