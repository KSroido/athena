# Athena (雅典娜) — AI Agent 公司化编排系统

> **项目代号**: Athena（雅典娜）— 希腊神话中智慧与战略女神
> **命名寓意**: 雅典娜是策略之神，她不靠蛮力而靠智慧与组织取胜。正如我们的系统——通过智能编排和组织一群专业 Agent 协同工作，而非让一个全能 Agent 独自承担一切。雅典娜引导英雄们各司其职，正如我们引导每个 Agent 在自己的专业领域发挥作用。

## 为什么叫 Athena？

| 维度 | 雅典娜的象征 | Athena 系统的对应 |
|------|------------|-----------------|
| 智慧与战略 | 统筹全局，制定策略 | AgentServer (CEO秘书) 统筹公司方向 |
| 引导英雄 | 指引奥德修斯等英雄各展所长 | HR Agent 招聘并引导专业 Agent |
| 工艺之神 | 纺织、造船等技艺的守护者 | 每个 Agent 专人专用，各有所长 |
| 猫头鹰之眼 | 洞察全局，看见本质 | 黑板系统让所有 Agent 共享项目全貌 |
| 从宙斯头中诞生 | 完整的智慧，一经诞生即具备能力 | Agent 被创建时即配备完整专业工具和上下文 |

---

## 一句话定义

**Athena 是一个像 IT 公司一样运作的 AI Agent 编排系统**——CEO（用户）开箱即用，通过 Web 界面下达项目需求，CEO秘书(AgentServer)接收并交办，系统自动招聘专业 Agent、分配任务、协调推进，每个 Agent 上下文隔离、专人专用，遇到问题立刻开会沟通对齐，确保软件工程每个环节都有专人负责，交付稳定产品。

### 初始公司架构

Athena 启动时只需要 3 个核心 Agent + 一套通用员工模板：

| 组件 | 职责 |
|------|------|
| **AgentServer (CEO秘书)** | 与CEO直接对话、获取项目、关键抉择时请示CEO。项目交给项目经理 Agent 拆解 |
| **HR Agent** | 感知能力缺口、按模板创建新 Agent、模板不足时组织小组编写新模板。招聘前检查公司规模上限 |
| **项目经理 Agent** | 从 AgentServer 接收项目、拆解需求、分配任务。需要招人时立刻和 HR 沟通（详细说明招什么样的人） |
| **通用员工模板 (soul.md 等)** | 定义各角色的 system prompt、工具集、黑板权限等 |

### 初始工作流

1. CEO 通过 Web 界面下达需求 → AgentServer (CEO秘书) 接收
2. AgentServer 将项目交给项目经理 Agent
3. 项目经理拆解需求 → 发现需要招人 → 立刻和 HR 沟通（详细说明招什么样的人）
4. 任何 Agent 认为需要招人都可以和 HR 沟通，但项目经理必须参加招聘会议，共同决定是否招人
5. HR 招聘新 Agent → 项目经理分配任务 → 各 Agent 独立工作
6. 任何 Agent 遇到别人领域的问题 → 立刻开会沟通

### HR 招聘的模板扩展

当现有模板不足以完成招聘任务时：
1. HR Agent 成立专门的小组
2. 相关任务放入公司黑板数据库 (BOARD.sqlite)
3. 小组编写新方向的 job description 和模板
4. 新模板审核通过后用于招聘新方向员工

---

## 核心设计原则

### 1. 专人专用
开发只做开发，测试只做测试，设计只做设计——绝不跨岗

### 2. 上下文隔离
每个 Agent 有独立的工作记忆和上下文窗口。部分记忆是共通的（公司总体任务等来自黑板），但剩下的记忆需要像现实员工那样各自管理，不互相干扰。

### 3. 黑板共享
项目目标、进展、确定性事实通过中央黑板共享给所有 Agent

### 4. 事实分级
确定性事实标记为"确定"（必须 100% 可靠），猜测标记为"猜测"

### 5. 自动招聘
HR Agent 感知能力缺口时自动创建新 Agent，按需配备工具和新 Agent 职责。可借助互联网获取相关职位的工作需求等信息来辅助定义新角色。

### 6. 遇到问题，立刻沟通
平时各 Agent 上下文完全独立，只做自己的事。**一旦遇到别人领域的问题，立刻找领域相关 Agent 开会对齐**，绝不自己琢磨别人领域的事。

举例：测试 Agent 发现 bug → 不自己分析 bug 来源 → 立刻告诉对应项目的开发 Agent："你需要注意有这个 bug"，并详细传递：报错信息、环境、测试方法、测试用例等。

**核心思想：专人专职，遇到问题立刻沟通，平常互不干扰。**

沟通触发条件：
1. 遇到自己领域外的问题 → 立刻找领域相关 Agent 沟通
2. 领域相关 Agent 不存在 → 立刻找 HR 招人
3. 多人无法达成一致 → 上报项目经理
4. 项目经理无法决策 → 上报 AgentServer (CEO秘书) → 请示CEO选择
5. CEO 未选择时，相关 Agent 先去完成公司中其他任务

### 7. 会议系统（结构化沟通机制）
Agent 间通过会议系统进行结构化沟通，与黑板模式互补：
- **黑板** → 持久化知识（事实、进展、决议），日常读写
- **会议** → 实时沟通（讨论、答疑），仅在遇到问题时发起

**会议规则**：
1. 多个 Agent 讨论时，各自思考内容不上报，但在会上发言需写入会议临时数据库
2. 临时数据库每行：发言人 | 发言人岗位 | 发言内容
3. 会议需形成简单决议（如："1.项目应该考虑A因素 2.目前缺少A的工作人员，应由HR招人"）
4. 形成决议后，删除临时数据库中的沟通内容，改为一条会议决议
5. 会议发起人负责分发决议给每个相关人（含未参与者）
6. 分发方式：所有相关人加入新会议 → 决议分发人宣读决议 → 其他人认可并受知 → 结束
7. 会议标志位：`need_resolution=1` 表示需要生成决议，`need_resolution=0` 表示是决议分发会议（不生成新决议）

### 8. 代码审查
专门的 Review Agent 审核所有代码变更：
- 上下文必须和开发、测试等人员隔离
- 每次审查只基于原始代码和原始需求
- 有不确定的地方 → 和需求 Agent 及开发 Agent、测试 Agent 商讨
- 商讨仍无法消除不确定 → 汇报 AgentServer (CEO秘书)
- AgentServer 向CEO发起会话，CEO作为公司最高决策者仅进行关键抉择

### 10. 公司规模上限
CEO可定义公司规模上限，HR 招聘前必须检查：
- **人数上限**：如 "最多 100 个 Agent"
- **资源上限**：如 "总内存不超过 16GB"
- 上限配置由 AgentServer 管理，CEO可随时调整
- 达到上限后，HR 暂停招聘，需要扩容时由 AgentServer 向CEO确认

### 11. 开箱即用
CEO安装后即可通过 Web 界面使用，无需复杂配置

---

## ✅ 设计矛盾与待确认项（已全部解决）

> 以下矛盾已在用户(CEO)审查后确认解决方案。

### ✅ 矛盾 1: "Agent 不直接通信" vs "Agent 需要沟通" — 已解决

**CEO决定**: 只有遇到问题才沟通，平常上下文完全独立。遇到别人领域的问题立刻开会，绝不自己琢磨。沟通产出（决议）写入黑板，沟通过程（讨论）不进入对方上下文。

双通道设计：
- **黑板通道** → 持久化知识（事实、决议、进展），日常读写
- **会议通道** → 实时沟通（讨论、答疑），仅在遇到问题时发起，过程不持久化，只保留决议
- 个人工作记忆永远对其他 Agent 不可见

### ✅ 矛盾 2: 会议数据归属 — 已解决

**CEO决定**: 每个会议独立 SQLite 文件 (`meeting_{id}.sqlite`)，与黑板完全分离。会议结束后可归档或删除。

### ✅ 矛盾 3: 沟通频率 vs 上下文隔离 — 已解决

**CEO决定**: 只有遇到问题才沟通。平时各 Agent 上下文完全独立，一旦遇到别人领域的问题立刻开会沟通对齐。沟通产出写入黑板，讨论过程不进入对方上下文。

### ✅ 待确认 4: 项目经理 Agent 角色 — 已确认

项目经理 Agent 已加入角色体系和黑板权限矩阵。职责：从 AgentServer 接收项目、拆解需求、分配任务、参加招聘会议。任何 Agent 需要招人都可以找 HR，但项目经理必须参加招聘会议共同决定。

### ✅ 待确认 5: AgentServer = CEO秘书 — 已确认

AgentServer 就是CEO秘书，与CEO直接对话，拿到项目后交给项目经理拆解。全文已统一称呼为 "AgentServer (CEO秘书)"。

---

## 与现有系统的对比

| 特性 | Hermes | MetaGPT | CrewAI | Agent-Blackboard | **Athena** |
|------|--------|---------|--------|-----------------|------------|
| **组织模型** | 单 Agent + Skills | 公司流水线 SOP | Crew 小组 | 黑板协调 | **公司架构 + 黑板** |
| **上下文隔离** | ❌ 单一上下文 | ⚠️ 部分隔离 | ⚠️ 部分隔离 | ✅ 完全隔离 | ✅ 完全隔离 |
| **Agent 动态创建** | ❌ 手动配置 | ❌ 预定义角色 | ❌ 预定义角色 | ⚠️ 注册制 | ✅ HR 自动招聘 |
| **知识持久化** | ✅ SQLite | ❌ 无持久化 | ❌ 内存态 | ✅ SQLite + 语义搜索 | ✅ SQLite + 语义搜索 |
| **专人专用** | ❌ 共享工具集 | ⚠️ 角色分工但上下文共享 | ✅ 角色分工 | ✅ 领域专家 | ✅ 专人专用 + 工具隔离 |
| **事实分级** | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ✅ 确定/猜测分级 |
| **代码审查** | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ✅ 专门 Review Agent |
| **Agent 间沟通** | ❌ 无机制 | ⚠️ 流水线传递 | ⚠️ 委托式 | ❌ 无直接沟通 | ✅ 会议系统 + 决议分发 |
| **Web 管理界面** | ❌ CLI | ❌ API | ❌ API | ❌ 无 | ✅ 内置 Web UI |
| **黑板上写回** | ❌ 无黑板 | ⚠️ 隐式共享 | ⚠️ 隐式共享 | ✅ 显式黑板 | ✅ 显式黑板 |
| **SOP 工作流** | ❌ 自由式 | ✅ 严格 SOP | ⚠️ 流程可选 | ❌ 自由式 | ✅ 可配置工作流 |

### Athena 的独特价值

Athena = **MetaGPT 的公司角色模型** + **Agent-Blackboard 的黑板模式** + **Hermes 的自我进化能力** + **独创的 HR 动态招聘 + 事实分级 + 上下文隔离 + 会议沟通机制**

---

## 技术栈选型

| 层次 | 技术 | 理由 |
|------|------|------|
| **后端语言** | Go (Golang) | 高性能、并发原生支持、单二进制部署、类型安全 |
| **后端框架** | Gin / Echo | Go 主流 Web 框架，性能优秀 |
| **数据库** | SQLite (go-sqlite3) + ChromaDB | SQLite 存储结构化数据，ChromaDB 提供语义搜索 |
| **前端** | Vue 3 + Vite | 轻量、现代、组件化好 |
| **LLM 调用** | 自研 Go LLM Client | 调用 OpenAI/Anthropic/GLM 等 API，统一接口 |
| **Agent 运行时** | 自研 goroutine-based Agent Loop | 每个 Agent 独立 goroutine，天然上下文隔离 |
| **任务队列** | Go channel + SQLite | 轻量级，channel 做实时调度，SQLite 做持久化 |
| **通信协议** | WebSocket + REST | WebSocket 实时推送，REST 管理操作 |
| **会议系统** | 每会议独立 SQLite 文件 | 会议数据独立于黑板，每个会议一个 .sqlite，会后可选归档或删除 |

---

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    Web UI (Vue 3)                        │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ 输入框    │  │ 项目看板     │  │ Agent 状态监控等   │  │
│  └──────────┘  └──────────────┘  └───────────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │ WebSocket + REST
┌────────────────────────┴────────────────────────────────┐
│                  Athena Server (Go)                       │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │           管理层 (Management)                       │  │
│  │  ┌──────────────────────────────────────────────┐  │  │
│  │  │ AgentServer (CEO秘书)                      │  │  │
│  │  │ - 与CEO直接对话、获取项目需求               │  │  │
│  │  │ - 项目规划与立项                             │  │  │
│  │  │ - 关键抉择时请示CEO                         │  │  │
│  │  │ - 评估 Agent 能力是否满足需求                  │  │  │
│  │  └──────────────────────────────────────────────┘  │  │
│  │  ┌──────────────┐  ┌───────────────────────────┐   │  │
│  │  │ HR Agent     │  │ 项目经理 Agent             │   │  │
│  │  │ - 感知缺口   │  │ - 拆解细化需求             │   │  │
│  │  │ - 创建 Agent │  │ - 分配任务                 │   │  │
│  │  │ - 配备工具   │  │ - 处理不确定性上报         │   │  │
│  │  │ - 扩展模板   │  │ - 决策：自动调研或上报CEO │   │  │
│  │  └──────────────┘  └───────────────────────────┘   │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │           执行层 (Workers) — 每个 Agent 独立 goroutine  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ 开发 A   │ │ 测试 B   │ │ 设计 C   │            │  │
│  │  │ 独立上下文│ │ 独立上下文│ │ 独立上下文│            │  │
│  │  │ 开发工具集│ │ 测试工具集│ │ 设计工具集│            │  │
│  │  └─────┬────┘ └─────┬────┘ └─────┬────┘            │  │
│  │        │            │             │                  │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ Review D │ │ 运维 E   │ │ 文档 F   │            │  │
│  │  │ 独立上下文│ │ 独立上下文│ │ 独立上下文│            │  │
│  │  └──────────┘ └──────────┘ └──────────┘            │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │           会议系统 (Meetings)                       │  │
│  │  - 每个会议独立 SQLite 文件                         │  │
│  │  - meeting_{id}.sqlite                             │  │
│  │  - 多对多讨论                                      │  │
│  │  - 会议决议生成与分发                              │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │           黑板 (Blackboard / 知识库)               │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ 项目目标 │ │ 确定事实 │ │ 猜测事实 │            │  │
│  │  └──────────┘ └──────────┘ └──────────┘            │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ 项目进展 │ │ 参与人员 │ │ 新发现   │            │  │
│  │  └──────────┘ └──────────┘ └──────────┘            │  │
│  │  ┌──────────┐                                       │  │
│  │  │ 会议决议 │ ← 决议持久化后写入黑板               │  │
│  │  └──────────┘                                       │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │              SQLite + ChromaDB                      │  │
│  │  - BOARD.sqlite (公司黑板数据库)                    │  │
│  │  - meeting_{id}.sqlite (每会议独立数据库)          │  │
│  │  - 语义搜索索引 (知识检索)                          │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## 数据库设计 (SQLite)

> ⚠️ 暂不完善，后续根据开发进展迭代

### ER 关系图

```
projects ──< project_members >── agents
   │                                │
   ├──< project_facts               │
   │                                ├──< agent_tasks
   ├──< project_goals               │
   │                                └──< agent_contexts
   └──< project_discoveries

meetings ──< meeting_participants     (每个会议独立 SQLite)
   │
   ├──< meeting_messages
   └──< meeting_resolutions
```

### 表结构

#### 1. projects — 项目索引表

```sql
CREATE TABLE projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    status      TEXT DEFAULT 'active',  -- active/paused/completed
    priority    INTEGER DEFAULT 5,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. project_goals — 项目目标表

```sql
CREATE TABLE project_goals (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,
    status      TEXT DEFAULT 'pending', -- pending/in_progress/completed/abandoned
    assigned_to TEXT REFERENCES agents(id),
    parent_goal TEXT REFERENCES project_goals(id),
    certainty   TEXT DEFAULT 'certain',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. project_facts — 项目事实表（核心！）

```sql
CREATE TABLE project_facts (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    source      TEXT,       -- 来源 Agent 或外部
    evidence    TEXT,       -- 支撑证据
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 4. project_discoveries — 新发现表

```sql
CREATE TABLE project_discoveries (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    discovered_by TEXT REFERENCES agents(id),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 5. agents — Agent 注册表

```sql
CREATE TABLE agents (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    role        TEXT NOT NULL,  -- investor/hr/pm/developer/tester/designer/reviewer/ops/doc
    project_id  TEXT REFERENCES projects(id),
    status      TEXT DEFAULT 'idle',  -- idle/working/in_meeting/offline
    tools       TEXT,           -- JSON: 可用工具列表
    model       TEXT DEFAULT 'default',
    created_by  TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 6. project_members — 项目参与人员表

```sql
CREATE TABLE project_members (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    role        TEXT NOT NULL,
    joined_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, agent_id)
);
```

#### 7. agent_tasks — Agent 任务表

```sql
CREATE TABLE agent_tasks (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    title       TEXT NOT NULL,
    description TEXT,
    status      TEXT DEFAULT 'pending',
    priority    INTEGER DEFAULT 5,
    result      TEXT,
    review_status TEXT,  -- pending_review/approved/rejected
    reviewed_by TEXT REFERENCES agents(id),
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);
```

#### 8. agent_contexts — Agent 上下文存储表

```sql
CREATE TABLE agent_contexts (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    context_type TEXT NOT NULL,  -- working_memory/session_log/skill/memory_md
    content     TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 9. blackboard_entries — 黑板条目表

```sql
CREATE TABLE blackboard_entries (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    category    TEXT NOT NULL,  -- goal/fact/discovery/decision/progress/resolution
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    author      TEXT,
    embedding   BLOB,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 10. meetings — 会议表 (每会议独立 meeting_{id}.sqlite)

```sql
CREATE TABLE meetings (
    id              TEXT PRIMARY KEY,
    project_id      TEXT NOT NULL,
    convener_id     TEXT NOT NULL,      -- 发起人 Agent ID
    need_resolution INTEGER DEFAULT 1,  -- 1=需要生成决议, 0=决议分发会议
    status          TEXT DEFAULT 'open', -- open/resolved/closed
    resolution      TEXT,               -- 会议决议内容
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at       DATETIME
);
```

#### 11. meeting_participants — 会议参与者表 (meeting_{id}.sqlite)

```sql
CREATE TABLE meeting_participants (
    id          TEXT PRIMARY KEY,
    meeting_id  TEXT NOT NULL REFERENCES meetings(id),
    agent_id    TEXT NOT NULL,
    role        TEXT DEFAULT 'participant', -- convener/participant
    acknowledged INTEGER DEFAULT 0,  -- 决议分发会议中是否已确认
    UNIQUE(meeting_id, agent_id)
);
```

#### 12. meeting_messages — 会议发言表 (meeting_{id}.sqlite)

```sql
CREATE TABLE meeting_messages (
    id          TEXT PRIMARY KEY,
    meeting_id  TEXT NOT NULL REFERENCES meetings(id),
    speaker_id  TEXT NOT NULL,      -- 发言人 Agent ID
    speaker_role TEXT NOT NULL,     -- 发言人岗位
    content     TEXT NOT NULL,      -- 发言内容
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Agent 角色体系与工具分配

### 角色 — 工具映射（专人专用！）

| 角色 | 职责 | 专属工具集 | 上下文内容 |
|------|------|-----------|-----------|
| **AgentServer (CEO秘书)** | 与CEO直接对话、获取项目、关键抉择请示CEO、项目交给项目经理 | 项目管理工具、HR 调度工具、CEO交互 | 全局视角 |
| **HR Agent** | 感知缺口、创建 Agent、分配工具、扩展模板 | Agent 模板库、工具注册表、互联网搜索、配置生成器 | 公司组织架构 |
| **项目经理 Agent** | 拆解细化需求、分配任务、处理不确定性上报 | 任务分解工具、需求分析工具 | 项目目标 + 需求文档 |
| **Developer Agent** | 代码开发 | 文件读写、代码执行、Git、Debug 工具 | 项目目标 + 技术上下文 + 自己的历史 |
| **Tester Agent** | 测试 | 测试框架、覆盖率工具、断言工具 | 项目目标 + 接口定义 + 自己的历史 |
| **Designer Agent** | 架构/接口设计 | 画图工具、API 设计工具 | 项目目标 + 需求文档 + 自己的历史 |
| **Reviewer Agent** | 代码审查（上下文隔离） | 文件读取、Linter、Diff 工具 | 原始代码 + 原始需求（不包含开发者的思考过程） |
| **Ops Agent** | 部署/运维 | Docker、CI/CD、监控工具 | 项目目标 + 环境配置 + 自己的历史 |
| **Doc Agent** | 文档编写 | 文件读写、模板工具 | 项目目标 + 代码文档 + 自己的历史 |

### 上下文注入策略

每个 Agent 启动时，系统自动注入：

```
Agent 上下文 =
    公司级共享信息 (来自黑板)         ← 所有 Agent 相同
    + 项目级共享信息 (来自黑板)        ← 同项目 Agent 相同
    + 会议决议 (来自黑板 resolution)   ← 相关 Agent 相同
    + 角色级指令 (角色 soul.md)        ← 同角色 Agent 相同
    + 个人工作记忆 (agent_contexts)    ← 每个 Agent 独有
```

**关键隔离点**：
- Developer A 和 Developer B 同项目，但**看不到对方的工作记忆**
- Tester 看不到 Developer 的调试过程，只看到黑板上的接口定义
- Reviewer 看到的是原始代码和原始需求，不是 Developer 的思考过程
- 会议讨论过程不对非参与者可见，只有决议写入黑板后所有相关人可见

---

## HR Agent 的招聘流程（参考 Hermes Skills 机制）

### 招聘触发

任何 Agent 都可以向 HR 提出招聘请求，但必须经过招聘会议审核：
- 项目经理拆解需求时发现需要新角色 → 立刻和 HR 沟通
- 任何 Agent 遇到自己领域外的问题且无对应同事 → 找 HR 招人
- 招聘会议：请求人 + HR + 项目经理 **必须参加**，共同决定是否招人

### 招聘上限检查

HR 招聘前必须检查公司规模上限（由CEO通过 AgentServer 设定）：
- 当前 Agent 数量 < 人数上限
- 预计新增内存 < 资源上限
- 达到上限 → HR 暂停招聘，通知 AgentServer 向CEO申请扩容

### 招聘流程

```
1. 任意 Agent 发现缺少某种能力 → 向 HR 提出招聘请求:
   {
     "requester_id": "agent-xxx",
     "requester_role": "tester",
     "reason": "项目需要Go后端开发，目前没有开发Agent",
     "role": "developer",
     "project_id": "proj-xxx",
     "required_skills": ["go", "gin", "sqlite"],
     "context_requirements": ["project_goals", "tech_stack"]
   }
2. HR 组织招聘会议 → 请求人 + HR + 项目经理参加
3. 会议讨论并确认: 是否真的需要招人？招什么样的人？
4. HR 检查公司规模上限 → 超限则上报 AgentServer 向CEO申请扩容
5. HR Agent 执行招聘:
   a. 从 Agent 模板库选择对应角色模板
   b. 如果模板不存在 → 成立小组编写新模板 (任务写入 BOARD.sqlite)
   c. 生成唯一 Agent ID 和名称
   d. 配置专属工具集 (根据 role + required_skills)
   e. 可借助互联网搜索该职位的工作需求，辅助定义 Agent 职责
   f. 注入初始上下文 (项目目标 + 角色 soul.md)
   g. 注册到 agents 表 + project_members 表
   h. 启动 Agent 运行时 (独立 goroutine)
6. 新 Agent 上线，项目经理分配任务
```

### Agent 模板结构 (YAML)

```yaml
# templates/developer.yaml
role: developer
name_template: "dev-{adjective}"
model: default
system_prompt: |
  你是一名专业的软件开发工程师。你的职责是编写高质量、可维护的代码。
  你只负责开发，不负责测试、设计或审查。
  遵循项目的技术规范和编码标准。
  将你的工作进展写入黑板，将发现的事实标记为"确定"或"猜测"。
  遇到别人领域的问题（如测试bug、设计疑问），立刻找对应Agent开会对齐，绝不自己琢磨。
  在会议中发言时，将发言写入会议临时数据库。

tools:
  - file_read
  - file_write
  - file_edit
  - bash
  - git
  - code_search
  - debug

blackboard_read:
  - project_goals
  - project_facts
  - tech_spec
  - api_definitions
  - meeting_resolutions

blackboard_write:
  - project_facts
  - project_discoveries
  - progress_updates

context_injection:
  - project_goals
  - tech_stack
  - coding_standards
```

---

## 会议系统设计

会议系统是 Athena 独创的 Agent 间结构化沟通机制，与黑板模式互补：

- **黑板** → 持久化知识（事实、进展、决议）
- **会议** → 实时沟通（讨论、答疑），仅在遇到问题时发起

### 会议存储设计

每个会议使用独立的 SQLite 文件 (`meeting_{id}.sqlite`)，与黑板数据库完全分离：
- 会议结束后，决议写入黑板，讨论内容删除
- 会议 SQLite 文件可归档或删除，不影响黑板数据
- 独立文件设计便于并行开会、生命周期管理

### 会议生命周期

```
1. 发起 → Agent 遇到别人领域的问题，发起会议
2. 创建会议 → 创建 meeting_{id}.sqlite, need_resolution=1
3. 邀请参与者 → 写入 meeting_participants
4. 讨论 → 参与者发言写入 meeting_messages
5. 形成决议 → 发起人总结写入 meetings.resolution
6. 清理 → 删除 meeting_messages 中的讨论内容
7. 分发 → 创建新会议 need_resolution=0, 参与者确认受知
8. 持久化 → 决议写入黑板 blackboard_entries (category=resolution)
9. 关闭 → 会议状态改为 closed, meeting_{id}.sqlite 可归档/删除
```

### 会议类型

| 类型 | need_resolution | 用途 |
|------|----------------|------|
| 讨论会议 | 1 | 多 Agent 讨论问题，生成决议 |
| 决议分发会议 | 0 | 向未参与者传达决议，仅确认受知 |

---

## 黑板模式设计（参考 TCH Bytex + Agent-Blackboard）

### 黑板架构

黑板是 Athena 的核心知识共享机制，参考了：
- **TCH Bytex 方案**: 黑板 + DAG 的多 Agent 协调
- **Agent-Blackboard 项目**: MCP 持久化 + 语义搜索 + 本体约束
- **经典黑板模式**: HEARSAY-II 的知识源 + 控制组件

```
┌─────────────────────────────────────────────┐
│               Blackboard                     │
│                                              │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 0: 项目元信息                       │ │
│  │ - 项目名称、目标、优先级                 │ │
│  │ - 项目状态、时间线                       │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 1: 确定性事实 (certainty=certain)   │ │
│  │ - 已验证的技术决策                       │ │
│  │ - 已确认的接口定义                       │ │
│  │ - 已通过的审查结论                       │ │
│  │ ⚠️ 写入要求: 必须 100% 确定！           │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 2: 猜测/假设 (certainty=conjecture)  │ │
│  │ - 待验证的方案                           │ │
│  │ - 初步分析结果                           │ │
│  │ - 需要进一步调查的发现                   │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 3: 工作进展                         │ │
│  │ - Agent 的工作日志                       │ │
│  │ - 任务完成报告                           │ │
│  │ - 阻塞和依赖                            │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 4: 新发现                           │ │
│  │ - 开发中发现的 bug                       │ │
│  │ - 测试中发现的边界条件                   │ │
│  │ - 设计中的新想法                         │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ 层级 5: 会议决议                         │ │
│  │ - 来自会议系统的决议                     │ │
│  │ - 所有相关人可见                         │ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
         ↑ 写入              ↓ 读取
    ┌────┴────┐         ┌────┴────┐
    │ Agent A │         │ Agent B │
    │ (写事实) │         │ (读决议) │
    └─────────┘         └─────────┘
```

### 黑板读写控制矩阵

| 角色 | 层级0 | 层级1 | 层级2 | 层级3 | 层级4 | 层级5 |
|------|-------|-------|-------|-------|-------|-------|
| AgentServer (CEO秘书) | RW | RW | RW | R | R | R |
| HR | R | R | R | R | R | R |
| 项目经理 | RW | RW | RW | RW | RW | R |
| Developer | R | R | RW | RW | RW | R |
| Tester | R | R | RW | RW | RW | R |
| Designer | R | RW | RW | RW | RW | R |
| Reviewer | R | RW | R | R | RW | R |
| Ops | R | R | RW | RW | RW | R |
| Doc | R | R | R | RW | R | R |

> R = 只读, RW = 可读写

---

## Agent 运行时架构

### 单个 Agent 的运行循环 (Go)

```go
func AgentLoop(ctx context.Context, agent *Agent) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case task := <-agent.TaskQueue:
            // 1. 构建上下文 (从黑板读取 + 个人上下文)
            context := BuildContext(agent)
            // context = 黑板共享信息 + 角色指令 + 个人工作记忆

            // 2. 检查是否有会议邀请
            if meeting := CheckMeetingInvite(agent); meeting != nil {
                HandleMeeting(agent, meeting)
                continue
            }

            // 3. 调用 LLM (使用角色专属 system_prompt + tools)
            response, err := LLMCall(ctx, agent.Model,
                agent.SystemPrompt,
                context,
                agent.Tools,  // 只能使用自己的工具
                task,
            )

            // 4. 执行工具调用
            for _, toolCall := range response.ToolCalls {
                result := ExecuteTool(agent, toolCall)
                // 工具执行结果加入 Agent 自己的上下文
            }

            // 5. 将发现写入黑板 (如果有)
            if response.BlackboardUpdates != nil {
                blackboard.Write(agent.ProjectID, response.BlackboardUpdates, agent.ID)
            }

            // 6. 更新个人工作记忆
            agent.UpdateWorkingMemory(response)

            // 7. 如果需要 Review (如代码提交)
            if response.NeedsReview {
                SubmitForReview(agent, response.Artifact)
            }
        }
    }
}
```

### Agent 间通信规则（修订版）

```
规则 1: Agent 的个人工作记忆对其他 Agent 不可见
规则 2: 知识持久化通过黑板（事实、进展、决议等）
规则 3: 实时沟通通过会议系统（讨论 → 决议 → 黑板）
规则 4: 每个 Agent 只能看到黑板中自己有权限的部分
规则 5: 任务分配通过项目经理 Agent
规则 6: 只有遇到问题才沟通：遇到别人领域的问题立刻找对应 Agent 开会
规则 7: 遇到领域问题且无对应 Agent → 找 HR 招人
规则 8: 不确定问题处理链路: Agent → 联系相关人 → 项目经理 → AgentServer (CEO秘书) → CEO
规则 9: 会议讨论过程不对非参与者可见，只有决议写入黑板后所有相关人可见
规则 10: 任何 Agent 可向 HR 提出招聘请求，但必须经过招聘会议审核（PM+HR+请求人）
```

---

## Web 管理界面设计

### 页面布局

```
┌─────────────────────────────────────────────────────┐
│  Athena                              [设置] [关于]   │
├──────────┬──────────────────────────────────────────┤
│ 侧边栏   │  主内容区                                │
│          │                                          │
│ 📋 项目  │  ┌──────────────────────────────────────┐│
│  ├ 项目A │  │ 输入框                               ││
│  └ 项目B │  │ [在这里输入你的需求，Athena 会处理...] ││
│          │  │                              [发送]   ││
│ 👥 团队  │  └──────────────────────────────────────┘│
│  ├ 开发   │                                          │
│  ├ 测试   │  ┌──────────────────────────────────────┐│
│  └ 设计   │  │ 项目看板                              ││
│          │  │ ┌────────┐ ┌────────┐ ┌────────┐      ││
│ 📊 黑板  │  │ │ 目标    │ │ 进行中  │ │ 完成    │      ││
│  ├ 事实   │  │ │ ○ 任务1 │ │ ○ 任务3 │ │ ✅ 任务5 │     ││
│  ├ 猜测   │  │ │ ○ 任务2 │ │ ○ 任务4 │ │ ✅ 任务6 │     ││
│  ├ 决议   │  │ └────────┘ └────────┘ └────────┘      ││
│  └ 进展   │  └──────────────────────────────────────┘│
│          │                                          │
│ 🤝 会议  │  ┌──────────────────────────────────────┐│
│  ├ 进行中 │  │ Agent 状态                             ││
│  └ 决议   │  │ 🟢 dev-alice  开发中  [任务3]         ││
│          │  │ 🟢 tester-bob 测试中  [任务5]         ││
│ ⚙ 设置   │  │ 🟡 reviewer   审查中  [会议2]         ││
│          │  │ ⚪ designer-cat 空闲                   ││
│          │  └──────────────────────────────────────┘│
└──────────┴──────────────────────────────────────────┘
```

### 核心 API

```go
// CEO输入
POST /api/projects/{id}/chat
Body: {"message": "给这个项目添加用户认证功能"}

// 项目管理
GET  /api/projects
POST /api/projects
GET  /api/projects/{id}

// Agent 管理
GET  /api/projects/{id}/agents
POST /api/projects/{id}/agents/hire

// 黑板操作
GET  /api/projects/{id}/blackboard
GET  /api/projects/{id}/facts
GET  /api/projects/{id}/discoveries
GET  /api/projects/{id}/resolutions

// 会议操作
GET  /api/projects/{id}/meetings
GET  /api/meetings/{id}
GET  /api/meetings/{id}/resolution

// CEO抉择
GET  /api/projects/{id}/decisions     # 获取待CEO抉择的选项
POST /api/projects/{id}/decisions     # CEO做出抉择

// 实时通知
WS   /ws/projects/{id}/events
```

---

## 项目目录结构

```
athena/
├── README.md
├── PLAN.md                       # 本文件
├── go.mod
├── go.sum
├── main.go                       # 入口
├── cmd/
│   └── athena/
│       └── main.go               # CLI 入口
│
├── internal/
│   ├── server/                   # HTTP 服务
│   │   ├── server.go             # Gin 服务入口
│   │   ├── router.go             # 路由注册
│   │   └── middleware.go         # 中间件
│   │
│   ├── core/                     # 核心模块
│   │   ├── agent_server.go       # AgentServer (CEO秘书)
│   │   ├── hr_agent.go           # HR Agent
│   │   ├── pm_agent.go           # 项目经理 Agent
│   │   ├── agent_runtime.go      # Agent 运行时
│   │   ├── agent_loop.go         # Agent 主循环
│   │   └── llm_client.go        # LLM 调用封装
│   │
│   ├── blackboard/               # 黑板系统
│   │   ├── board.go              # 黑板核心逻辑
│   │   ├── fact_manager.go       # 事实管理 (确定/猜测)
│   │   ├── access_control.go     # 读写权限控制
│   │   └── search.go             # 语义搜索 (ChromaDB)
│   │
│   ├── meeting/                  # 会议系统
│   │   ├── manager.go            # 会议管理
│   │   ├── database.go           # 每会议独立 SQLite 管理
│   │   └── resolution.go         # 决议生成与分发
│   │
│   ├── db/                       # 数据库层
│   │   ├── database.go           # SQLite 连接管理
│   │   ├── models.go             # 数据模型
│   │   └── migrations/           # 数据库迁移
│   │
│   ├── templates/                # Agent 模板
│   │   ├── investor.yaml         # CEO秘书模板 (AgentServer)
│   │   ├── hr.yaml               # HR 模板
│   │   ├── pm.yaml               # 项目经理模板
│   │   ├── developer.yaml
│   │   ├── tester.yaml
│   │   ├── designer.yaml
│   │   ├── reviewer.yaml
│   │   ├── ops.yaml
│   │   └── doc.yaml
│   │
│   ├── tools/                    # 工具定义
│   │   ├── base.go               # 工具接口
│   │   ├── file_tools.go
│   │   ├── code_tools.go
│   │   ├── test_tools.go
│   │   ├── git_tools.go
│   │   ├── design_tools.go
│   │   └── review_tools.go
│   │
│   ├── api/                      # API 路由
│   │   ├── projects.go
│   │   ├── agents.go
│   │   ├── blackboard.go
│   │   ├── chat.go
│   │   ├── meetings.go
│   │   └── websocket.go
│   │
│   └── prompts/                  # Prompt 模板 (soul.md)
│       ├── investor.md
│       ├── hr.md
│       ├── pm.md
│       ├── developer.md
│       ├── tester.md
│       ├── designer.md
│       ├── reviewer.md
│       ├── ops.md
│       └── doc.md
│
├── frontend/                     # Vue 3 前端
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   ├── views/
│   │   │   ├── Dashboard.vue
│   │   │   ├── Project.vue
│   │   │   ├── Blackboard.vue
│   │   │   └── Meetings.vue
│   │   ├── components/
│   │   │   ├── ChatInput.vue
│   │   │   ├── ProjectBoard.vue
│   │   │   ├── AgentStatus.vue
│   │   │   ├── FactList.vue
│   │   │   └── MeetingList.vue
│   │   └── stores/
│   │       └── project.ts
│   └── index.html
│
├── tests/
│   ├── blackboard_test.go
│   ├── agent_runtime_test.go
│   ├── hr_agent_test.go
│   └── meeting_test.go
│
└── scripts/
    ├── setup.sh
    └── dev.sh
```

---

## 开发路线图

### Phase 1: 基础骨架 (2 周)

- [ ] Go 项目脚手架搭建 (Gin + SQLite + Vue 3)
- [ ] 数据库 Schema 实现与迁移脚本
- [ ] Agent 运行时核心 (agent_loop.go, goroutine 模型)
- [ ] LLM Client 封装 (统一多模型接口)
- [ ] 基础 API (项目 CRUD + WebSocket)
- [ ] 最简前端 (输入框 + 项目列表)

### Phase 2: 黑板与核心 Agent (2 周)

- [ ] 黑板系统实现 (board.go + fact_manager.go)
- [ ] 事实分级系统 (确定/猜测)
- [ ] AgentServer (CEO秘书) 实现
- [ ] HR Agent 实现 (模板化 Agent 创建 + 模板扩展)
- [ ] 项目经理 Agent 实现 (需求拆解 + 不确定性处理)
- [ ] Developer Agent + 基础工具集
- [ ] 上下文注入与隔离机制

### Phase 3: 会议系统与审查 (2 周)

- [ ] 会议系统实现 (manager.go + database.go + resolution.go)
- [ ] 会议决议生成与分发流程
- [ ] Reviewer Agent + 代码审查流程 (上下文隔离)
- [ ] 不确定问题上报链路 (Agent → 相关人 → PM → AgentServer → CEO)
- [ ] CEO抉择 API 与界面

### Phase 4: 全角色 Agent + 打磨 (2 周)

- [ ] Tester Agent + 测试工具集
- [ ] Designer Agent + 设计工具集
- [ ] Ops Agent + 部署工具集
- [ ] Doc Agent + 文档工具集
- [ ] 语义搜索集成 (ChromaDB)
- [ ] 完整 Web 管理界面
- [ ] Agent 状态监控 + 会议可视化

### Phase 5: 高级特性 (后续)

- [ ] Agent 自我进化 (参考 Hermes Skills)
- [ ] DAG 任务编排
- [ ] 多项目并行支持
- [ ] 外部 MCP 工具集成
- [ ] 插件系统
- [ ] ABANDON 机制 (防止 Agent 陷入死循环)
- [ ] HR 互联网搜索辅助招聘

---

## 关键设计决策

### Q1: 为什么不用 CrewAI / MetaGPT 作为底层框架？

**A**: 它们不满足核心需求——
- CrewAI: Agent 共享上下文，无持久化，无动态创建
- MetaGPT: 严格 SOP 流水线，Agent 上下文不隔离
- 我们需要: 上下文隔离 + 动态招聘 + 黑板持久化 + 事实分级 + 会议沟通

### Q2: 为什么选 SQLite 而不是 PostgreSQL？

**A**: 开箱即用原则。SQLite 零配置，单文件，用户(CEO)无需安装数据库服务。后续可通过替换 database.go 支持 PostgreSQL。

### Q3: 为什么后端用 Go 而不是 Python？

**A**: Go 的优势——
- 每个 Agent 独立 goroutine，天然上下文隔离
- 原生并发，无需 asyncio
- 单二进制部署，用户(CEO)无需安装 Python 环境
- 类型安全，编译期检查
- 性能更好，适合同时运行多个 Agent

### Q4: 如何防止 Agent 写入错误的"确定性事实"？

**A**: 三重保护——
1. Prompt 层: 反复强调只有 100% 确定的事实才能标记为"确定"
2. 审查层: Reviewer Agent 审核黑板写入的事实
3. 升级层: 其他 Agent 可对"确定"事实提出质疑，降级为"猜测"

### Q5: HR Agent 如何决定招聘什么角色？

**A**: 专人专职 + 会议审核——
1. 任何 Agent 发现缺少某种能力 → 向 HR 提出招聘请求（附详细理由和岗位需求）
2. HR 组织招聘会议 → 请求人 + HR + 项目经理参加，共同决定是否招人
3. HR 检查公司规模上限（人数/资源），超限上报 AgentServer 向CEO申请
4. 确认招聘后，HR 从模板库选择角色模板，自动配置工具和上下文
5. 模板不足时，HR 成立小组编写新模板
6. 可借助互联网搜索职位需求辅助定义
7. 新 Agent 上线后自动注册到项目和黑板，项目经理分配任务

### Q6: Agent 间沟通与上下文隔离如何平衡？

**A**: 双通道设计——
- **黑板通道**: 持久化知识（事实、进展、决议），所有 Agent 按权限访问
- **会议通道**: 实时沟通（讨论、答疑），过程不持久化，只保留决议
- 个人工作记忆永远对其他 Agent 不可见

---

## 参考项目

| 项目 | 借鉴点 |
|------|--------|
| **Hermes Agent** (NousResearch) | Agent 运行时架构、Skills 系统、LLM 多模型支持、工具可见性隔离 |
| **MetaGPT** (FoundationAgents) | 公司角色模型、SOP 工作流、角色间文档传递 |
| **CrewAI** | Agent 角色定义 (role/goal/backstory)、YAML 配置解耦 |
| **Agent-Blackboard** (claudioed) | 黑板模式实现、MCP 持久化、语义搜索、本体约束 |
| **TCH Bytex** (Bytex) | 黑板 + DAG 架构、Agent 协调、上下文隔离 |
| **CHYing-agent** (yhy0) | 工具可见性隔离 (visibility)、ABANDON 死循环检测、编译器模式 |
| **腾讯 WorkBuddy** | 任务拆解与需求细化方式 |

---

*本文档是 Athena 项目的规划文档，将随开发进展持续更新。*
*最后更新: 2026-05-02 (v3: 用户=CEO, AgentServer=CEO秘书)*
