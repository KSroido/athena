**[English](README.md)** | [中文文档](README_ZH.md)

# Athena

像真实IT公司一样运作的 AI Agent 编排系统。

Athena 启动专业化 Agent（PM、Developer、Tester、Reviewer、Designer），通过黑板架构协作——具备结构化提示词、多轮验收循环和升级上报机制。

## 架构

```
┌─────────────────────────────────────────────────────┐
│  CEO (用户)                                         │
│    │ POST /api/chat {"message": "写一个贪吃蛇游戏"}  │
│    ▼                                                │
│  AgentServer (CEO秘书)                              │
│    │ 意图识别: new_project / update / query / HR     │
│    ▼                                                │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │ PM Agent │───▶│ Developer│───▶│ Blackboard│     │
│  │          │◀───│  Agent   │    │ (SQLite)  │     │
│  │ 验收 ←───┘    │          │    │           │     │
│  │  ↑ submit_    └──────────┘    │  goal     │     │
│  │  │ for_review                 │  fact     │     │
│  │  │ (SteerCh 通知)             │  criteria │     │
│  │  └────────────────────────────│  verify   │     │
│  └──────────┐                    └──────────┘      │
│             │ hr_request                            │
│             ▼                                       │
│  ┌──────────┐                                       │
│  │    HR    │ ← 角色模板，公司规模上限               │
│  └──────────┘                                       │
└─────────────────────────────────────────────────────┘
```

**核心工作流：**

1. CEO 发需求 → AgentServer 意图识别 → 创建项目 + 黑板
2. HR 招聘 PM → PM 定义验收标准 → 向 Developer 分配任务
3. Developer 编码 → 使用 `submit_for_review` → 通过 SteerCh 唤醒 PM
4. PM 读取产出文件，逐条对照验收标准
5. 不通过 → PM 发送整改任务 → 回到步骤 3
6. 通过 → PM 写入 `[PASS]` 到黑板
7. ≥100 轮 → PM 写入 `[ESCALATION]` → CEO 决策

## Agent Soul（6层提示词架构）

每个 Agent 的系统提示词遵循结构化的6层架构：

| 层级 | 内容 | 示例 |
|------|------|------|
| 1. 身份 | 我是谁，哪个项目 | "你是 Athena 系统中的项目经理 Agent" |
| 2. 原则 | 核心行为准则 | "需求回溯：每轮验收必须对照CEO原始需求逐条确认" |
| 3. 流程 | 标准操作步骤 | "1.读取黑板→2.定义验收标准→3.招聘→4.分配任务→5.验收循环" |
| 4. 工具 | 何时使用哪个工具 | "submit_for_review: 完成开发后必须使用此工具提交验收" |
| 5. 约束 | 什么不能做 | "禁止未经 file_read 读取文件就判定验收通过" |
| 6. 自检 | 完成前检查清单 | "是否逐条对照了CEO原始需求？" |

## 环境要求

- Go 1.25+
- GCC（CGO / go-sqlite3 编译需要）
- OpenAI 兼容的 LLM API（OpenAI、Azure、腾讯云 LKEAP 等）

## 安装

```bash
git clone https://github.com/KSroido/athena.git
cd athena
go build -o athena ./cmd/athena
```

> **注意：** go-sqlite3 需要 CGO，确保系统安装了 GCC。
> Linux: `sudo apt install build-essential`
> macOS: `xcode-select --install`

## 配置

创建 `config/athena.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  base_url: "https://api.openai.com/v1"   # OpenAI 兼容端点
  api_key: "sk-..."                        # API Key
  model: "gpt-4o"                          # 必须支持 tool calling

company:
  max_agents: 100       # 公司人数上限
  max_memory_mb: 16384

agents:
  data_dir: "./data"    # 项目工作区 + 黑板 + Agent记忆

logging:
  level: "info"
  file: "./data/logs/athena.log"
```

也可通过环境变量配置：

```bash
export ATHENA_LLM_BASE_URL="https://api.openai.com/v1"
export ATHENA_LLM_API_KEY="sk-..."
export ATHENA_LLM_MODEL="gpt-4o"
export ATHENA_PORT=8080
```

## 启动

```bash
./athena -config config/athena.yaml
```

输出：

```
=====================================
  Athena — AI Agent 编排系统
  像IT公司一样运作
=====================================
Athena server starting on 0.0.0.0:8080
Frontend: http://localhost:8080
```

## API 参考

### CEO 对话

```bash
# 创建新项目（自动识别为 new_project 意图）
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "写一个纯前端的贪吃蛇web小游戏，支持方向键控制"}'

# 查询项目进展
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "项目进展怎么样了"}'

# 向已有项目追加新需求
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "给贪吃蛇加上排行榜功能"}'
```

### 项目管理

```bash
# 列出所有项目
curl http://localhost:8080/api/projects

# 手动创建项目（绕过 AgentServer）
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Snake Game", "original_requirement": "写一个贪吃蛇游戏", "priority": 5}'

# 获取项目详情
curl http://localhost:8080/api/projects/{id}
```

### 黑板

```bash
# 读取黑板条目（全部类别）
curl http://localhost:8080/api/projects/{id}/blackboard

# 按类别筛选
curl "http://localhost:8080/api/projects/{id}/blackboard?category=verification"

# 写入黑板
curl -X POST http://localhost:8080/api/projects/{id}/blackboard \
  -H "Content-Type: application/json" \
  -d '{"category": "fact", "content": "API接口使用RESTful风格", "certainty": "certain", "author": "ceo"}'
```

**黑板类别：**

| 类别 | 说明 | 写入者 |
|------|------|--------|
| `goal` | CEO 原始需求 | CEO, PM |
| `fact` | 确认的事实 | 全部 Agent |
| `acceptance_criteria` | PM 定义的验收标准 | PM |
| `verification` | 验收轮次结果 | Developer（提交）、PM（审核） |
| `progress` | 工作进展 | 全部 Agent |
| `discovery` | 新发现 | 全部 Agent |
| `resolution` | 会议决议 | 全部 Agent |
| `auxiliary` | 错误日志、诊断信息 | Developer, Tester |
| `decision` | 关键决策 | PM |

### Agent

```bash
# 列出运行中的 Agent
curl http://localhost:8080/api/agents

# 列出项目下的 Agent
curl http://localhost:8080/api/projects/{id}/agents
```

### 公司 / HR

```bash
# 列出所有公司成员
curl http://localhost:8080/api/company

# 手动招聘 Agent
curl -X POST http://localhost:8080/api/company/hire \
  -H "Content-Type: application/json" \
  -d '{"role": "developer", "project_id": "abc123", "reason": "需要开发工程师"}'
```

**可用角色：** `pm`、`developer`、`tester`、`reviewer`、`designer`

## Agent 工具集

每个角色拥有特定的工具集：

| 工具 | PM | Developer | Tester | Reviewer | Designer |
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

## 验收循环

PM 验收流程的状态机：

```
                  ┌─────────────────────┐
                  │  PM 定义             │
                  │  acceptance_criteria │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  PM 分配任务        │
                  │  给 Developer       │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  Developer 编码     │
                  │  + submit_for_review│
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
              ┌──▶│  PM 验收            │
              │   │  (读文件,           │
              │   │   对照标准)         │
              │   └──────┬─────────────┘
              │          │
              │    ┌─────▼─────┐
              │    │  通过?    │
              │    └──┬────┬───┘
              │       │    │
              │    否  │    │ 是
              │  ┌─────▼┐  ┌▼──────────┐
              │  │整改 + │  │[PASS]     │
              │  │重提   │  │→ 交付     │
              │  └──┬────┘  └───────────┘
              │     │
              │  ┌──▼────────────┐
              └──│轮次 < 100?    │
                 └──┬────────┬───┘
                    │        │
                 是 │     否 │
                    │   ┌────▼─────────┐
                    │   │[ESCALATION]  │
                    │   │→ CEO 决策    │
                    │   └──────────────┘
                    │
                    └──▶ 回到 PM 验收
```

- 每轮验收记录在黑板中（`category: verification`）
- 轮次 ≥80：PM steer 消息中注入预警
- 轮次 ≥100：PM 必须写入 `[ESCALATION]` 并停止验收循环
- CEO 随后可决定：放宽标准、更换 Developer、或取消项目

## 项目结构

```
athena/
├── cmd/athena/main.go           # 入口
├── config/athena.yaml           # 配置文件
├── internal/
│   ├── api/handlers.go          # HTTP 路由处理 (Gin)
│   ├── blackboard/
│   │   ├── board.go             # 黑板 SQLite + FTS5
│   │   └── access_control.go    # 角色-层级访问控制矩阵
│   ├── config/config.go         # 配置加载 + 校验
│   ├── core/
│   │   ├── agent_loop.go        # Agent ReAct 循环 + 工具创建
│   │   ├── agent_loop_v2.go     # 进程内运行 + 验收通知处理
│   │   ├── agent_manager.go     # goroutine Agent 生命周期管理
│   │   ├── agent_server.go      # CEO 秘书（意图路由）
│   │   ├── llm_client.go        # Eino OpenAI 兼容客户端
│   │   └── prompts.go           # 6层结构化提示词
│   ├── db/
│   │   ├── database.go          # 主 SQLite 数据库
│   │   └── models.go            # 数据模型
│   ├── hr/hr.go                 # 招聘 + 角色模板
│   ├── server/server.go         # Gin HTTP 服务器
│   └── tools/
│       ├── eino_tools.go        # 黑板读写工具
│       └── tools_v2.go          # assign_task, hr_request, file, term, submit_for_review
└── data/                        # 运行时数据（gitignore）
    ├── athena.sqlite            # 主数据库
    ├── board/                   # 按项目隔离的黑板数据库
    ├── workspace/{project_id}/  # 项目文件产出
    └── agents/{agent_id}/       # Agent 个人记忆 (memory.md)
```

## LLM 兼容性

Athena 使用 Eino 的 OpenAI 兼容客户端。任何支持 OpenAI Chat Completions API 且具备 **tool calling** 能力的提供商均可使用：

| 提供商 | base_url | 备注 |
|--------|----------|------|
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{model}` | |
| 腾讯云 LKEAP | `https://api.lkeap.cloud.tencent.com/plan/v3` | glm-5.1 |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat |
| 本地 (Ollama) | `http://localhost:11434/v1` | 需支持 tool calling |
| 本地 (vLLM) | `http://localhost:8000/v1` | |

**硬性要求：** 模型必须支持 OpenAI 风格的 function/tool calling。不支持 tool calling 的模型无法工作。

## 开发

```bash
# 编译（需要 CGO）
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go build -o athena ./cmd/athena

# 测试
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go test ./...

# 热重载（需要 air）
air
```

## License

MIT
