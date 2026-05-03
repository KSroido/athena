**[English](README.md)** | [中文文档](README_ZH.md)

<div align="center">

# Athena

**Multi-Agent Orchestration System that runs like a real IT company**

[![GitHub stars](https://img.shields.io/github/stars/KSroido/athena?style=social)](https://github.com/KSroido/athena/stargazers)
[![GitHub license](https://img.shields.io/github/license/KSroido/athena)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksroido/athena)](https://goreportcard.com/report/github.com/ksroido/athena)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)

📖 [Architecture](#architecture) · 🚀 [Quick Start](#quick-start) · 🔧 [Configuration](#configure) · 📡 [API Reference](#api-reference) · 🤝 [Contributing](#contributing)

</div>

---

> **Athena** spawns specialized agents — PM, Developer, Tester, Reviewer, Designer — that collaborate through a blackboard architecture with structured prompts, multi-round verification loops, and escalation mechanisms.

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=KSroido/athena&type=Date)](https://star-history.com/#KSroido/athena&Date)

## 🌟 Why Athena?

| Feature | Description |
|---------|-------------|
| **IT Company Model** | Agents act like real roles: PM defines requirements, Developer codes, Tester verifies, Reviewer audits |
| **Blackboard Architecture** | Shared memory with structured categories (goal, fact, criteria, verification, decision) |
| **Verification Loop** | PM reads output files and checks against acceptance criteria line-by-line — no rubber-stamping |
| **6-Layer Agent Soul** | Structured prompts: Identity → Principles → Workflow → Tools → Constraints → SelfCheck |
| **Multi-Provider LLM** | Built-in fallback chain with 429 detection, automatic cooldown, and provider rotation |
| **Dynamic HR** | Hire any role on demand — HR generates specialized souls via LLM, with project-aware fitness checks |
| **100-Round Escalation** | Verification loops capped at 100 rounds; PM escalates to CEO for decision |
| **Go + SQLite** | Single binary, zero external dependencies. FTS5-powered blackboard search |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  CEO (User)                                         │
│    │ POST /api/chat {"message": "build snake game"} │
│    ▼                                                │
│  AgentServer (CEO Secretary)                        │
│    │ Intent: new_project / update / query / HR      │
│    ▼                                                │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │ PM Agent │───▶│ Developer│───▶│ Blackboard│     │
│  │          │◀───│  Agent   │    │ (SQLite)  │     │
│  │ verify ←─┘    │          │    │           │     │
│  │  ↑ submit_    └──────────┘    │  goal     │     │
│  │  │ for_review                 │  fact     │     │
│  │  │ (SteerCh notify)           │  criteria │     │
│  │  └────────────────────────────│  verify   │     │
│  └──────────┐                    └──────────┘      │
│             │ hr_request                            │
│             ▼                                       │
│  ┌──────────┐                                       │
│  │    HR    │ ← Role templates, company size limit   │
│  └──────────┘                                       │
└─────────────────────────────────────────────────────┘
```

**Core workflow:**

1. CEO sends requirement → AgentServer recognizes intent → creates project + blackboard
2. HR hires PM → PM defines acceptance criteria → assigns task to Developer
3. Developer codes → uses `submit_for_review` → PM is woken up via SteerCh
4. PM reads output files, checks against criteria line-by-line
5. Not passed → PM sends corrective task → loop back to step 3
6. Passed → PM writes `[PASS]` to blackboard
7. ≥100 rounds → PM writes `[ESCALATION]` → CEO decides

## Agent Soul (6-Layer Prompt)

Each agent's system prompt follows a structured 6-layer architecture:

| Layer | Content | Example |
|-------|---------|---------|
| 1. Identity | Who I am, which project | "You are the PM Agent in the Athena system" |
| 2. Principles | Core behavioral rules | "Requirement traceability: every verification round must check against CEO's original requirements item by item" |
| 3. Workflow | Step-by-step SOP | "1. Read blackboard → 2. Define acceptance criteria → 3. Hire → 4. Assign task → 5. Verification loop" |
| 4. Tools | When to use which | "submit_for_review: must use this tool after completing development to submit for verification" |
| 5. Constraints | What I cannot do | "Never mark verification as passed without reading the output file with file_read" |
| 6. SelfCheck | Checklist before finish | "Have I checked against each item of the CEO's original requirements?" |

## Quick Start

### Requirements

- Go 1.25+
- GCC (for CGO / go-sqlite3)
- An OpenAI-compatible LLM API

### Install

```bash
git clone https://github.com/KSroido/athena.git
cd athena
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go build -o athena ./cmd/athena
```

### Configure

Create `config/athena.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  max_retries: 3
  retry_cooldown: 30
  providers:
    - base_url: "https://api.openai.com/v1"
      api_key: "sk-..."
      model: "gpt-4o"
      weight: 100
    # Fallback provider (optional)
    - base_url: "https://api.deepseek.com/v1"
      api_key: "sk-..."
      model: "deepseek-chat"
      weight: 50

company:
  max_agents: 100
  max_memory_mb: 16384

agents:
  data_dir: "./data"
```

Environment variables also supported:

```bash
export ATHENA_LLM_BASE_URL="https://api.openai.com/v1"
export ATHENA_LLM_API_KEY="sk-..."
export ATHENA_LLM_MODEL="gpt-4o"
```

### Run

```bash
./athena -config config/athena.yaml
```

Output:

```
=====================================
  Athena — AI Agent Orchestration
  Runs like a real IT company
=====================================
[llm] initialized provider: https://api.openai.com/v1/gpt-4o (weight=100)
Athena server starting on 0.0.0.0:8080
Frontend: http://localhost:8080
```

### Try It

```bash
# Create a project
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Build a frontend-only snake game with arrow key controls"}'
```

## API Reference

### CEO Chat

```bash
# Create a new project (auto-detected as new_project intent)
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Build a frontend-only snake game with arrow key controls"}'

# Query project status
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "How is the project going?"}'

# Send new requirement to existing project
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Add a leaderboard feature to the snake game"}'
```

### Projects

```bash
# List all projects
curl http://localhost:8080/api/projects

# Create project manually
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Snake Game", "original_requirement": "Build a snake game", "priority": 5}'

# Get project details
curl http://localhost:8080/api/projects/{id}
```

### Blackboard

```bash
# Read blackboard entries
curl http://localhost:8080/api/projects/{id}/blackboard

# Write to blackboard
curl -X POST http://localhost:8080/api/projects/{id}/blackboard \
  -H "Content-Type: application/json" \
  -d '{"category": "fact", "content": "API uses RESTful style", "certainty": "certain", "author": "ceo"}'
```

**Blackboard categories:**

| Category | Description | Who Writes |
|----------|-------------|------------|
| `goal` | CEO original requirements | CEO, PM |
| `fact` | Confirmed facts | All agents |
| `acceptance_criteria` | PM-defined verification criteria | PM |
| `verification` | Verification round results | Developer (submit), PM (review) |
| `progress` | Work progress updates | All agents |
| `discovery` | New findings | All agents |
| `resolution` | Meeting resolutions | All agents |
| `auxiliary` | Error logs, diagnostics | Developer, Tester |
| `decision` | Key decisions | PM |

### Agents

```bash
# List running agents
curl http://localhost:8080/api/agents

# List agents for a project
curl http://localhost:8080/api/projects/{id}/agents
```

### Company / HR

```bash
# List all company members
curl http://localhost:8080/api/company

# Hire an agent manually
curl -X POST http://localhost:8080/api/company/hire \
  -H "Content-Type: application/json" \
  -d '{"role": "developer", "project_id": "abc123", "reason": "Need a developer"}'
```

**Available roles:** `pm`, `developer`, `tester`, `reviewer`, `designer`

## Agent Tools

| Tool | PM | Developer | Tester | Reviewer | Designer |
|------|:--:|:---------:|:------:|:--------:|:--------:|
| `blackboard_read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `blackboard_write` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `memory_read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `memory_write` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `meeting` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `assign_task` | ✅ | | | | |
| `hr_request` | ✅ | | | | |
| `file_read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `file_write` | ✅ | ✅ | ✅ | | ✅ |
| `term` | | ✅ | ✅ | | |
| `submit_for_review` | | ✅ | | | |

## Verification Loop

```
                  ┌─────────────────────┐
                  │  PM defines         │
                  │  acceptance_criteria │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  PM assigns task    │
                  │  to Developer       │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  Developer codes    │
                  │  + submit_for_review│
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
              ┌──▶│  PM verifies        │
              │   │  (reads files,      │
              │   │   checks criteria)  │
              │   └──────┬─────────────┘
              │          │
              │    ┌─────▼─────┐
              │    │  Passed?  │
              │    └──┬────┬───┘
              │       │    │
              │    No  │    │ Yes
              │  ┌─────▼┐  ┌▼──────────┐
              │  │Fix +  │  │[PASS]     │
              │  │resub  │  │→ Deliver  │
              │  └──┬────┘  └───────────┘
              │     │
              │  ┌──▼────────────┐
              └──│round < 100?   │
                 └──┬────────┬───┘
                    │        │
                 Yes│     No │
                    │   ┌────▼─────────┐
                    │   │[ESCALATION]  │
                    │   │→ CEO decides │
                    │   └──────────────┘
                    │
                    └──▶ loop back to PM verifies
```

## LLM Compatibility

Athena uses Eino's OpenAI-compatible client with **multi-provider fallback**. Any provider supporting OpenAI Chat Completions API with tool calling works:

| Provider | base_url | Notes |
|----------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{model}` | |
| Volcengine Ark | `https://ark.cn-beijing.volces.com/api/v3` | Doubao, GLM |
| Tencent LKEAP | `https://api.lkeap.cloud.tencent.com/plan/v3` | glm-5.1 |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat |
| Local (Ollama) | `http://localhost:11434/v1` | Must support tool calling |
| Local (vLLM) | `http://localhost:8000/v1` | |

**Requirement:** The model must support OpenAI-style function/tool calling.

**Fallback features:**
- 429/rate-limit detection with per-provider cooldown
- Automatic provider rotation on failure
- Configurable retry count and cooldown duration
- Retry-After header parsing

## Project Structure

```
athena/
├── cmd/athena/main.go           # Entry point
├── config/athena.yaml           # Configuration
├── internal/
│   ├── api/handlers.go          # HTTP handlers (Gin)
│   ├── blackboard/
│   │   ├── board.go             # Blackboard SQLite + FTS5
│   │   └── access_control.go    # Role-level access matrix
│   ├── config/config.go         # Config loading + validation
│   ├── core/
│   │   ├── agent_loop.go        # Agent ReAct loop + tool creation
│   │   ├── agent_loop_v2.go     # RunInProcess + verification steer
│   │   ├── agent_manager.go     # Goroutine-based agent lifecycle
│   │   ├── agent_server.go      # CEO Secretary (intent routing)
│   │   ├── llm_client.go        # Multi-provider LLM with 429 fallback
│   │   └── prompts.go           # 6-layer structured prompts
│   ├── db/
│   │   ├── database.go          # Main SQLite DB
│   │   └── models.go            # Data models
│   ├── hr/hr.go                 # Hiring + role templates + fitness check
│   ├── server/server.go         # Gin HTTP server
│   └── tools/
│       ├── eino_tools.go        # Blackboard read/write tools
│       └── tools_v2.go          # assign_task, hr_request, file, term, submit_for_review
└── data/                        # Runtime data (gitignored)
    ├── athena.sqlite            # Main DB
    ├── board/                   # Per-project blackboard DBs
    ├── workspace/{project_id}/  # Project file output
    └── agents/{agent_id}/       # Agent memory (memory.md)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

```bash
# Build
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go build -o athena ./cmd/athena

# Test
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go test ./...

# Run with hot reload (requires air)
air
```

## License

[MIT](LICENSE)
