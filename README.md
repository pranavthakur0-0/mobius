# Mobius

A high-performance, modular **AI Agent Harness** written in idiomatic Go.

Mobius implements a multi-turn **ReAct** (Reason → Act → Observe) loop that connects to LLM providers, executes tools autonomously, and maintains persistent conversation sessions.

---

## Features

- **ReAct Agent Loop** — Step-budgeted reasoning with automatic tool execution
- **Multi-Provider LLM Support** — DeepSeek, OpenAI, Anthropic, Gemini, Ollama (OpenAI-compatible adapter)
- **Built-in Tool Suite** — `run_command`, `view_file`, `edit_file`, `write_file`, `list_dir`, `grep_search`
- **Multi-Session REPL** — Create, list, and switch between independent chat sessions
- **Auto-Generated Titles** — LLM-powered 3–4 word session titles on first message
- **Persistent Memory** — Conversation context preserved across turns within a session
- **JSONL Trajectory Logging** — Full trace of every agent step in `.mobius/traces/`
- **TOML Configuration** — Declarative provider & model setup via `model_config.toml`

---

## Quick Start

### Prerequisites

- **Go 1.25+**
- An API key for at least one supported provider (DeepSeek, OpenAI, etc.)

### Setup

```bash
# Clone the repository
git clone https://github.com/your-username/mobius.git
cd mobius

# Configure your API key
echo "DEEPSEEK_API_KEY=your-key-here" > .env

# Run the interactive REPL
go run cmd/mobius/main.go
```

### One-Shot Mode

```bash
go run cmd/mobius/main.go "explain how goroutines work"
```

---

## REPL Commands

| Command       | Description                              |
| ------------- | ---------------------------------------- |
| `/newchat`    | Create a new chat session                |
| `/listchats`  | List all sessions (active marked with *) |
| `exit`        | Exit Mobius                              |

---

## Project Structure

```
mobius/
├── cmd/mobius/main.go          # CLI entrypoint
├── pkg/
│   ├── agent/                  # ReAct agent loop & configuration
│   ├── llm/                    # LLM provider adapters & types
│   ├── tools/                  # Tool registry & implementations
│   ├── context/                # Conversation history & prompt templates
│   ├── session/                # Multi-session manager
│   ├── tracer/                 # JSONL trajectory logger
│   └── cli/                    # REPL loop & goal runner
├── model_config.toml           # Provider & model configuration
├── .env                        # API keys (gitignored)
└── .mobius/traces/             # Agent trajectory logs (gitignored)
```

---

## Configuration

Edit `model_config.toml` to set the default model and configure providers:

```toml
default_model = "deepseek-chat"

[providers.deepseek]
name     = "DeepSeek"
type     = "openai"
base_url = "https://api.deepseek.com/v1"
env_key  = "DEEPSEEK_API_KEY"
models   = ["deepseek-chat", "deepseek-reasoner"]
```

---

## License

MIT
