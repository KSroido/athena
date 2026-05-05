**[中文文档](README_ZH.md)** | [English](README.md)

<div align="center">

# Athena

**像真实IT公司一样运作的多智能体编排系统**

[![GitHub stars](https://img.shields.io/github/stars/KSroido/athena?style=social)](https://github.com/KSroido/athena/stargazers)
[![GitHub license](https://img.shields.io/github/license/KSroido/athena)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksroido/athena)](https://goreportcard.com/report/github.com/ksroido/athena)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)

📖 [架构](#架构) · 🚀 [快速开始](#快速开始) · 🔧 [配置](#配置) · 📡 [API 参考](#api-参考) · 🤝 [贡献](#贡献)

</div>

---

> **Athena** 生成专业化智能体 — 项目经理、开发工程师、测试工程师、代码审查员、设计师 — 通过黑板架构协作，配合结构化提示词、多轮验收循环和升级机制。

## ⭐ Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=KSroido/athena&type=Date)](https://star-history.com/#KSroido/athena&Date)

## 🌟 为什么选择 Athena？

| 特性 | 说明 |
|------|------|
| **IT公司模型** | 智能体扮演真实角色：PM定义需求，开发编码，测试验证，审查审计 |
| **黑板架构** | 共享记忆，结构化分类（目标、事实、标准、验证、决策） |
| **验收循环** | PM读取输出文件，逐条对照验收标准 — 不走形式 |
| **6层Agent Soul** | 结构化提示词：身份→原则→流程→工具→约束→自检 |
| **多Provider LLM** | 内置回退链，429检测，自动冷却，Provider轮换 |
| **动态HR** | 按需招聘任意角色 — HR通过LLM生成专业Soul，带项目适配性检查 |
| **100轮升级机制** | 验收循环上限100轮；PM升级至CEO决策 |
| **Go + SQLite** | 单二进制，零外部依赖。FTS5驱动黑板搜索 |

## 架构

```
┌─────────────────────────────────────────────────────┐
│  CEO (用户)                                          │
│    │ POST /api/chat {"message": "做个贪吃蛇游戏"}    │
│    ▼                                                │
│  AgentServer (CEO秘书)                               │
│    │ 意图: new_project / update / query / HR         │
│    ▼                                                │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │ PM Agent │───▶│ Developer│───▶│ Blackboard│     │
│  │          │◀───│  Agent   │    │ (SQLite)  │     │
│  │ verify ←─┘    │          │    │           │     │
│  │  ↑ submit_    └──────────┘    │  goal     │     │
│  │  │ for_review                 │  fact     │     │
│  │  │ (SteerCh通知)              │  criteria │     │
│  │  └────────────────────────────│  verify   │     │
│  └──────────┐                    └──────────┘      │
│             │ hr_request                            │
│             ▼                                       │
│  ┌──────────┐                                       │
│  │    HR    │ ← 角色模板, 公司人数上限               │
│  └──────────┘                                       │
└─────────────────────────────────────────────────────┘
```

**核心工作流：**

1. CEO发送需求 → AgentServer识别意图 → 创建项目+黑板
2. HR招聘PM → PM定义验收标准 → 分配任务给开发
3. 开发编码 → 使用`submit_for_review` → PM通过SteerCh被唤醒
4. PM读取输出文件，逐条对照标准
5. 未通过 → PM发送修正任务 → 回到步骤3
6. 通过 → PM向黑板写入`[PASS]`
7. ≥100轮 → PM写入`[ESCALATION]` → CEO决策

## Agent Soul (6层提示词)

每个智能体的系统提示词遵循6层结构化架构：

| 层级 | 内容 | 示例 |
|------|------|------|
| 1. 身份 | 我是谁，哪个项目 | "你是Athena系统的PM Agent" |
| 2. 原则 | 核心行为规则 | "需求可追溯：每轮验收必须逐条对照CEO原始需求" |
| 3. 流程 | 逐步SOP | "1. 读黑板→2. 定义验收标准→3. 招聘→4. 分配任务→5. 验收循环" |
| 4. 工具 | 何时用哪个 | "submit_for_review：开发完成后必须使用此工具提交验收" |
| 5. 约束 | 不能做什么 | "不读输出文件不得标记验收通过" |
| 6. 自检 | 完成前清单 | "是否逐条对照了CEO原始需求？" |

## 快速开始

### 环境要求

- Go 1.25+
- GCC（用于CGO / go-sqlite3）
- 兼容OpenAI的LLM API

### 安装

```bash
git clone https://github.com/KSroido/athena.git
cd athena
./install.sh
```

`install.sh` 会完成完整的本地安装流程：

| 步骤 | 行为 |
|------|------|
| 构建 | 使用 `CGO_CFLAGS=-DSQLITE_ENABLE_FTS5`、`CGO_LDFLAGS=-lm`、`GOTOOLCHAIN=auto` 编译 `./cmd/athena` |
| 安装二进制 | 默认安装到 `~/.local/bin/athena` |
| 创建配置 | 当目标配置不存在时，将 `config/athena.example.yaml` 复制到 `~/.config/athena/athena.yaml` |
| 加入 PATH | 把安装目录追加到 shell 启动文件（`~/.bashrc`、`~/.zshrc` 或 `~/.config/fish/config.fish`） |
| 输出运行命令 | 打印可直接执行的 `athena -config ...` 命令 |

安装后，打开新 shell，或手动执行：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

然后验证：

```bash
athena -h
athena -config ~/.config/athena/athena.yaml
```

安装选项：

| 选项 | 说明 |
|------|------|
| `--dir DIR` | 安装二进制到 `DIR`，替代默认 `~/.local/bin` |
| `--config-dir DIR` | 在 `DIR` 下创建/读取配置，替代默认 `~/.config/athena` |
| `--no-config` | 跳过配置文件创建 |
| `--help` | 显示安装脚本帮助 |

示例：

```bash
./install.sh --dir ~/.local/bin --config-dir ~/.config/athena
ATHENA_INSTALL_DIR=/usr/local/bin ./install.sh --no-config
make install
```

仅手动构建，不安装：

```bash
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go build -o athena ./cmd/athena
```

### 配置

创建 `config/athena.yaml`：

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
    # 备用Provider（可选）
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

也支持环境变量：

```bash
export ATHENA_LLM_BASE_URL="https://api.openai.com/v1"
export ATHENA_LLM_API_KEY="sk-..."
export ATHENA_LLM_MODEL="gpt-4o"
```

### 运行

```bash
./athena -config config/athena.yaml
```

输出：

```
=====================================
  Athena — AI Agent 编排系统
  像IT公司一样运作
=====================================
[llm] initialized provider: https://api.openai.com/v1/gpt-4o (weight=100)
Athena server starting on 0.0.0.0:8080
Frontend: http://localhost:8080
```

### 试用

```bash
# 创建项目
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "做一个前端贪吃蛇游戏，方向键控制"}'
```

## API 参考

### CEO 聊天

```bash
# 创建新项目（自动识别为new_project意图）
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "做一个前端贪吃蛇游戏"}'

# 查询项目状态
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "项目进展如何？"}'

# 向现有项目发送新需求
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "给贪吃蛇加排行榜"}'
```

### 项目

```bash
# 列出所有项目
curl http://localhost:8080/api/projects

# 手动创建项目
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "贪吃蛇", "original_requirement": "做一个贪吃蛇游戏", "priority": 5}'

# 获取项目详情
curl http://localhost:8080/api/projects/{id}
```

### 黑板

```bash
# 读取黑板条目
curl http://localhost:8080/api/projects/{id}/blackboard

# 写入黑板
curl -X POST http://localhost:8080/api/projects/{id}/blackboard \
  -H "Content-Type: application/json" \
  -d '{"category": "fact", "content": "API使用RESTful风格", "certainty": "certain", "author": "ceo"}'
```

**黑板分类：**

| 分类 | 说明 | 写入者 |
|------|------|--------|
| `goal` | CEO原始需求 | CEO, PM |
| `fact` | 确认事实 | 所有智能体 |
| `acceptance_criteria` | PM定义的验收标准 | PM |
| `verification` | 验收轮次结果 | Developer（提交）, PM（审查） |
| `progress` | 工作进展 | 所有智能体 |
| `discovery` | 新发现 | 所有智能体 |
| `resolution` | 会议决议 | 所有智能体 |
| `auxiliary` | 错误日志、诊断 | Developer, Tester |
| `decision` | 关键决策 | PM |

### 智能体

```bash
# 列出运行中的智能体
curl http://localhost:8080/api/agents

# 列出项目的智能体
curl http://localhost:8080/api/projects/{id}/agents
```

### 公司 / HR

```bash
# 列出所有公司成员
curl http://localhost:8080/api/company

# 手动招聘智能体
curl -X POST http://localhost:8080/api/company/hire \
  -H "Content-Type: application/json" \
  -d '{"role": "developer", "project_id": "abc123", "reason": "需要开发工程师"}'
```

**可用角色：** `pm`, `developer`, `tester`, `reviewer`, `designer`

## 智能体工具

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

```
                  ┌─────────────────────┐
                  │  PM 定义             │
                  │  acceptance_criteria │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  PM 分配任务         │
                  │  给 Developer        │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
                  │  Developer 编码      │
                  │  + submit_for_review │
                  └──────────┬──────────┘
                             │
                  ┌──────────▼──────────┐
              ┌──▶│  PM 验收            │
              │   │  (读文件,           │
              │   │   对照标准)         │
              │   └──────┬─────────────┘
              │          │
              │    ┌─────▼─────┐
              │    │  通过？    │
              │    └──┬────┬───┘
              │       │    │
              │    否  │    │ 是
              │  ┌─────▼┐  ┌▼──────────┐
              │  │修正+  │  │[PASS]     │
              │  │重新   │  │→ 交付     │
              │  │提交   │  └───────────┘
              │  └──┬────┘
              │     │
              │  ┌──▼────────────┐
              └──│轮次 < 100？    │
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

## LLM 兼容性

Athena使用Eino的兼容OpenAI客户端，支持**多Provider回退**。任何支持OpenAI Chat Completions API + tool calling的Provider均可：

| Provider | base_url | 备注 |
|----------|----------|------|
| OpenAI | `https://api.openai.com/v1` | gpt-4o, gpt-4o-mini |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{model}` | |
| 字节火山 Ark | `https://ark.cn-beijing.volces.com/api/v3` | Doubao, GLM |
| 腾讯云 LKEAP | `https://api.lkeap.cloud.tencent.com/plan/v3` | glm-5.1 |
| DeepSeek | `https://api.deepseek.com/v1` | deepseek-chat |
| 本地 (Ollama) | `http://localhost:11434/v1` | 需支持tool calling |
| 本地 (vLLM) | `http://localhost:8000/v1` | |

**要求：** 模型必须支持OpenAI风格的function/tool calling。

**回退特性：**
- 429/限流检测，per-provider冷却
- 失败时自动Provider轮换
- 可配置重试次数和冷却时长
- Retry-After头解析

## 项目结构

```
athena/
├── cmd/athena/main.go           # 入口
├── config/athena.yaml           # 配置
├── internal/
│   ├── api/handlers.go          # HTTP处理器 (Gin)
│   ├── blackboard/
│   │   ├── board.go             # 黑板 SQLite + FTS5
│   │   └── access_control.go    # 角色级访问矩阵
│   ├── config/config.go         # 配置加载+校验
│   ├── core/
│   │   ├── agent_loop.go        # Agent ReAct循环+工具创建
│   │   ├── agent_loop_v2.go     # RunInProcess + 验收引导
│   │   ├── agent_manager.go     # 基于Goroutine的Agent生命周期
│   │   ├── agent_server.go      # CEO秘书（意图路由）
│   │   ├── llm_client.go        # 多Provider LLM + 429回退
│   │   └── prompts.go           # 6层结构化提示词
│   ├── db/
│   │   ├── database.go          # 主SQLite DB
│   │   └── models.go            # 数据模型
│   ├── hr/hr.go                 # 招聘+角色模板+适配性检查
│   ├── server/server.go         # Gin HTTP服务器
│   └── tools/
│       ├── eino_tools.go        # 黑板读写工具
│       └── tools_v2.go          # assign_task, hr_request, file, term, submit_for_review
└── data/                        # 运行时数据（gitignored）
    ├── athena.sqlite            # 主DB
    ├── board/                   # 每项目黑板DB
    ├── workspace/{project_id}/  # 项目文件输出
    └── agents/{agent_id}/       # Agent记忆 (memory.md)
```

## 贡献

欢迎贡献！请随时提交Pull Request。

1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 发起Pull Request

## 开发

```bash
# 构建
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go build -o athena ./cmd/athena

# 测试
CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go test ./...

# 热重载运行（需要air）
air
```

## 许可证

[MIT](LICENSE)
