# Athena (雅典娜) — AI Agent 公司化编排系统

> **项目代号**: Athena（雅典娜）— 希腊神话中智慧与战略女神
> **命名寓意**: 雅典娜是策略之神，她不靠蛮力而靠智慧与组织取胜。正如我们的系统——通过智能编排和组织一群专业 Agent 协同工作，而非让一个全能 Agent 独自承担一切。雅典娜引导英雄们各司其职，正如我们引导每个 Agent 在自己的专业领域发挥作用。

## 为什么叫 Athena？

| 维度 | 雅典娜的象征 | Athena 系统的对应 |
|------|------------|-----------------|
| 智慧与战略 | 统筹全局，制定策略 | AgentServer 作为 CEO 统筹项目方向 |
| 引导英雄 | 指引奥德修斯等英雄各展所长 | HR Agent 招聘并引导专业 Agent |
| 工艺之神 | 纺织、造船等技艺的守护者 | 每个 Agent 专人专用，各有所长 |
| 猫头鹰之眼 | 洞察全局，看见本质 | 黑板系统让所有 Agent 共享项目全貌 |
| 从宙斯头中诞生 | 完整的智慧，一经诞生即具备能力 | Agent 被创建时即配备完整专业工具和上下文 |

---

## 一句话定义

**Athena 是一个像 IT 公司一样运作的 AI Agent 编排系统**——用户开箱即用，通过 Web 界面下达项目需求，系统自动招聘专业 Agent(从已有的模板和代码中继承出来, 如果目前的模板不足以完成招聘任务, HR agent应当成立一个专门的小组(相关任务会放入公司的黑板数据库 BOARD.sqlite)书写相关的job desire和模板等, 用于招聘新方向的员工)、分配任务、协调推进，每个 Agent 上下文隔离、专人专用，通过黑板模式共享项目知识。最初公司只需要一个agent sever(与用户对话 获取需求等) 和一个hr agent 还有一个项目经理agent(拆解和细化项目需求的专家, 这部分可以参考hermes, 腾讯workbuddy等agent), 以及一套通用的员工模板(soul.md等)等

---

## 核心设计原则

1. **专人专用**: 开发只做开发，测试只做测试，设计只做设计——绝不跨岗
2. **上下文隔离**: **每个** Agent 有独立的工作记忆和上下文窗口，互不干扰, 部分记忆是共通的 比如公司总体任务等, 但是剩下的记忆需要像现实的员工那样, 各自管理, 不要互相干扰
3. **黑板共享**: 项目目标、进展、确定性事实通过中央黑板共享给所有 Agent
4. **事实分级**: 确定性事实标记为"确定"（必须 100% 可靠），猜测标记为"猜测"
5. **自动招聘**: HR Agent 感知能力缺口时自动创建新 Agent，按需配备工具, 以及新agent职责(可以借助互联网获取相关职位的工作需求等信息)
6.  **agent组会 自动沟通** : 当项目遇到问题或者不确定节点时, 发现这种问题的agent需要联系相关人员(找相关人员的办法需要写在soul文件里), 他会先确认信息的来源(员工还是互联网还是其他什么), 并结合其他信息(比如, 这个信息是其他员工在会议上告诉他的, 那么或许进一步和这位员工沟通是一个好办法), 如果这几个当事人都无法沟通结果(比如: 原始需求确实非常模糊), 则上报项目经理, 由项目经理决定 1. 自动调研并采取最主流做法 2.给出一些合适的选项询问server, server会给用户展现这些选项, 选择完成后继续任务, 如果用户一直没有选择, 则相关的agent先去完成公司中其他任务
7.  **agent相互沟通机制**: agent要勤于与其他agent沟通项目 多讨论, 如果对方知道相关问题的答案, 则告知答案, 答案的搜寻方法, 思考模式等, 如果接收方觉得有价值 要写在他的memory.md (构建多对多沟通机制, 比如三个agent讨论, 自己思考的内容不需要上报, 但是需要在会上发言的话, 则需要写入该会议的临时数据库, 数据库每一行是发言人 发言人的岗位 发言内容等 会议需要形成会议决议, 简单的那种就行, 然后删除临时数据库里面的沟通 改成一个会议决议, 举例: 1.项目应该考虑A因素 2. 目前缺少A的工作人员 应该由hr招人; 这种简单的会议决议即可, 并且会议发起人要负责把会议决议分发给每个相关人(相关人可能没参与这个会议) 分发方式是这些所有相关人加入一个新的会议 会议内容是决议分发人先说会议决议, 然后其他人全都认可并受到 就可以结束会议, 会议决议的分发不需要保存新的会议决议(因此开会时需要有一个标志位, 1代表有会议决议需要由会议发起人生成, 0代表该会议不需要会议决议, 因为这个会议就是会议决议的分发会议))
8. **代码审查**: 专门的 Review Agent 审核所有代码变更, 他的上下文必须要和开发, 测试等人员隔离, 每次审查时只基于原始代码和原始需求, 如果有不确定的地方, 需要和需求agent以及开发agent 测试agent等人进行商讨, 除非这种商讨依旧无法消除其中的不确定因素, 才需要汇报投资人agent, 再由他向agentserver进行用户会话, 用户作为公司管理人员仅进行这些关键抉择
9.  **开箱即用**: 用户安装后即可通过 Web 界面使用，无需复杂配置

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
| **Web 管理界面** | ❌ CLI | ❌ API | ❌ API | ❌ 无 | ✅ 内置 Web UI |
| **黑板上写回** | ❌ 无黑板 | ⚠️ 隐式共享 | ⚠️ 隐式共享 | ✅ 显式黑板 | ✅ 显式黑板 |
| **SOP 工作流** | ❌ 自由式 | ✅ 严格 SOP | ⚠️ 流程可选 | ❌ 自由式 | ✅ 可配置工作流 |

### Athena 的独特价值

Athena = **MetaGPT 的公司角色模型** + **Agent-Blackboard 的黑板模式** + **Hermes 的自我进化能力** + **独创的 HR 动态招聘 + 事实分级 + 上下文隔离**

---

## 技术栈选型

| 层次 | 技术 | 理由 |
|------|------|------|
| **后端框架** | FastAPI (Python) | 异步支持好、生态丰富、与 AI SDK 集成方便 |
| **数据库** | SQLite + ChromaDB | SQLite 存储结构化数据，ChromaDB 提供语义搜索 |
| **前端** | Vue 3 + Vite | 轻量、现代、组件化好 |
| **LLM 调用** | litellm | 统一多模型接口，支持 OpenAI/Anthropic/GLM 等 |
| **Agent 运行时** | 自研 async Agent Loop | 参考 Hermes 的 agent loop，但实现上下文隔离 |
| **任务队列** | asyncio.Queue + SQLite | 轻量级，无需外部依赖 |
| **通信协议** | WebSocket + REST | WebSocket 用于实时状态推送，REST 用于管理操作 |

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
│                  Athena Server (FastAPI)                  │
│  ┌────────────────────────────────────────────────────┐  │
│  │              AgentServer (CEO)                      │  │
│  │  - 接收用户需求                                     │  │
│  │  - 项目规划与立项                                   │  │
│  │  - 评估当前 Agent 能力是否满足需求                   │  │
│  └──────────────┬─────────────────────────────────────┘  │
│                 │                                         │
│  ┌──────────────┴─────────────────────────────────────┐  │
│  │              HR Agent (人力资源)                     │  │
│  │  - 感知能力缺口                                     │  │
│  │  - 根据模板创建新 Agent                             │  │
│  │  - 为新 Agent 配备专属工具集                        │  │
│  │  - 管理 Agent 生命周期                               │  │
│  └──────────────┬─────────────────────────────────────┘  │
│                 │ 创建并管理                              │
│  ┌──────────────┴─────────────────────────────────────┐  │
│  │              Worker Agents (员工)                    │  │
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
│  │              Blackboard (黑板/知识库)               │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ 项目目标 │ │ 确定事实 │ │ 猜测事实 │            │  │
│  │  └──────────┘ └──────────┘ └──────────┘            │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │  │
│  │  │ 项目进展 │ │ 参与人员 │ │ 新发现   │            │  │
│  │  └──────────┘ └──────────┘ └──────────┘            │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │              SQLite + ChromaDB                      │  │
│  │  - 结构化数据存储 (项目、Agent、任务)               │  │
│  │  - 语义搜索索引 (知识检索)                          │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## 数据库设计 (SQLite)
(暂不完善)
### ER 关系图

```
projects ──< project_members >── agents
   │                                │
   ├──< project_facts               │
   │                                ├──< agent_tasks
   ├──< project_goals               │
   │                                └──< agent_contexts
   └──< project_discoveries
```

### 表结构

#### 1. projects — 项目索引表

```sql
CREATE TABLE projects (
    id          TEXT PRIMARY KEY,       -- UUID
    name        TEXT NOT NULL,          -- 项目名称
    description TEXT,                   -- 项目描述
    status      TEXT DEFAULT 'active',  -- active/paused/completed
    priority    INTEGER DEFAULT 5,      -- 1-10
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 2. project_goals — 项目目标表

```sql
CREATE TABLE project_goals (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,          -- 目标描述
    status      TEXT DEFAULT 'pending', -- pending/in_progress/completed/abandoned
    assigned_to TEXT REFERENCES agents(id),  -- 负责的 Agent
    parent_goal TEXT REFERENCES project_goals(id), -- 子目标关系
    certainty   TEXT DEFAULT 'certain', -- certain/conjecture
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. project_facts — 项目事实表（核心！）

```sql
CREATE TABLE project_facts (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    content     TEXT NOT NULL,          -- 事实内容
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    -- certainty = 'certain': 必须 100% 确定，任何 Agent 都可信赖
    -- certainty = 'conjecture': 猜测/假设，需进一步验证
    source      TEXT,                   -- 来源 Agent 或外部
    evidence    TEXT,                   -- 支撑证据
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
    id          TEXT PRIMARY KEY,       -- UUID
    name        TEXT NOT NULL,          -- Agent 名称，如 "dev-alice"
    role        TEXT NOT NULL,          -- developer/tester/designer/reviewer/ops/doc
    project_id  TEXT REFERENCES projects(id), -- 所属项目
    status      TEXT DEFAULT 'idle',    -- idle/working/offline
    tools       TEXT,                   -- JSON: 可用工具列表
    model       TEXT DEFAULT 'default', -- 使用的 LLM 模型
    created_by  TEXT,                   -- 创建者 (HR Agent)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 6. project_members — 项目参与人员表

```sql
CREATE TABLE project_members (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_id    TEXT NOT NULL REFERENCES agents(id),
    role        TEXT NOT NULL,          -- 在项目中的角色
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
    status      TEXT DEFAULT 'pending', -- pending/in_progress/completed/failed
    priority    INTEGER DEFAULT 5,
    result      TEXT,                   -- 任务结果
    review_status TEXT,                 -- pending_review/approved/rejected
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
    context_type TEXT NOT NULL,         -- working_memory/session_log/skill
    content     TEXT NOT NULL,          -- JSON 或 Markdown
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 9. blackboard_entries — 黑板条目表

```sql
CREATE TABLE blackboard_entries (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    category    TEXT NOT NULL,          -- goal/fact/discovery/decision/progress
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture')),
    author      TEXT,                   -- 写入的 Agent ID
    embedding   BLOB,                   -- 语义向量 (ChromaDB 同步)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Agent 角色体系与工具分配

### 角色 — 工具映射（专人专用！）

| 角色 | 职责 | 专属工具集 | 上下文内容 |
|------|------|-----------|-----------|
| **AgentServer (CEO)** | 接收需求、项目规划、能力评估 | 项目管理工具、HR 调度工具 | 全局视角 |
| **HR Agent** | 感知缺口、创建 Agent、分配工具 | Agent 模板库、工具注册表、配置生成器 | 公司组织架构 |
| **Developer Agent** | 代码开发 | 文件读写、代码执行、Git、Debug 工具 | 项目目标 + 技术上下文 + 自己的历史 |
| **Tester Agent** | 测试 | 测试框架、覆盖率工具、断言工具 | 项目目标 + 接口定义 + 自己的历史 |
| **Designer Agent** | 架构/接口设计 | 画图工具、API 设计工具 | 项目目标 + 需求文档 + 自己的历史 |
| **Reviewer Agent** | 代码审查 | 文件读取、Linter、Diff 工具 | 项目目标 + 代码规范 + 自己的历史 |
| **Ops Agent** | 部署/运维 | Docker、CI/CD、监控工具 | 项目目标 + 环境配置 + 自己的历史 |
| **Doc Agent** | 文档编写 | 文件读写、模板工具 | 项目目标 + 代码文档 + 自己的历史 |

### 上下文注入策略

每个 Agent 启动时，系统会自动注入：

```
Agent 上下文 = 
    公司级共享信息 (来自黑板)     ← 所有 Agent 相同
    + 项目级共享信息 (来自黑板)    ← 同项目 Agent 相同
    + 角色级指令 (角色 Prompt)     ← 同角色 Agent 相同
    + 个人工作记忆 (agent_contexts) ← 每个 Agent 独有
```

**关键隔离点**：
- Developer A 和 Developer B 虽然同项目，但**看不到对方的工作记忆**
- Tester 看不到 Developer 的调试过程，只看到黑板上的接口定义
- Reviewer 看到的是待审代码，不是 Developer 的思考过程

---

## HR Agent 的招聘流程（参考 Hermes Skills 机制）

```
1. AgentServer 评估需求 → 发现缺少某种能力的 Agent
2. 向 HR Agent 发出招聘请求:
   {
     "role": "developer",
     "project_id": "proj-xxx",
     "required_skills": ["python", "fastapi", "sqlite"],
     "context_requirements": ["project_goals", "tech_stack"]
   }
3. HR Agent 执行招聘:
   a. 从 Agent 模板库选择对应角色模板
   b. 生成唯一 Agent ID 和名称
   c. 配置专属工具集 (根据 role + required_skills)
   d. 注入初始上下文 (项目目标 + 角色指令)
   e. 注册到 agents 表 + project_members 表
   f. 启动 Agent 运行时
4. 新 Agent 上线，开始接收任务
```

### Agent 模板结构 (YAML)

```yaml
# templates/developer.yaml
role: developer
name_template: "dev-{adjective}"  # dev-swift, dev-clever, etc.
model: default
system_prompt: |
  你是一名专业的软件开发工程师。你的职责是编写高质量、可维护的代码。
  你只负责开发，不负责测试、设计或审查。
  遵循项目的技术规范和编码标准。
  将你的工作进展写入黑板，将发现的事实标记为"确定"或"猜测"。

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
  - project_facts        # 只读：确定性事实和猜测
  - tech_spec            # 技术规范
  - api_definitions      # 接口定义

blackboard_write:
  - project_facts        # 可写：开发中发现的事实
  - project_discoveries  # 可写：新发现
  - progress_updates     # 可写：进展更新

context_injection:
  - project_goals
  - tech_stack
  - coding_standards
```

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
└─────────────────────────────────────────────┘
         ↑ 写入              ↓ 读取
    ┌────┴────┐         ┌────┴────┐
    │ Agent A │         │ Agent B │
    │ (写事实) │         │ (读事实) │
    └─────────┘         └─────────┘
```

### 黑板读写控制矩阵

| 角色 | 层级0 | 层级1 | 层级2 | 层级3 | 层级4 |
|------|-------|-------|-------|-------|-------|
| CEO | RW | RW | RW | R | R |
| HR | R | R | R | R | R |
| Developer | R | R | RW | RW | RW |
| Tester | R | R | RW | RW | RW |
| Designer | R | RW | RW | RW | RW |
| Reviewer | R | RW | R | R | RW |
| Ops | R | R | RW | RW | RW |
| Doc | R | R | R | RW | R |

> R = 只读, RW = 可读写

---

## Agent 运行时架构

### 单个 Agent 的运行循环

```python
async def agent_loop(agent: Agent):
    """每个 Agent 的独立运行循环"""
    while agent.is_active():
        # 1. 从任务队列获取任务
        task = await agent.get_next_task()
        if not task:
            await agent.wait_for_task()
            continue
        
        # 2. 构建上下文 (从黑板读取 + 个人上下文)
        context = await build_context(agent)
        # context = 黑板共享信息 + 角色指令 + 个人工作记忆
        
        # 3. 调用 LLM (使用角色专属 system_prompt + tools)
        response = await llm_call(
            model=agent.model,
            system=agent.system_prompt,
            context=context,
            tools=agent.tools,  # 只能使用自己的工具
            task=task
        )
        
        # 4. 执行工具调用
        for tool_call in response.tool_calls:
            result = await execute_tool(agent, tool_call)
            # 工具执行结果加入 Agent 自己的上下文
        
        # 5. 将发现写入黑板 (如果有)
        if response.blackboard_updates:
            await blackboard.write(
                project_id=agent.project_id,
                updates=response.blackboard_updates,
                author=agent.id
            )
        
        # 6. 更新个人工作记忆
        await agent.update_working_memory(response)
        
        # 7. 如果需要 Review (如代码提交)
        if response.needs_review:
            await submit_for_review(agent, response.artifact)
```

### Agent 间通信规则

```
规则 1: Agent 之间不直接通信
规则 2: 所有知识交换通过黑板
规则 3: 任务分配通过 AgentServer
规则 4: 每个 Agent 只能看到黑板中自己有权限的部分
规则 5: Agent 的个人工作记忆对其他 Agent 不可见
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
│  └ 进展   │  │ └────────┘ └────────┘ └────────┘      ││
│          │  └──────────────────────────────────────┘│
│ ⚙ 设置   │                                          │
│          │  ┌──────────────────────────────────────┐│
│          │  │ Agent 状态                             ││
│          │  │ 🟢 dev-alice  开发中  [任务3]         ││
│          │  │ 🟢 tester-bob 测试中  [任务5]         ││
│          │  │ ⚪ designer-cat 空闲                   ││
│          │  └──────────────────────────────────────┘│
└──────────┴──────────────────────────────────────────┘
```

### 核心 API

```python
# 用户输入
POST /api/projects/{id}/chat
Body: {"message": "给这个项目添加用户认证功能"}

# 项目管理
GET  /api/projects                    # 列出所有项目
POST /api/projects                    # 创建项目
GET  /api/projects/{id}               # 项目详情

# Agent 管理
GET  /api/projects/{id}/agents        # 列出项目下的 Agent
POST /api/projects/{id}/agents/hire   # 手动招聘 Agent

# 黑板操作
GET  /api/projects/{id}/blackboard    # 读取黑板
GET  /api/projects/{id}/facts         # 读取事实
GET  /api/projects/{id}/discoveries   # 读取发现

# 实时通知
WS   /ws/projects/{id}/events         # WebSocket 事件流
```

---

## 项目目录结构

```
athena/
├── README.md
├── PLAN.md                    # 本文件
├── pyproject.toml
├── athena/
│   ├── __init__.py
│   ├── server.py              # FastAPI 入口
│   ├── config.py              # 配置管理
│   │
│   ├── core/                  # 核心模块
│   │   ├── __init__.py
│   │   ├── agent_server.py    # AgentServer (CEO)
│   │   ├── hr_agent.py        # HR Agent
│   │   ├── agent_runtime.py   # Agent 运行时
│   │   ├── agent_loop.py      # Agent 主循环
│   │   └── llm_client.py     # LLM 调用封装
│   │
│   ├── blackboard/            # 黑板系统
│   │   ├── __init__.py
│   │   ├── board.py           # 黑板核心逻辑
│   │   ├── fact_manager.py    # 事实管理 (确定/猜测)
│   │   ├── access_control.py  # 读写权限控制
│   │   └── search.py          # 语义搜索 (ChromaDB)
│   │
│   ├── db/                    # 数据库层
│   │   ├── __init__.py
│   │   ├── database.py        # SQLite 连接管理
│   │   ├── models.py          # 数据模型
│   │   └── migrations/        # 数据库迁移
│   │
│   ├── templates/             # Agent 模板
│   │   ├── developer.yaml
│   │   ├── tester.yaml
│   │   ├── designer.yaml
│   │   ├── reviewer.yaml
│   │   ├── ops.yaml
│   │   └── doc.yaml
│   │
│   ├── tools/                 # 工具定义
│   │   ├── __init__.py
│   │   ├── base.py            # 工具基类
│   │   ├── file_tools.py      # 文件操作工具
│   │   ├── code_tools.py      # 代码相关工具
│   │   ├── test_tools.py      # 测试工具
│   │   ├── git_tools.py       # Git 工具
│   │   ├── design_tools.py    # 设计工具
│   │   └── review_tools.py    # 审查工具
│   │
│   ├── api/                   # API 路由
│   │   ├── __init__.py
│   │   ├── projects.py
│   │   ├── agents.py
│   │   ├── blackboard.py
│   │   ├── chat.py
│   │   └── websocket.py
│   │
│   └── prompts/               # Prompt 模板
│       ├── ceo.md
│       ├── hr.md
│       ├── developer.md
│       ├── tester.md
│       ├── designer.md
│       ├── reviewer.md
│       ├── ops.md
│       └── doc.md
│
├── frontend/                  # Vue 3 前端
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   ├── views/
│   │   │   ├── Dashboard.vue  # 主面板
│   │   │   ├── Project.vue    # 项目详情
│   │   │   └── Blackboard.vue # 黑板查看
│   │   ├── components/
│   │   │   ├── ChatInput.vue
│   │   │   ├── ProjectBoard.vue
│   │   │   ├── AgentStatus.vue
│   │   │   └── FactList.vue
│   │   └── stores/
│   │       └── project.ts
│   └── index.html
│
├── tests/
│   ├── test_blackboard.py
│   ├── test_agent_runtime.py
│   └── test_hr_agent.py
│
└── scripts/
    ├── setup.sh
    └── dev.sh
```

---

## 开发路线图

### Phase 1: 基础骨架 (2 周)

- [ ] 项目脚手架搭建 (FastAPI + SQLite + Vue 3)
- [ ] 数据库 Schema 实现与迁移脚本
- [ ] Agent 运行时核心 (agent_loop.py)
- [ ] LLM Client 封装 (litellm)
- [ ] 基础 API (项目 CRUD + WebSocket)
- [ ] 最简前端 (输入框 + 项目列表)

### Phase 2: 黑板与核心 Agent (2 周)

- [ ] 黑板系统实现 (board.py + fact_manager.py)
- [ ] 事实分级系统 (确定/猜测)
- [ ] AgentServer (CEO) 实现
- [ ] HR Agent 实现 (模板化 Agent 创建)
- [ ] Developer Agent + 基础工具集
- [ ] 上下文注入与隔离机制

### Phase 3: 全角色 Agent + 审查 (2 周)

- [ ] Tester Agent + 测试工具集
- [ ] Designer Agent + 设计工具集
- [ ] Reviewer Agent + 代码审查流程
- [ ] Ops Agent + 部署工具集
- [ ] Doc Agent + 文档工具集
- [ ] 语义搜索集成 (ChromaDB)

### Phase 4: Web 界面 + 打磨 (2 周)

- [ ] 完整 Web 管理界面
- [ ] 实时状态推送 (WebSocket)
- [ ] 项目看板可视化
- [ ] Agent 状态监控
- [ ] 黑板浏览器
- [ ] 安装脚本与文档

### Phase 5: 高级特性 (后续)

- [ ] Agent 自我进化 (参考 Hermes Skills)
- [ ] DAG 任务编排
- [ ] 多项目并行支持
- [ ] 外部 MCP 工具集成
- [ ] 插件系统
- [ ] ABANDON 机制 (防止 Agent 陷入死循环)

---

## 关键设计决策

### Q1: 为什么不用 CrewAI / MetaGPT 作为底层框架？

**A**: 它们不满足核心需求——
- CrewAI: Agent 共享上下文，无持久化，无动态创建
- MetaGPT: 严格 SOP 流水线，Agent 上下文不隔离
- 我们需要的是: 上下文隔离 + 动态招聘 + 黑板持久化 + 事实分级

### Q2: 为什么选 SQLite 而不是 PostgreSQL？

**A**: 开箱即用原则。SQLite 零配置，单文件，用户无需安装数据库服务。后续可通过替换 database.py 支持 PostgreSQL。

### Q3: 为什么每个 Agent 用独立的 LLM 调用而不是共享一个？

**A**: 上下文隔离的核心要求。每个 Agent 必须有独立的 system_prompt + context + tools，共享调用会污染上下文。

### Q4: 如何防止 Agent 写入错误的"确定性事实"？

**A**: 三重保护——
1. Prompt 层: 反复强调只有 100% 确定的事实才能标记为"确定"
2. 审查层: Reviewer Agent 审核黑板写入的事实
3. 升级层: 其他 Agent 可对"确定"事实提出质疑，降级为"猜测"

### Q5: HR Agent 如何决定招聘什么角色？

**A**: 参考 Hermes 的 delegation 机制——
1. AgentServer 分析需求，生成能力需求清单
2. HR Agent 对比现有 Agent 的角色和技能
3. 缺口角色从模板库选择，自动配置工具和上下文
4. 新 Agent 上线后自动注册到项目和黑板

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

---

*本文档是 Athena 项目的初始规划，将随开发进展持续更新。*
