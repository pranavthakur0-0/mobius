# Contributing to Mobius

Thank you for your interest in contributing to **Mobius**! We welcome community contributions to help build the fastest, most reliable AI agent harness in Go.

---

## 🛠️ Development Setup

### 1. Prerequisites
* **Go 1.24+**
* `git`

### 2. Fork & Clone
```bash
git clone https://github.com/pranavthakur0-0/mobius.git
cd mobius
```

### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your LLM API keys
```

---

## 🧪 Testing & Code Standards

Before opening a pull request, ensure all tests pass and your code adheres to standard Go conventions:

```bash
# Run all unit tests
go test -v ./...

# Format all Go files
go fmt ./...

# Verify that binary builds cleanly
go build ./cmd/mobius
```

### Coding Guidelines:
1. **Standard Library First**: Minimize external third-party dependencies wherever possible.
2. **Package Separation**:
   * `pkg/agent`: ReAct loop and execution engine.
   * `pkg/agentctx`: Token estimation, prompt templates, checkpoint summarizer, and compactor.
   * `pkg/artifact`: Disk-backed artifact offloading and truncation interceptor.
   * `pkg/budget`: Token metering and spend limits.
   * `pkg/events`: Telemetry and JSONL event store.
   * `pkg/llm`: Provider adapters and multi-provider interfaces.
   * `pkg/tools`: Standardized agent tool implementations.
3. **Idiomatic Go**: Use `camelCase` for internal variables/functions, `PascalCase` for exported identifiers. Avoid `snake_case`.

---

## 🔀 Pull Request Process

1. Create a descriptive feature branch: `git checkout -b feat/your-feature-name`
2. Write clean code with appropriate unit tests.
3. Commit with standard conventional commit messages (e.g. `feat: add docker sandbox isolation`, `fix: handle nil tool call ID`).
4. Push your branch and open a Pull Request against `main`.

---

## 💬 Community & Support

If you encounter bugs, have feature requests, or want to discuss architecture, feel free to open a [GitHub Issue](https://github.com/pranavthakur0-0/mobius/issues).
