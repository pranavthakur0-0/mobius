# Mobius

<p align="center">
  <img src="public/mascot.png" alt="Mobius Mascot" width="280" />
</p>

<p align="center">
  <strong>High-Performance, Modular AI Agent Harness & Evaluation Engine in Idiomatic Go</strong>
</p>

<p align="center">
  <a href="#architecture"><img src="https://img.shields.io/badge/Architecture-Modular%20Harness-blue?style=flat-square" alt="Architecture"></a>
  <a href="#features"><img src="https://img.shields.io/badge/Language-Go%201.24+-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="#license"><img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License"></a>
  <a href="#context-compaction"><img src="https://img.shields.io/badge/Context-Auto--Compacting-orange?style=flat-square" alt="Context Engine"></a>
  <a href="#testing"><img src="https://img.shields.io/badge/Linter-golangci--lint-brightgreen?style=flat-square" alt="Linter"></a>
</p>

---

## ⚡ Overview

**Mobius** is a modular, high-throughput autonomous AI agent harness built from the ground up in idiomatic Go. Designed with standard-library-first principles, it delivers low-latency ReAct reasoning loops, automated context compaction, artifact offloading, granular budget accounting, and robust event telemetry.

---

## 🏛️ Architecture

```text
                               ┌────────────────────────────────────────┐
                               │           Interactive REPL             │
                               │      (Multi-Session & /models)         │
                               └───────────────────┬────────────────────┘
                                                   │
                                                   ▼
                               ┌────────────────────────────────────────┐
                               │           ReAct Agent Loop             │
                               │    (Step Budget & Timeout Guard)       │
                               └───────┬────────────────────────┬───────┘
                                       │                        │
                      ┌────────────────┴──────────────┐         │
                      ▼                               ▼         ▼
        ┌───────────────────────────┐   ┌───────────────────────────┐   ┌───────────────────────────┐
        │  Context Compactor        │   │  Artifact Interceptor     │   │  Budget & Pricing Tracker │
        │  • 80% Watermark Sensor   │   │  • >2k char offload       │   │  • Live USD & INR rates   │
        │  • 8-Section Checkpoint   │   │  • Disk-backed store      │   │  • Hard max_cost ceiling  │
        │  • Tool-Pairing Integrity │   │  • In-memory previews     │   │  • Token accounting       │
        └───────────────────────────┘   └───────────────────────────┘   └───────────────────────────┘
                      │                               │                         │
                      └───────────────────────────────┼─────────────────────────┘
                                                      │
                                                      ▼
                                        ┌───────────────────────────┐
                                        │    Structured EventStore  │
                                        │  (.mobius/events/*.jsonl) │
                                        └───────────────────────────┘
```

---

## 🚀 Key Features

### 1. Autonomous ReAct Runtime Loop
* **Step-Budgeted Reasoning**: Enforces configurable `max_steps` ceilings and execution timeouts per turn.
* **Universal Tool Suite**: Built-in tools: `run_command`, `view_file`, `edit_file`, `write_file`, `list_dir`, `grep_search`.

### 2. Automatic Context Compaction (`pkg/agentctx`)
* **Sub-Microsecond Token Meter**: Zero-allocation heuristic estimator ($\sim 4$ chars/token + framing overhead) monitors live context window consumption.
* **High-Watermark Trigger**: Automatically triggers compaction when conversation reaches 80% of model capacity.
* **Tool-Pairing Integrity Guard**: Scans history boundaries to guarantee assistant `tool_calls` and corresponding `tool` results are never separated.
* **KV-Cache Alignment**: Replays the conversation prefix with an 8-section checkpoint directive (`<compacted-summary>`), maximizing upstream KV-cache hit rates.
* **Shrink Check**: Verifies that new checkpoint summaries strictly reduce context size before committing mutations.

### 3. Large-Observation Artifact Offloading (`pkg/artifact`)
* **Threshold Interception**: Tool outputs exceeding 2,000 characters are automatically offloaded to `.mobius/artifacts/`.
* **Compact Previews**: Injects structured summaries and previews into model context while preserving full outputs for downstream tools.

### 4. Dynamic Pricing & Budget Guard (`pkg/budget`)
* **Real-Time Cost Accounting**: Tracks prompt and completion tokens across every generation.
* **Dual Currency Display**: Live formatted spend tracking in USD and INR.
* **Hard Stop Safety**: Immediately aborts loops if lifetime spend crosses configured `max_cost`.

### 5. Multi-Provider LLM Switching (`/models`)
* Hot-swap between providers (**OpenAI, DeepSeek, Anthropic, Gemini, Ollama**) and models on the fly during active REPL sessions.

### 6. Durable Event Telemetry (`pkg/events`)
* Append-only structured JSONL event streaming (`.mobius/events/<thread_id>.jsonl`) capturing every user prompt, model thought, tool invocation, and token count.

---

## 📦 Project Structure

```text
mobius/
├── .github/                     # GitHub Actions CI workflow & issue templates
├── bin/                         # Compiled binaries
├── cmd/
│   └── mobius/
│       └── main.go              # CLI & REPL entrypoint
├── config/                      # Configuration files
│   ├── model_config.toml        # Provider & model definitions with pricing
│   └── agent.toml               # Agent step limits & runtime budgets
├── docs/                        # Architecture & roadmap documentation
│   ├── ARCHITECTURE.md
│   └── ROADMAP.md
├── examples/                    # Programmatic Go SDK usage & custom tools
│   ├── basic_agent/
│   └── custom_tool/
├── pkg/
│   ├── agent/                   # ReAct runtime loop & execution engine
│   ├── agentctx/                # Token sensor, 8-section summarizer & compactor
│   ├── artifact/                # Large output interceptor & disk-backed store
│   ├── budget/                  # Token accounting, dynamic pricing & budget limits
│   ├── cli/                     # Interactive REPL, arrow-key menus & terminal UX
│   ├── events/                  # Append-only JSONL EventStore telemetry
│   ├── llm/                     # Provider adapters (OpenAI, Claude, Gemini, DeepSeek)
│   ├── session/                 # Multi-session memory manager
│   ├── tools/                   # Tool registry & native implementations
│   └── utils/                   # ID generators & shared helpers
├── scripts/                     # Local setup & verification scripts
│   └── setup.sh
├── .editorconfig                # Universal indentation & formatting rules
├── .env.example                 # API keys template
├── .golangci.yml                # Production linter configuration
├── Makefile                     # Build & test automation
└── README.md
```

---

## 🛠️ Getting Started

### Prerequisites
* **Go 1.24+**
* API key for your preferred LLM provider (e.g. DeepSeek, OpenAI, Anthropic, or Gemini)

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/pranavthakur0-0/mobius.git
cd mobius

# Set up your environment variables
cp .env.example .env
# Edit .env with your API keys

# Build the binary
make build
```

---

## 💻 Usage

### 1. Interactive REPL Mode
Launch the interactive terminal:
```bash
./bin/mobius
# Or: make run
```

### 2. One-Shot Execution Mode
Pass a task directly as a command-line argument:
```bash
./bin/mobius "Analyze the repository and fix any race conditions in pkg/agent"
```

### 3. Programmatic Go SDK Usage
You can embed `mobius` into your own Go applications:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "mobius/pkg/agent"
    "mobius/pkg/agentctx"
    "mobius/pkg/events"
    "mobius/pkg/llm"
    "mobius/pkg/tools"
)

func main() {
    provider := llm.NewOpenAIProvider(os.Getenv("DEEPSEEK_API_KEY"), "https://api.deepseek.com/v1")
    registry := tools.NewDefaultRegistry(".")
    eventStore, _ := events.NewFileEventStore(".mobius/events", 256)
    defer eventStore.Close()

    ag, _ := agent.NewAgent(provider, registry, "deepseek-chat", 0.14, 0.28, eventStore)
    conv := agentctx.NewConversationContext(agentctx.BuildSystemPrompt(""))

    response, _ := ag.Run(context.Background(), conv, "Explain the project structure.")
    fmt.Println(response)
}
```
*(See [`examples/basic_agent/main.go`](examples/basic_agent/main.go) and [`examples/custom_tool/main.go`](examples/custom_tool/main.go) for more)*

---

## ⌨️ REPL Commands

| Command | Description |
| :--- | :--- |
| `/models` | Open the interactive menu to switch models and providers dynamically |
| `/newchat` | Start a new isolated conversation session with fresh context |
| `/listchats` | List all saved sessions (active session marked with `*`) |
| `/switch <id>` | Switch to an existing session by ID |
| `exit` / `quit` | Exit Mobius |

---

## ⚙️ Configuration

Configure providers and models in `config/model_config.toml`:

```toml
default_model = "deepseek-chat"

[providers.deepseek]
name     = "DeepSeek"
type     = "openai"
base_url = "https://api.deepseek.com/v1"
env_key  = "DEEPSEEK_API_KEY"
models   = ["deepseek-chat", "deepseek-reasoner"]

[providers.openai]
name     = "OpenAI"
type     = "openai"
base_url = "https://api.openai.com/v1"
env_key  = "OPENAI_API_KEY"
models   = ["gpt-4o", "gpt-4o-mini", "o3-mini"]
```

Configure agent runtime budgets in `config/agent.toml`:

```toml
max_steps = 25
max_cost = 2.00
timeout_seconds = 300
```

---

## 🧪 Development & Testing

Mobius includes full build automation via `Makefile`:

```bash
# Run unit test suite
make test

# Run tests with race detection
make test-race

# Run linter (golangci-lint)
make lint

# Format code (gofmt)
make fmt

# Format, vet, lint, test, and build in one command
make all
```

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more information.
