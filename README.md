**[English](README.md)** | [中文文档](README_ZH.md)

# Athena

AI Agent orchestration system that operates like a real IT company.

Athena spawns specialized agents (PM, Developer, Tester, Reviewer, Designer) that collaborate through a blackboard architecture — with structured prompts, multi-round verification loops, and escalation mechanisms.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  CEO (User)                                         │
│    │ POST /api/chat {"message": "build a snake game"}│
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

## Requirements

- Go 1.25+
- GCC (for CGO / go-sqlite3)
- An OpenAI-compatible LLM API (OpenAI, Azure, Tencent LKEAP, etc.)

## Install

```bash
git clone https://github.com/KSroido/athena.git
cd athena
go build -o athena ./cmd/athena
```

## Configure

Create `config/athena.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  base_url: "https://api.openai.com/v1"   # OpenAI-compatible endpoint
  api_key: "sk-..."                        # Your API key
  model: "gpt-4o"                          # Model with tool calling support

company:
  max_agents: 100       # Company headcount limit
  max_memory_mb: 16384

agents:
  data_dir: "./data"    # Project workspace + blackboard + agent memory

logging:
  level: "info"
  file: "./data/logs/athena.log"
```

Alternatively, configure via environment variables:

```bash
export ATHENA_LLM_BASE_URL="https://api.openai.com/v1"
export ATHENA_LLM_API_KEY="sk-..."
export ATHENA_LLM_MODEL="gpt-4o"
export ATHENA_PORT=8080
```

## Run

```bash
./athena -config config/athena.yaml
```

Output:

```
=====================================
  Athena — AI Agent Orchestration
  Runs like a real IT company
=====================================
Athena server starting on 0.0.0.0:8080
Frontend: http://localhost:8080
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

# Create project manually (bypasses AgentServer)
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Snake Game", "original_requirement": "Build a snake game", "priority": 5}'

# Get project details
curl http://localhost:8080/api/projects/{id}
```

### Blackboard

```bash
# Read blackboard entries (all categories)
curl http://localhost:8080/api/projects/{id}/blackboard

# Read specific category
curl "http://localhost:8080/api/projects/{id}/blackboard?category=verification"

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

Each role receives a specific set of tools:

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

The PM verification process follows this state machine:

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

- Each round is tracked in blackboard (`category: verification`)
- Round ≥80: warning injected into PM steer message
- Round ≥100: PM must write `[ESCALATION]` and stop the loop
- CEO can then decide to: relax criteria, change developer, or cancel project

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
│   │   ├── llm_client.go        # Eino OpenAI-compatible client
│   │   └── prompts.go           # 6-layer structured prompts
│   ├── db/
│   │   ├── database.go          # Main SQLite DB
│   │   └── models.go            # Data models
│   ├── hr/hr.go                 # Hiring + role templates
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

## LLM Compatibility

Athena uses Eino's OpenAI-compatible client. Any provider supporting the OpenAI Chat Completions API with **tool calling** works:

| Provider | base_url | Notes |
|----------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{model}` | |
| Tencent LKEAP | `https://api.lkeap.cloud.tencent.com/plan/v3` | glm-5.1 |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat |
| Local (Ollama) | `http://localhost:11434/v1` | Must support tool calling |
| Local (vLLM) | `http://localhost:8000/v1` | |

**Requirement:** The model must support OpenAI-style function/tool calling. Models without tool calling support will not work.

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

MIT
