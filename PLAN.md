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

**Athena 是一个像 IT 公司一样运作的 AI Agent 编排系统**——用户（CEO）开箱即用，通过 Web 界面下达项目需求，CEO秘书(AgentServer)接收并交办，系统自动招聘专业 Agent、分配任务、协调推进，每个 Agent 上下文隔离、专人专用，遇到问题立刻开会沟通对齐，确保软件工程每个环节都有专人负责，交付稳定产品。

### 核心设计哲学

Athena 的所有设计决策都源于三条核心哲学：

> **1. Less is More**
>
> 框架做最少的假设和约束。不预设工作流、不硬编码角色、不灌输方法论。每个 Agent 只拿到它职责所需的最小上下文，剩下的自己搞定。系统不替 Agent 做决定，只提供工具和通道。

> **2. 黑板涌现**
>
> 没有中心化的"大脑"在指挥一切。项目的知识、方向、决策，都从每个 Agent 向黑板写入自己的结果中涌现出来——就像侦探办案，每个人在黑板上写下自己调查到的线索，最终汇总成实现路径。黑板是最关键的项目上下文。

> **3. 就像是一个真正的公司那样**
>
> 缺人招人，缺工具找工具，缺思路或遇到问题就找相关负责人开会头脑风暴。项目组内共享各自的结果放到黑板上，每个员工自己维护自己的专属工具和任务上下文。不搞花哨的编排逻辑——一个真正的公司怎么运转，Athena 就怎么运转。
>
> **更深层的目的**："像真公司"不仅仅是一个好玩的隐喻——它让各个 Agent 涌现出额外的集体智能：互相检查、监督对方的工作，减少幻觉的出现。即便出现幻觉也能及时发现和纠正，避免项目偏移。就像一个真正的工程团队，代码审查、测试验证、交叉确认，多双眼睛比一双眼睛更可靠。

### ⚠️ 关于本文档

本文档定位是**设计指南**，不是实现规格书。涉及的技术细节（数据库 schema、API 路由、伪代码、目录结构等）暂不完善，后续根据开发进展迭代。**不需要在这些方面钻牛角尖**——重要的是设计哲学和架构方向对了，具体实现时自然会对齐。

### 初始公司架构

Athena 启动时只需要 3 个核心 Agent + 一套通用员工模板：

| 组件 | 职责 |
|------|------|
| **AgentServer (CEO秘书)** | 与CEO直接对话、获取项目、关键抉择时请示CEO。项目交给项目经理 Agent 拆解。**内置 LLM 调用接口，用于：1)识别CEO要求是传递给项目经理还是HR的；2)根据CEO描述匹配对应项目UUID** |
| **HR Agent** | 感知能力缺口、按模板创建新 Agent、模板不足时组织小组编写新模板。招聘前检查公司规模上限 |
| **项目经理 Agent** | 从 AgentServer 接收项目、拆解需求、分配任务、**验收交付（最高标准：完善性+低bug+高可运行+高性能，依据测试报告+代码+审查结论，循环验证直至完善，避免CEO返工）**。需要招人时立刻和 HR 沟通（详细说明招什么样的人）。**需求可随时更新（如 Hermes steer）** |
| **通用员工模板 (soul.md 等)** | 定义各角色的 system prompt、工具集、黑板权限等 |

### 初始工作流

1. CEO 通过 Web 界面下达需求 → AgentServer (CEO秘书) 接收
2. AgentServer 内置 LLM 识别意图 → 将项目交给对应项目经理 Agent（附带项目UUID）
3. 项目经理拆解需求 → 发现需要招人 → 告知 HR（详细说明招什么样的人）→ HR 检查上限后直接招聘
4. 任何 Agent 发现需要招人都可以告知 HR，HR 检查上限后直接招聘
5. HR 招聘新 Agent → 项目经理分配任务 → 各 Agent 独立工作
6. 任何 Agent 遇到别人领域的问题 → 立刻开会沟通
7. 项目经理验收交付：以最高标准验收，**必须进行需求回溯**（对照 CEO 原始需求逐条确认，回顾黑板所有内容和内容间实现路径）→ 不通过则整改 → 循环验证直至完善 → 避免CEO返工
8. 项目经理验收通过 → 告知 AgentServer → AgentServer 告知CEO
9. CEO认为还需改进 → 通过 AgentServer 向项目经理下发新需求（**并行，不阻塞执行层工作**）

### HR 招聘的模板扩展

当现有模板不足以完成招聘任务时：
1. HR Agent 成立专门的小组
2. 相关任务放入公司黑板数据库 (BOARD.sqlite)
3. 小组编写新方向的 job description 和模板
4. 新模板审核通过后用于招聘新方向员工

---

## 核心设计原则

### 1. 专人专用，一岗一项目一Agent
开发只做开发，测试只做测试，设计只做设计——绝不跨岗。**一个岗位在一个项目中只有一个 Agent**，两个项目至少两个开发 Agent。HR 决定是否需要同项目多 Agent（一般来说一岗一人，控制成本）。

### 2. 上下文隔离，按需共享
上下文完全隔离，非本职上下文控制在最少——Agent 只看到自己职责范围内的事，注意力更集中。需要共享时通过黑板读取和会议沟通，而非把别人的上下文全部灌进来。

设计思路：
- **完全隔离是默认态**：每个 Agent 只注入本职所需的最小上下文（项目目标 + 角色指令 + 个人工作记忆）
- **按需共享是触发态**：遇到跨领域问题时开会，仅传递与问题直接相关的信息（决议、事实），不传递对方的思考过程
- **非本职上下文最小化**：测试 Agent 不看开发调试过程，开发 Agent 不看测试用例细节，各人只关注自己领域
- **共享产出不共享过程**：黑板上的事实和决议是共享的，但每个人得出结论的思考过程是私有的
- **重要辅助知识必须共享**：报错日志、关键判断依据、诊断信息等虽然属于"过程"的一部分，但对他人的判断有直接帮助，必须通过黑板或会议共享

### 3. 黑板涌现
项目目标、进展、确定性事实通过中央黑板共享给所有 Agent。黑板不是简单的数据库——它是项目组集体智慧的涌现空间。每个 Agent 写入自己的结果，最终汇总成项目的实现路径。就像侦探办案：不同侦探在黑板上写下自己调查到的线索，真相从线索中涌现。

### 4. 事实分级与质量保证

确定性事实标记为"确定"（必须 100% 可靠），猜测标记为"猜测"。

**LLM 质量保证机制**：

#### 4.1 确定性事实交叉验证
`certainty=certain` 的黑板写入需要交叉验证：
- 至少 2 个 Agent 确认，或 PM 审批
- 未经交叉验证的"确定"事实标记为 `pending_verification`（待验证）
- 验证通过后升级为 `certain`，验证失败降级为 `conjecture`

#### 4.2 置信度衰减
长时间无人引用或验证的"确定"事实，自动降级为"待验证"：
- 写入时记录 `last_verified_at`
- 超过 N 天（可配置，默认 7 天）无引用 → 自动降级为 `pending_verification`
- 降级后相关 Agent 收到通知，可重新验证或接受降级

#### 4.3 Agent 自省评分机制
每次 LLM 调用后，Agent 需要显式标注自己推理的置信度（0-10 分）：
- 0-3 分：低置信度，可能是猜测或不确定
- 4-6 分：中等置信度，有一定依据但不确定
- 7-9 分：高置信度，有充分依据
- **10 分：完全确定，必须给出推理流程并写入黑板**
- 10 分结果的推理流程对其他 Agent 可见，任何 Agent 都可检查和质疑

**自省评分 prompt 示例**：
```
请对你刚才的推理结果进行置信度评分（0-10分）：

评分标准：
- 0分：完全不确定，纯猜测
- 3分：有初步依据但可能性低
- 5分：有中等依据，可能正确
- 7分：有充分依据，大概率正确
- 9分：有强证据支撑，几乎确定
- 10分：完全确定，有完整推理链路

注意：评分10分的结果必须附带完整的推理流程，写入黑板供其他Agent审查。

评分示例：
- "根据代码第47行的TypeError，parse_body()在Content-Type缺失时返回None" → 9分（有代码证据）
- "我觉得这个bug可能是并发问题" → 3分（没有证据支撑）
- "测试用例全部通过，包括边界条件测试" → 10分（推理：测试覆盖→执行结果→全部通过→结论可靠）
```

#### 4.4 PM 验收需求回溯
PM 验收时必须进行"需求回溯"——对照 CEO 原始需求逐条确认，而非只看代码和测试：
- 验收时回顾黑板的所有内容和内容之间的实现路径
- 确保每个 CEO 需求都有对应的黑板事实链路：需求 → 设计决策 → 实现事实 → 测试验证
- 实现路径断裂或不完整 → 要求补充或返工

> 💡 此验收流程写入 PM 的 soul.md（角色指令），确保每次验收都执行。

### 5. 自动招聘
HR Agent 感知能力缺口时自动创建新 Agent，按需配备工具和新 Agent 职责。可借助互联网获取相关职位的工作需求等信息来辅助定义新角色。

**招聘流程简化**：项目经理发现需要招人 → 告知 HR（详细说明招什么样的人）→ HR 检查公司规模上限 → 上限内直接招聘，无需招聘会议。上限外则上报 AgentServer 向CEO申请扩容。

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
4. 形成决议后，**压缩归档**临时数据库中的讨论内容（保留关键论点和决策依据，删除冗余发言），写入归档记录
5. 决议写入黑板 → 所有相关 Agent 在下次读取黑板时自然获取（无需单独分发会议）
6. 关闭会议

**会议生命周期（5 步）**：
```
1. 发起 → Agent 遇到别人领域的问题，发起会议
2. 讨论 → 参与者发言写入 meeting_messages
3. 形成决议 → 发起人总结写入 meetings.resolution
4. 归档 → 压缩讨论内容（保留决策依据，删除冗余），决议写入黑板 blackboard_entries (category=resolution)
5. 关闭 → 会议状态改为 closed, meeting_{id}.sqlite 可归档/删除
```

> 💡 取消了原来的"决议分发会议"（第7步）——决议写入黑板后，Agent 自然会在下次读取时获取。讨论内容归档而非删除，调试时可回溯"当时为什么做了这个决定"。

### 8. 代码审查（最高标准）
专门的 Review Agent 以最高标准审核所有代码变更：
- 上下文必须和开发、测试等人员隔离
- 每次审查只基于原始代码和原始需求
- **审查维度**：代码正确性、健壮性、性能、安全性、可维护性、边界条件、异常处理
- 有不确定的地方 → 和需求 Agent 及开发 Agent、测试 Agent 商讨
- 商讨仍无法消除不确定 → 汇报 AgentServer (CEO秘书)
- AgentServer 向CEO发起会话，CEO作为公司最高决策者仅进行关键抉择
- **审查不通过 → 要求开发 Agent 修改 → 修改后重新审查 → 循环直至通过**

### 9. 公司规模上限与裁员机制
CEO可定义公司规模上限，HR 招聘前必须检查：
- **人数上限**：如 "最多 100 个 Agent"
- **资源上限**：如 "总内存不超过 16GB"
- 上限配置由 AgentServer 管理，CEO可随时调整
- 达到上限后，HR 暂停招聘，需要扩容时由 AgentServer 向CEO确认

**裁员机制**：当公司资源紧张时，HR 展示 Agent 列表（含闲置时长），CEO 勾选删除即可：
- 优先展示：已完成项目的 Agent、闲置超时的 Agent
- 如果裁员仍不够：HR 建议 CEO 搁置某些现有计划（暂停项目释放资源）
- 实在无法裁员：HR 建议 CEO 增加员工上限

### 10. 工具配备与MCP集成
每个岗位的 Agent 必须配备专属工具集。工具缺失时的解决优先级：

1. **优先：搜索互联网寻找成熟工具** → 下载配置 → 通过 MCP 给对应 Agent 调用（优先使用 stdio 传输，避免占用端口）
2. **其次：公司内部开发** → HR 组织开发小组，自行开发缺失工具
3. **兜底：上报CEO** → 工具卡点无法解决时，告知CEO问题所在

MCP 工具集成原则：
- **stdio 优先**：MCP Server 优先使用 stdio 传输方式，不占用额外端口
- **专人专用**：每个 Agent 只能调用自己角色模板中定义的 MCP 工具
- **按需加载**：工具随 Agent 创建时配置，Agent 销毁时释放

### Term Tool 设计（命令行安全审查机制）

所有 Agent 的命令行执行必须通过 `term` tool，该工具内置 LLM 安全审查：

**执行流程**：
```
1. Agent 调用 term tool，传入要执行的命令
2. term tool 将命令提交给 LLM 安全审查：
   - LLM 评估命令的危害性（是否可能导致不可逆的数据丢失、系统损坏等）
   - 安全审查 prompt 包含危险命令示例，帮助 LLM 识别高风险操作
3. LLM 判定安全 → 执行命令，返回结果
4. LLM 判定危险 → 拒绝执行，返回拒绝原因和修改建议
```

**安全审查 prompt 示例**：
```
你是一个命令行安全审查员。请评估以下命令是否安全执行。

危险命令示例（必须拒绝）：
- rm -rf /          → 删除整个文件系统
- rm -rf /*         → 同上
- format C:         → 格式化磁盘
- dd if=/dev/zero   → 覆盖磁盘数据
- :(){ :|:& };:     → fork 炸弹
- chmod -R 777 /    → 破坏整个系统权限
- > /etc/passwd     → 清空系统用户文件

需要评估的命令: {command}

请回答 SAFE 或 DANGEROUS，如果危险请说明原因和建议的安全替代方案。
```

**审查范围**：
- **需要审查**：所有通过 term tool 执行的命令行操作
- **不需要审查**：MCP Server 的工具调用（开源工具，由 MCP 协议管理）
- **不需要审查**：内置工具（文件读写、代码搜索等，功能有限且可控）

### 11. 开箱即用
CEO安装后即可通过 Web 界面使用，无需复杂配置

### 12. 完全放权，本地优先
Agent 完全运行在本地，给予 Agent 最大自由度。后续可打包 Docker 镜像，但初期以本地裸跑为主。

**命令行安全**：Agent 的命令行执行必须通过 term tool call，该工具内置 LLM 安全审查——每次执行前，LLM 评估命令的危害性，拒绝危险命令（如 `rm -rf /`、`format`、`dd` 等）。开源工具（MCP Server 等）的调用不需要额外审查。

---

## LLM 配置与调用策略

### LLM 接口

- **兼容 OpenAI API 格式**：所有 LLM 调用走 OpenAI 兼容接口（`/v1/chat/completions`）
- **安装时配置**：用户安装 Athena 时，交互式输入 `base_url` 和 `api_key`，写入配置文件
- **不设默认模型**：用户必须自行指定使用的模型名称，系统不内置默认模型
- **Token 限额**：不主动设置 token 限额
  - 429（Rate Limit）处理：**只暂停触发 429 的 Agent**，其他 Agent 继续工作。触发 429 的 Agent 指数退避重试（1s → 2s → 4s → 8s → ...上限 60s）
  - 成本预警：累计 API 费用达到预算 80% 时通知 CEO，达到 100% 时暂停所有 Agent

### 多模型策略

- **默认**：所有 Agent 使用相同模型（安装时配置的那个）
- **支持差异化**：agents 表已有 `model` 字段，不同角色可配置不同模型（如 PM 用强模型，Doc 用便宜模型）
- **当前状态**：写入 TODO，后期再研究（与 Docker 打包同级优先级）

---

## Agent 运行时错误恢复

### 设计原则

参考 Hermes 的错误处理设计，Agent 遇到失败时的处理链路：

1. **工具执行失败** → 首先报错告诉 Agent（让 Agent 知道什么失败了、为什么失败）
2. **Agent 自主寻求替代方案** → 检查是限额了还是能用其他方式替代
   - 例：`wget` 不行就用 `curl`，`pip install` 失败就尝试 `pip3 install`
3. **降级执行** → 如果 LLM 工具调用失败，降级为命令行 + Python 形式执行
   - 例：文件编辑工具失败 → 降级为 `term: sed` 或 `python: open().write()`
4. **LLM 调用超时** → 重试 1 次 → 仍超时则暂停该 Agent，通知 PM
5. **Agent 子进程崩溃** → supervisor 自动重启，PM 收到通知，决定是否重新分配任务

### 崩溃恢复与状态持久化

借鉴 Hermes 的 session 保存机制（实现时需调研 Hermes 具体的 session 序列化逻辑），确保 Agent 状态可恢复：

- **Agent 状态持久化**：`agents` 表增加 `state TEXT` 字段（JSON），记录 Agent 当前正在做什么（当前任务 ID、执行阶段、中间结果）
- **定期 checkpoint**：每隔 N 次 LLM 调用，Agent 将自己的关键状态写入 `agent_contexts`（type=checkpoint）
- **重启恢复**：服务启动时，从数据库恢复 Agent 状态，跳过已完成的任务，从未完成的 checkpoint 处继续
- **任务幂等性**：`agent_tasks` 增加 `idempotency_key TEXT` 字段，防止崩溃后重复执行同一任务

### 错误分级

| 级别 | 场景 | 处理方式 |
|------|------|---------|
| 轻微 | 单次工具调用失败 | Agent 自主重试/替代方案 |
| 中等 | 降级到命令行+Python | Agent 继续工作，记录到 memory.md |
| 严重 | LLM 调用超时/429 | 暂停 Agent，通知 PM |
| 致命 | 子进程崩溃 | supervisor 自动重启，PM 重新分配任务 |

---

## 安全边界与黑板鉴权

### 安全边界

- **无沙箱**：Agent 直接本地运行，文件系统、sudo 等权限不设限制
- **命令行安全**：所有命令行执行必须通过 term tool call，内置 LLM 安全审查机制：
  - 每次执行前，LLM 评估命令的危害性（如 `rm -rf /`、`format`、`dd` 等危险操作会被拒绝）
  - 开源工具（MCP Server 等）的调用不需要额外审查，直接放行
  - term tool prompt 中提供危险命令示例，帮助 LLM 识别高风险操作
- **Docker 打包**：后续考虑，初期不管（写入 TODO）

### 黑板操作鉴权

黑板写入操作通过 tool call 鉴权，确认来源 Agent。写入流程：

1. Agent 调用 `blackboard_write` tool → 鉴权（验证项目参与权 + 角色写权限）
2. 通过鉴权后 → 写入 Go channel（非直接写 SQLite）
3. 单一 writer goroutine 从 channel 批量读取 → commit 到项目对应的 `board_{project_id}.sqlite`
4. 批量 commit 消除多 Agent 同时写入时的锁竞争

| 操作 | 鉴权规则 |
|------|---------|
| **写入** | 只能写入 Agent 参与的项目（通过 project_members 表验证） |
| **删除** | 只能删除自己写入的数据（每条数据有 `author` 字段记录写入者） |
| **修改** | 只能修改自己写入的数据 |
| **读取** | 按角色权限矩阵（见黑板读写控制矩阵） |

### 黑板争议解决

如果有 Agent 认为黑板上其他 Agent 的数据有问题：
1. 质疑方发起会议，邀请数据写入方 + 相关人员（如 Review Agent）
2. 各方在会上展示思考过程和依据（就像现实企业开会讨论）
3. 形成决议：数据正确 / 数据需修正 / 数据降级（如"确定"→"猜测"）
4. 决议写入黑板，相关方按决议执行

**黑板本质**：项目组内部拉平需求和进度、头脑风暴、工程方向提示的工具。就像侦探办案——不同侦探在黑板上写下自己调查到的线索，最后汇总成项目的实现路径。项目的知识、方向、决策，都从每个 Agent 向黑板写入自己的结果中涌现出来。

---

## Agent ID 命名规则

### 命名格式

```
{项目UUID}-{角色}-{编号}
```

- **项目 UUID**：项目唯一标识，创建项目时自动生成
- **角色**：dev / test / review / design / ops / doc 等
- **编号**：同项目同角色的序号（从 1 开始）

**示例**：
- `a1b2c3-dev-1`：项目 a1b2c3 的第 1 个开发 Agent
- `a1b2c3-dev-2`：项目 a1b2c3 的第 2 个开发 Agent
- `a1b2c3-review-1`：项目 a1b2c3 的第 1 个审查 Agent
- `d4e5f6-dev-1`：项目 d4e5f6 的第 1 个开发 Agent

### 命名权

- **所有职位的招聘和命名都由 HR 决定**
- HR 创建 Agent 时生成 ID 和名称，写入 HR 的 soul（记忆）
- HR 的 soul 中记录：已招聘的 Agent 列表、项目-角色映射、命名历史

### 数据库关联

- projects 表的 `id` 字段使用 UUID
- agents 表的 `id` 字段使用 `{UUID}-{role}-{N}` 格式
- agent_tasks / project_members 等表通过 agent_id 和 project_id 关联

---

## AgentServer 意图识别与项目 UUID 机制

### AgentServer 内置 LLM 调用接口

AgentServer 作为CEO秘书，需要理解CEO的泛化意图，内置一套 LLM 调用接口用于：

1. **意图路由**：CEO说了一句话 → 判断是给项目经理的还是给HR的
   - 例："这个图书管理系统的搜索功能太慢了" → 路由给对应项目经理
   - 例："我们需要更多开发人员" → 路由给HR
   - 例："那个项目做好了吗？" → 路由给对应项目经理

2. **项目匹配**：CEO无需提及具体项目名或UUID → LLM 根据描述匹配项目
   - 遍历数据库中已完成和未完成的项目
   - 每个项目的 `projects` 表包含 `original_requirement`（CEO原始需求）和 `requirement_summary`（PM梳理后的需求摘要）
   - LLM 比对CEO描述与这两个字段 → 返回最匹配的项目 UUID

### 项目 UUID 机制

- CEO给出项目需求后，系统自动生成项目 UUID
- UUID 是项目在数据库中的唯一标识，所有下游传递以 UUID 为准
- AgentServer 向项目经理传达时附带 UUID
- 示例：用户输入"上次的书籍管理系统做好了吗？" → LLM 查询匹配到"图书管理系统"→ 返回该项目的 UUID

### 需求更新（Steer 模式）

CEO可随时通过 AgentServer 向项目经理提新要求，**并行不阻塞**：
- 执行层员工工作时，不影响CEO通过 AgentServer 向项目经理提要求
- AgentServer 接到CEO的新需求 → LLM 识别项目 UUID + 意图 → 传递给对应项目经理
- 项目经理收到后更新黑板，实时通知相关 Agent

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

项目经理 Agent 已加入角色体系和黑板权限矩阵。职责：从 AgentServer 接收项目、拆解需求、分配任务、参加招聘会议、**验收交付（高性能+无bug）**、**需求随时更新（steer模式）**。

### ✅ 待确认 5: AgentServer = CEO秘书 — 已确认

AgentServer 就是CEO秘书，与CEO直接对话，拿到项目后交给项目经理拆解。全文已统一称呼为 "AgentServer (CEO秘书)"。

---

## 与现有系统的对比

| 特性 | Hermes | MetaGPT | CrewAI | Agent-Blackboard | Eino | **Athena** |
|------|--------|---------|--------|-----------------|------|------------|
| **组织模型** | 单 Agent + Skills | 公司流水线 SOP | Crew 小组 | 黑板协调 | DeepAgent 子Agent委派 | **公司架构 + 黑板涌现** |
| **上下文隔离** | ❌ 单一上下文 | ⚠️ 部分隔离 | ⚠️ 部分隔离 | ✅ 完全隔离 | ⚠️ 子Agent隔离 | ✅ 完全隔离+按需共享 |
| **Agent 动态创建** | ✅ delegate_task | ❌ 预定义角色 | ❌ 预定义角色 | ⚠️ 注册制 | ✅ DeepAgent | ✅ HR 自动招聘 |
| **知识持久化** | ✅ SQLite | ❌ 无持久化 | ❌ 内存态 | ✅ SQLite + 语义搜索 | ✅ Checkpoint | ✅ SQLite + FTS5 全文搜索 |
| **专人专用** | ❌ 共享工具集 | ⚠️ 角色分工但上下文共享 | ✅ 角色分工 | ✅ 领域专家 | ✅ Tool 接口隔离 | ✅ 专人专用 + 工具隔离 |
| **事实分级** | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ✅ 确定/猜测分级 |
| **代码审查** | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ❌ 无 | ✅ 专门 Review Agent |
| **Agent 间沟通** | ❌ 无机制 | ⚠️ 流水线传递 | ⚠️ 委托式 | ❌ 无直接沟通 | ⚠️ SubAgent 通信 | ✅ 会议系统 + 黑板决议 |
| **Web 管理界面** | ❌ CLI | ❌ API | ❌ API | ❌ 无 | ❌ 无 | ✅ 内置 Web UI |
| **黑板写入** | ❌ 无黑板 | ⚠️ 隐式共享 | ⚠️ 隐式共享 | ✅ 显式黑板 | ❌ 无黑板 | ✅ 显式黑板 |
| **SOP 工作流** | ❌ 自由式 | ✅ 严格 SOP | ⚠️ 流程可选 | ❌ 自由式 | ✅ Graph 编排 | ✅ 可配置工作流 |
| **中断恢复** | ⚠️ Session 保存 | ❌ 无 | ❌ 无 | ❌ 无 | ✅ Checkpoint | ✅ Checkpoint + 自研状态持久化 |
| **LLM 多提供商** | ⚠️ 有限 | ⚠️ 有限 | ✅ LangChain 生态 | ❌ 单一 | ✅ ChatModel 统一接口 | ✅ 基于 Eino ChatModel |

### Athena 的独特价值

Athena = **Eino (Go AI 框架)** + **MetaGPT 的公司角色模型(主要架构参考)** + **Agent-Blackboard 的黑板模式** + **Hermes 的自省记忆与任务拆解(主要交互参考)** + **独创的 HR 动态招聘 + 事实分级 + 上下文完全隔离按需共享 + 会议沟通机制**

**核心哲学**：Less is More（框架最少约束）、黑板涌现（集体智慧从贡献中涌现）、就像一个真正的公司那样（缺人招人、缺工具找工具、缺思路开会）。

**技术策略**：站在 Eino 的肩膀上，不重复造轮子（LLM 调用、Agent 循环、工具接口、中断恢复），专注 Athena 的独创价值（黑板、会议、HR、公司模型、事实分级）。

---

## 技术栈选型

| 层次 | 技术 | 理由 |
|------|------|------|
| **后端语言** | Go (Golang) | 高性能、并发原生支持、单二进制部署、类型安全。MCP 官方提供 Go SDK：https://modelcontextprotocol.io/docs/sdk |
| **后端框架** | Gin / Echo | Go 主流 Web 框架，性能优秀 |
| **AI 框架** | **Eino** (github.com/cloudwego/eino) | 字节跳动 CloudWeGo 团队开源，Go 原生设计，内置 ReAct Agent + Checkpoint 中断恢复 + 事件流。详见下方「Go AI 框架选型」 |
| **数据库** | SQLite (go-sqlite3) + FTS5 | SQLite 存储结构化数据，FTS5 提供全文搜索（替代 ChromaDB，保持零外部依赖） |
| **前端** | Vue 3 + Vite | 轻量、现代、组件化好 |
| **LLM 调用** | Eino ChatModel 组件 | 使用 Eino 的 ChatModel 抽象层，内置 OpenAI/Claude/Gemini/Ollama 等提供商适配，Athena 扩展 GLM 等国产模型 |
| **Agent 内循环** | Eino ChatModelAgent (ReAct) | 使用 Eino 内置的 ReAct 循环作为每个 Agent 子进程的内循环，Athena 在外层管理子进程生命周期 |
| **Agent 外层运行时** | 自研 subprocess-based Supervisor | 每个 Agent 独立子进程运行 Eino Agent，主进程作为 supervisor 管理生命周期。子进程崩溃不影响其他 Agent |
| **任务队列** | Go channel + SQLite | 轻量级，channel 做实时调度，SQLite 做持久化 |
| **通信协议** | WebSocket + REST | WebSocket 实时推送，REST 管理操作 |
| **会议系统** | 每会议独立 SQLite 文件 | 会议数据独立于黑板，每个会议一个 .sqlite，会后可选归档或删除 |

### Go AI 框架选型

Go 生态目前有三个主流 AI Agent 框架可选，Athena 选择 **Eino** 作为基础框架：

| 维度 | LangChainGo | Eino ⭐选用 | Google ADK-Go |
|------|-------------|-----------|---------------|
| **来源** | 社区（LangChain Go 移植） | 字节跳动 CloudWeGo | Google 官方 |
| **设计理念** | Python LangChain 照搬 | Go 原生设计，借鉴 LangChain+LlamaIndex | Google Cloud 体系优先 |
| **版本** | v0.1.14 (2025.10) | **v0.8.13** (2026.4, 178 releases) | v1.2.0 (2026.4) |
| **Stars** | 9.2k | **11k** | 7.7k |
| **Agent 内循环** | 基础 | ✅ **ReAct 内置** | Multi-agent 编排 |
| **多 Agent 协调** | 有限 | ✅ **DeepAgent (子Agent委派)** | ✅ A2A 协议 |
| **中断恢复** | ❌ | ✅ **Checkpoint + Interrupt/Resume** | ✅ Session 管理 |
| **流处理** | 基础 | ✅ **自动流拼接/合并/复制** | 基础 |
| **图编排** | Chain 链式 | ✅ **Graph + GraphTool (图即工具)** | 无 |
| **LLM 提供商** | OpenAI/Gemini/Ollama | **OpenAI/Claude/Gemini/Ollama/Ark** | Gemini 深度集成 |
| **中文社区** | 弱 | ✅ **强（字节跳动+中文文档）** | 弱 |
| **回调切面** | 基础 Callback | ✅ **OnStart/OnEnd/OnError/Stream** | 基础 |
| **厂商锁定** | 无 | 无 | ⚠️ Google Cloud 倾向 |

**为什么选 Eino？**

1. **Go 原生设计**：不是 Python 照搬，API 惯例符合 Go 风格，写起来地道
2. **ReAct 内置**：Athena 每个 Agent 子进程的核心循环就是 ReAct，开箱即用
3. **Checkpoint 中断恢复**：与 Athena 的崩溃恢复需求（借鉴 Hermes session）天然对齐，Eino 已实现状态持久化和恢复
4. **DeepAgent 子Agent 委派**：PM Agent → 员工 Agent 的任务分配模式可映射为 DeepAgent 的 SubAgent 机制
5. **GraphTool (图即工具)**：将确定性工作流暴露为 Agent 可调用的工具，桥接 PM 排序分配和 Agent 自主决策
6. **事件流自动处理**：Agent 的输出以事件迭代器模式消费，框架自动处理流式拼接/合并
7. **回调切面**：OnStart/OnEnd/OnError 切面注入日志、追踪、指标，Athena 可在切面中实现审计日志
8. **社区活跃度高**：178 个 release，字节跳动持续投入，中文文档和社区完善
9. **组件抽象分离**：核心仓库只定义抽象，实现在 eino-ext，Athena 可实现自定义组件（如 GLM ChatModel）

**Eino 不替代什么（Athena 仍需自研）**：

| Athena 自研模块 | 原因 |
|----------------|------|
| 黑板系统 (SQLite + FTS5) | Eino 无黑板概念，这是 Athena 的核心创新 |
| 会议系统 | Eino 无会议机制，Athena 独创的结构化沟通 |
| HR 招聘/裁员 | Eino 无动态 Agent 创建/销毁 |
| 子进程 Supervisor | Eino Agent 在进程内运行 goroutine，Athena 需要子进程隔离 |
| 事实分级 + 质量保证 | Eino 无事实分级概念 |
| MCP 工具集成 | 使用 Go MCP SDK，Eino Tool 接口做桥接 |
| 公司模型编排 | Athena 独有的公司化组织架构 |
| Term Tool 安全审查 | Athena 独创的 LLM 命令行安全审查 |
| Web UI (Vue 3) | 前端完全自研 |

**Athena + Eino 的分层架构**：

```
┌─────────────────────────────────────────────────┐
│  Athena 自研层（公司化编排 + 黑板 + 会议 + HR）   │
│  ├─ 黑板系统 (board.go + channel + FTS5)        │
│  ├─ 会议系统 (每会议独立 SQLite)                 │
│  ├─ HR 系统 (模板化招聘 + 裁员)                  │
│  ├─ 子进程 Supervisor (进程级隔离)               │
│  ├─ 事实分级 + 质量保证                          │
│  ├─ MCP 工具集成 (Go SDK + Eino Tool 桥接)      │
│  ├─ Term Tool (LLM 安全审查)                    │
│  └─ Web UI + API (Gin + Vue 3)                 │
├─────────────────────────────────────────────────┤
│  Eino 框架层（AI Agent 基础能力）                 │
│  ├─ ChatModel → 统一 LLM 调用 (多提供商)        │
│  ├─ ChatModelAgent → ReAct 内循环               │
│  ├─ Checkpoint → Agent 状态持久化 + 中断恢复     │
│  ├─ Tool 接口 → 内置工具 + MCP 桥接             │
│  ├─ Callback → 日志/追踪/审计切面               │
│  └─ Stream → 事件流自动处理                      │
└─────────────────────────────────────────────────┘
```

**关键映射：Eino 组件 → Athena 用途**

| Eino 组件 | Athena 用途 |
|-----------|------------|
| `ChatModel` | Agent 的 LLM 调用层，替换"自研 Go LLM Client" |
| `ChatModelAgent` | Agent 子进程的 ReAct 内循环 |
| `DeepAgent` | PM Agent 的多 Agent 协调（SubAgent 委派） |
| `Tool` / `BaseTool` | 内置工具 + MCP 工具的统一接口 |
| `GraphTool` | PM 工作流编排暴露为工具给 Agent 调用 |
| `Checkpoint` | 崩溃恢复的底层机制（映射到 agents.state JSON） |
| `Callback (OnStart/OnEnd/OnError)` | 审计日志、性能追踪、成本统计的切面 |
| `compose.Graph` | PM Agent 的确定性任务编排图 |
| `compose.ToolsNode` | 工具调用节点，集成 term tool 等 |

**子进程中的 Eino 使用模式**：

```go
// athena-agent 子进程内部（由 supervisor 启动）
func main() {
    // 1. 解析命令行参数（agent ID, role, project 等）
    agentID := flag.String("id", "", "Agent ID")
    role := flag.String("role", "", "Agent role")
    projectID := flag.String("project", "", "Project ID")

    // 2. 使用 Eino ChatModel 创建 LLM 客户端
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL: config.LLMBaseURL,
        APIKey:  config.LLMAPIKey,
        Model:   config.LLMModel,
    })

    // 3. 注册 Athena 自定义工具（通过 Eino Tool 接口）
    tools := []tool.BaseTool{
        NewBlackboardReadTool(agentID, projectID),   // 黑板读取
        NewBlackboardWriteTool(agentID, projectID),   // 黑板写入
        NewTermTool(agentID),                         // 命令行执行 + LLM 安全审查
        NewMeetingTool(agentID, projectID),           // 会议发言/参与
        // MCP 工具通过 Tool 接口桥接...
    }

    // 4. 使用 Eino ChatModelAgent 创建 ReAct 循环
    agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Model: chatModel,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: tools,
            },
        },
    })

    // 5. 通过 stdin/stdout 与主进程通信（接收任务、上报结果）
    // 主循环：读 stdin → 构建上下文 → Runner.Query → 写 stdout
    RunAgentWithIO(agent, os.Stdin, os.Stdout)
}
```

---

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    Web UI (Vue 3)                        │
│  ┌──────────────┐  ┌───────────────────┐  ┌──────────┐  │
│  │ 项目看板     │  │ Agent 状态监控等   │  │ 输入框    │  │
│  └──────────────┘  └───────────────────┘  └──────────┘  │
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
│  │  │ - LLM意图识别: CEO要求→PM或HR              │  │  │
│  │  │ - LLM项目匹配: CEO描述→项目UUID             │  │  │
│  │  └──────────────────────────────────────────────┘  │  │
│  │  ┌──────────────┐  ┌───────────────────────────┐   │  │
│  │  │ HR Agent     │  │ 项目经理 Agent             │   │  │
│  │  │ - 感知人才缺口│  │ - 拆解细化需求             │   │  │
│  │  │ - 创建 Agent │  │ - 分配任务                 │   │  │
│  │  │ - 配备工具   │  │ - 处理不确定性上报         │   │  │
│  │  │ - 扩展模板   │  │ - 决策：自动调研或上报CEO │   │  │
│  │  │              │  │ - 验收交付(最高标准)     │   │  │
│  │  │              │  │ - 需求随时更新(steer模式)  │   │  │
│  │  └──────────────┘  └───────────────────────────┘   │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │           执行层 (Workers) — 每个 Agent 独立子进程      │
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
│  │                    ↕ 语义搜索                        │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │ SQLite FTS5 (全文搜索)                        │   │  │
│  │  │ - blackboard_entries 全文索引                  │   │  │
│  │  │ - 写入时自动更新 FTS 索引                      │   │  │
│  │  │ - 支持关键词 + BM25 排序搜索                   │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │              SQLite (项目级拆分)                     │  │
│  │  - board_{project_id}.sqlite (每项目独立黑板数据库) │  │
│  │  - meeting_{id}.sqlite (每会议独立数据库)          │  │
│  │  - 写入走 Go channel + 批量 commit                │  │
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
    original_requirement TEXT,  -- CEO原始需求输入（AgentServer UUID匹配的数据源）
    requirement_summary TEXT,   -- 项目经理梳理后的需求摘要（AgentServer UUID匹配的数据源）
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
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture', 'pending_verification')),
    author      TEXT,       -- 写入者 Agent ID（与黑板鉴权的 author 对应）
    evidence    TEXT,       -- 支撑证据
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_project_facts_project ON project_facts(project_id);
CREATE INDEX idx_project_facts_certainty ON project_facts(certainty);
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
    role        TEXT NOT NULL,  -- ceo_secretary/hr/pm/developer/tester/designer/reviewer/ops/doc
    status      TEXT DEFAULT 'idle',  -- idle/working/in_meeting/offline
    tools       TEXT,           -- JSON: 可用工具列表 (含MCP工具)
    mcp_servers TEXT,           -- JSON: MCP Server 配置 (优先stdio)
    model       TEXT DEFAULT 'default',
    state       TEXT,           -- JSON: Agent 当前状态（崩溃恢复用，记录当前任务ID、执行阶段、中间结果）
    created_by  TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 注：agent 与 project 的关联通过 project_members 表管理，不在此表冗余存储
CREATE INDEX idx_agents_role ON agents(role);
CREATE INDEX idx_agents_status ON agents(status);
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

CREATE INDEX idx_project_members_project ON project_members(project_id);
CREATE INDEX idx_project_members_agent ON project_members(agent_id);
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
    depends_on  TEXT,           -- 逗号分隔的 task_id 列表，表示任务依赖
    idempotency_key TEXT,       -- 任务幂等键，防止崩溃后重复执行
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE INDEX idx_agent_tasks_project ON agent_tasks(project_id);
CREATE INDEX idx_agent_tasks_agent ON agent_tasks(agent_id);
CREATE INDEX idx_agent_tasks_status ON agent_tasks(status);
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

> 注：黑板数据库按项目拆分为 `board_{project_id}.sqlite`，每个项目独立数据库文件，消除多项目间的写锁竞争。写入操作走 Go channel + 批量 commit，Agent 写入 channel → 单一 writer goroutine 批量 commit 到 SQLite。

```sql
CREATE TABLE blackboard_entries (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL,  -- 逻辑关联（同一 SQLite 文件内，无需外键）
    category    TEXT NOT NULL,  -- goal/fact/discovery/decision/progress/resolution/auxiliary
    content     TEXT NOT NULL,
    certainty   TEXT NOT NULL CHECK(certainty IN ('certain', 'conjecture', 'pending_verification')),
    author      TEXT,
    confidence_score INTEGER,   -- Agent 自省评分 0-10（10分必须附带推理流程）
    reasoning   TEXT,           -- 10分结果的推理流程（其他 Agent 可检查）
    last_verified_at DATETIME,  -- 最后验证时间（用于置信度衰减）
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- FTS5 全文搜索索引（替代 ChromaDB 语义搜索）
CREATE VIRTUAL TABLE blackboard_entries_fts USING fts5(
    content,
    category,
    author,
    content=blackboard_entries,
    content_rowid=rowid,
    tokenize='unicode61'  -- 支持中文分词
);

-- 触发器：写入时自动更新 FTS 索引
CREATE TRIGGER blackboard_entries_ai AFTER INSERT ON blackboard_entries BEGIN
    INSERT INTO blackboard_entries_fts(rowid, content, category, author)
    VALUES (new.rowid, new.content, new.category, new.author);
END;
CREATE TRIGGER blackboard_entries_ad AFTER DELETE ON blackboard_entries BEGIN
    INSERT INTO blackboard_entries_fts(blackboard_entries_fts, rowid, content, category, author)
    VALUES ('delete', old.rowid, old.content, old.category, old.author);
END;
```

#### 10. meetings — 会议表 (每会议独立 meeting_{id}.sqlite)

```sql
CREATE TABLE meetings (
    id              TEXT PRIMARY KEY,
    project_id      TEXT NOT NULL,
    convener_id     TEXT NOT NULL,      -- 发起人 Agent ID
    status          TEXT DEFAULT 'open', -- open/resolved/closed
    resolution      TEXT,               -- 会议决议内容
    archived_discussion TEXT,           -- 压缩归档的讨论内容（保留关键论点和决策依据）
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at       DATETIME
);
```

#### 11. meeting_participants — 会议参与者表 (meeting_{id}.sqlite)

> 注：此表位于独立会议 SQLite 文件内，meeting_id 冗余但保留用于跨会议查询场景。

```sql
CREATE TABLE meeting_participants (
    id          TEXT PRIMARY KEY,
    meeting_id  TEXT NOT NULL,
    agent_id    TEXT NOT NULL,
    role        TEXT DEFAULT 'participant', -- convener/participant
    UNIQUE(meeting_id, agent_id)
);
```

#### 12. meeting_messages — 会议发言表 (meeting_{id}.sqlite)

> 注：此表位于独立会议 SQLite 文件内。meeting_id 保留用于逻辑关联（跨库无物理外键约束）。

```sql
CREATE TABLE meeting_messages (
    id          TEXT PRIMARY KEY,
    meeting_id  TEXT NOT NULL,
    speaker_id  TEXT NOT NULL,      -- 发言人 Agent ID
    speaker_role TEXT NOT NULL,     -- 发言人岗位
    content     TEXT NOT NULL,      -- 发言内容
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 13. audit_log — 审计日志表

```sql
CREATE TABLE audit_log (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL,      -- 执行操作的 Agent ID
    action      TEXT NOT NULL,      -- 操作类型: blackboard_write/task_update/meeting_resolution/agent_create/...
    target_type TEXT NOT NULL,      -- 操作对象类型: blackboard_entry/task/meeting/agent/...
    target_id   TEXT NOT NULL,      -- 操作对象 ID
    details     TEXT,               -- JSON: 操作详情
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_agent ON audit_log(agent_id);
CREATE INDEX idx_audit_log_target ON audit_log(target_type, target_id);
CREATE INDEX idx_audit_log_time ON audit_log(created_at);
```

---

## Agent 角色体系与工具分配

### 角色 — 工具映射（专人专用！）

| 角色 | 职责 | 专属工具集 | 上下文内容 |
|------|------|-----------|-----------|
| **AgentServer (CEO秘书)** | 与CEO直接对话、获取项目、关键抉择请示CEO、项目交给项目经理 | 项目管理工具、HR 调度工具、CEO交互、**LLM意图识别** | 全局视角 |
| **HR Agent** | 感知缺口、创建/销毁 Agent、分配工具、扩展模板、裁员方案 | Agent 模板库、工具注册表、互联网搜索、配置生成器 | 公司组织架构 |
| **项目经理 Agent** | 拆解细化需求、分配任务、处理不确定性上报、**验收交付（最高标准）** | 任务分解工具、需求分析工具 | 项目目标 + 需求文档 |
| **Developer Agent** | 代码开发 | 文件读写、代码执行、Git、Debug 工具、**term (命令行执行+LLM安全审查)** | 项目目标 + 技术上下文 + 自己的历史 |
| **Tester Agent** | 测试、出具测试报告（md/html格式） | 测试框架、覆盖率工具、断言工具 | 项目目标 + 接口定义 + 自己的历史 |
| **Designer Agent** | 架构/接口设计 | 画图工具、API 设计工具 | 项目目标 + 需求文档 + 自己的历史 |
| **Reviewer Agent** | 代码审查（最高标准，上下文隔离） | 文件读取、Linter、Diff 工具 | 原始代码 + 原始需求（不包含开发者的思考过程） |
| **Ops Agent** | 部署/运维 | Docker、CI/CD、监控工具、**term (命令行执行+LLM安全审查)** | 项目目标 + 环境配置 + 自己的历史 |
| **Doc Agent** | 文档编写 | 文件读写、模板工具 | 项目目标 + 代码文档 + 自己的历史 |

### 上下文注入策略（简化视图）

> 💡 以下是上下文注入的**逻辑分组**（5层），实际 System Prompt 拼装见「Hermes Prompt 组装借鉴 → Athena Agent System Prompt」的**8层完整结构**。两者是同一机制的不同粒度描述：5层是逻辑归类，8层是运行时拼装顺序。

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

任何 Agent 都可以向 HR 提出招聘请求：
- 项目经理拆解需求时发现需要新角色 → 告知 HR（详细说明招什么样的人）
- 任何 Agent 遇到自己领域外的问题且无对应同事 → 找 HR 招人
- HR 检查公司规模上限 → 上限内直接招聘，上限外上报 AgentServer 向CEO申请扩容

### 招聘上限检查

HR 招聘前必须检查公司规模上限（由CEO通过 AgentServer 设定）：
- 当前 Agent 数量 < 人数上限
- 预计新增内存 < 资源上限
- 达到上限 → HR 暂停招聘，通知 AgentServer 向CEO申请扩容

### 招聘流程

```
1. 任意 Agent 发现缺少某种能力 → 向 HR 提出招聘请求:
   {
     "requester_id": "a1b2c3-test-1",
     "requester_role": "tester",
     "reason": "项目需要Go后端开发，目前没有开发Agent",
     "role": "developer",
     "project_id": "a1b2c3",
     "required_skills": ["go", "gin", "sqlite"],
     "context_requirements": ["project_goals", "tech_stack"]
   }
2. HR 评估招聘请求 → 检查公司规模上限 → 超限则上报 AgentServer 向CEO申请扩容
3. HR Agent 执行招聘:
   a. 从 Agent 模板库选择对应角色模板
   b. 如果模板不存在 → 成立小组编写新模板 (任务写入项目黑板数据库)
   c. 生成唯一 Agent ID 和名称
   d. 配置专属工具集 (根据 role + required_skills)
   e. 可借助互联网搜索该职位的工作需求，辅助定义 Agent 职责
   f. 注入初始上下文 (项目目标 + 角色 soul.md)
   g. 注册到 agents 表 + project_members 表
   h. 启动 Agent 运行时 (独立子进程，由 supervisor 管理)
4. 新 Agent 上线，项目经理分配任务
```

### Agent 模板结构 (YAML)

```yaml
# templates/developer.yaml
role: developer
name_template: "{project_uuid}-dev-{N}"
model: default
system_prompt: |
  你是一名专业的软件开发工程师。你的职责是编写高质量、可维护的代码。
  你只负责开发，不负责测试、设计或审查。
  遵循项目的技术规范和编码标准。
  将你的工作进展写入黑板，将发现的事实标记为"确定"或"猜测"。
  遇到别人领域的问题（如测试bug、设计疑问），立刻找对应Agent开会对齐，绝不自己琢磨。
  你想要在会议中发言时，利用发言tool将发言写入会议临时数据库。
  完成复杂工作后，将学到的经验保存为技能(skills/)。
  发现重要事实、工具技巧、项目约定等你认为有助于自身工作的重要内容，则写入个人记忆(memory.md)。

tools:
  - file_read
  - file_write
  - file_edit
  - term            # 命令行执行工具，内置 LLM 安全审查（拒绝 rm -rf / 等危险命令）
  - git
  - code_search
  - debug

mcp_servers:           # MCP 工具 (stdio 优先)
  - name: filesystem
    transport: stdio
    command: "mcp-filesystem"
  - name: database
    transport: stdio
    command: "mcp-sqlite"

blackboard_read:
  - project_goals
  - project_facts
  - tech_spec
  - api_definitions
  - meeting_resolutions
  - error_logs          # 重要辅助知识：报错日志必须共享

blackboard_write:
  - project_facts
  - project_discoveries
  - progress_updates
  - error_logs          # 开发发现的关键报错写入黑板供测试等查看

context_injection:
  - project_goals
  - tech_stack
  - coding_standards
```

### 裁员流程

```
1. 触发条件: 公司资源紧张 / 项目完成 / Agent 闲置超时
2. HR 生成裁员方案:
   a. 扫描已完成项目的 Agent → 标记为可释放
   b. 扫描闲置超时的 Agent → 标记为可释放
   c. 计算释放后资源占用 → 是否满足需求
3. HR 将方案提交 AgentServer → 展示给CEO
4. CEO 在前端界面选择:
   - ✅ 同意裁员: 释放选中的 Agent
   - ⏸️ 搁置项目: 暂停某项目，释放其全部 Agent
   - ⬆️ 增加上限: 扩大公司规模限制
5. 执行: HR 通知 supervisor 终止 Agent 子进程、清理上下文、从数据库注销
```

### 工具配备流程

```
1. Agent 创建时 → 按角色模板分配基础工具
2. Agent 工作中发现工具缺失 → 上报项目经理
3. 项目经理评估 → 确认需要新工具
4. 解决优先级:
   a. 🔍 搜索互联网 → 找到成熟工具 → 下载配置 → MCP 注册 (stdio 优先)
   b. 🛠️ 公司内部开发 → HR 组织开发小组 → 自行开发缺失工具
   c. 📢 上报CEO → 工具卡点无法解决，告知CEO问题所在
5. 工具就绪 → HR 更新 Agent 的工具配置 → Agent 可调用
```

---

## 会议系统设计

会议系统是 Athena 独创的 Agent 间结构化沟通机制，与黑板模式互补：

- **黑板** → 持久化知识（事实、进展、决议）
- **会议** → 实时沟通（讨论、答疑），仅在遇到问题时发起

### 会议存储设计

每个会议使用独立的 SQLite 文件 (`meeting_{id}.sqlite`)，与黑板数据库完全分离：
- 会议结束后，决议写入黑板，讨论内容**压缩归档**（保留关键论点和决策依据，删除冗余发言）
- 会议 SQLite 文件可归档或删除，不影响黑板数据
- 独立文件设计便于并行开会、生命周期管理
- 调试时可回溯归档的讨论内容（"当时为什么做了这个决定"）

### 会议生命周期

```
1. 发起 → Agent 遇到别人领域的问题，发起会议，创建 meeting_{id}.sqlite
2. 讨论 → 参与者发言写入 meeting_messages
3. 形成决议 → 发起人总结写入 meetings.resolution，压缩归档讨论内容
4. 写入黑板 → 决议写入 blackboard_entries (category=resolution)，相关 Agent 下次读取时自然获取
5. 关闭 → 会议状态改为 closed, meeting_{id}.sqlite 可归档/删除
```

> 💡 5 步而非原来的 9 步。取消了"决议分发会议"——决议写入黑板后，Agent 自然会在下次读取时获取。讨论内容归档压缩而非删除。

---

## 黑板模式设计（参考 TCH Bytex + Agent-Blackboard）

### 黑板架构

黑板是 Athena 的核心知识共享机制，参考了：
- **TCH Bytex 方案**: 黑板 + DAG 的多 Agent 协调
- **Agent-Blackboard 项目**: MCP 持久化 + 语义搜索 + 本体约束
- **经典黑板模式**: HEARSAY-II 的知识源 + 控制组件

**黑板存储**：按项目拆分为独立 SQLite 文件 (`board_{project_id}.sqlite`)，消除多项目间的写锁竞争。写入操作走 Go channel + 批量 commit。

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
│  │ 层级 4.5: 辅助知识 (重要判断依据)        │ │
│  │ - 报错日志、堆栈信息                     │ │
│  │ - 关键诊断数据                           │ │
│  │ - 环境配置信息                           │ │
│  │ ⚠️ 属于过程但必须共享，辅助他人判断      │ │
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

| 角色 | 层级0 | 层级1 | 层级2 | 层级3 | 层级4 | 层级4.5(辅助) | 层级5 |
|------|-------|-------|-------|-------|-------|-------------|-------|
| AgentServer (CEO秘书) | RW | RW | RW | R | R | R | R |
| HR | R | R | R | R | R | R | R |
| 项目经理 | RW | RW | RW | RW | RW | RW | R |
| Developer | R | R | RW | RW | RW | RW | R |
| Tester | R | R | RW | RW | RW | RW | R |
| Designer | R | RW | RW | RW | RW | R | R |
| Reviewer | R | RW | R | R | RW | RW | R |
| Ops | R | R | RW | RW | RW | RW | R |
| Doc | R | R | R | RW | R | R | R |

> R = 只读, RW = 可读写。层级4.5(辅助知识): 报错日志等对判断有直接帮助的过程信息，必须共享

---

## Agent 运行时架构

### 单个 Agent 的运行循环 (子进程模型 + Eino)

每个 Agent 运行在独立的子进程中，子进程内部使用 Eino 的 ChatModelAgent 作为 ReAct 内循环。Athena Server（主进程）作为 supervisor 管理所有 Agent 子进程的生命周期：

```go
// Supervisor 管理所有 Agent 子进程
type AgentSupervisor struct {
    agents    map[string]*AgentProcess  // agent_id → 子进程
    mu        sync.RWMutex
    cmdChan   chan BlackboardWrite      // 黑板写入 channel（批量写入消除锁竞争）
}

// AgentProcess 封装单个 Agent 子进程
type AgentProcess struct {
    ID        string
    Cmd       *exec.Cmd
    Stdin     io.WriteCloser    // 主进程 → Agent（下发任务、会议邀请等）
    Stdout    io.ReadCloser     // Agent → 主进程（工具调用、黑板写入等）
    Status    AgentStatus
    RestartCount int
}

// 启动 Agent 子进程
func (s *AgentSupervisor) StartAgent(ctx context.Context, agent *Agent) error {
    cmd := exec.CommandContext(ctx, "athena-agent",
        "--id", agent.ID,
        "--role", agent.Role,
        "--project", agent.ProjectID,
    )
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()

    proc := &AgentProcess{ID: agent.ID, Cmd: cmd, Stdin: stdin, Stdout: stdout}

    // 监听子进程 stdout，解析 Agent 的工具调用和黑板写入
    go s.listenAgent(ctx, proc)

    // 监听子进程退出，自动重启
    go s.watchAgent(ctx, proc)
    return nil
}
```

```go
// === athena-agent 子进程内部（使用 Eino） ===

func main() {
    agentID := flag.String("id", "", "Agent ID")
    role := flag.String("role", "", "Agent role")
    projectID := flag.String("project", "", "Project ID")
    flag.Parse()

    ctx := context.Background()

    // 1. 使用 Eino ChatModel 创建 LLM 客户端
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL: config.LLMBaseURL,
        APIKey:  config.LLMAPIKey,
        Model:   config.LLMModel,
    })

    // 2. 注册 Athena 自定义工具（通过 Eino Tool 接口桥接）
    tools := []tool.BaseTool{
        NewBlackboardReadTool(*agentID, *projectID),
        NewBlackboardWriteTool(*agentID, *projectID),
        NewTermTool(*agentID),           // 命令行执行 + LLM 安全审查
        NewMeetingTool(*agentID, *projectID),
        NewMemoryTool(*agentID),          // 个人记忆读写
        NewSkillTool(*agentID),           // 技能管理
        // MCP 工具通过 Tool 接口动态注册...
    }

    // 3. 构建 System Prompt（8层组装，见「Hermes Prompt 组装借鉴」）
    systemPrompt := BuildAgentPrompt(*agentID, *role, *projectID)

    // 4. 使用 Eino ChatModelAgent 创建 ReAct 内循环
    agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Model: chatModel,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: tools,
            },
        },
    })

    // 5. 注册 Eino Callback 切面（审计日志、成本追踪、性能指标）
    callbackHandlers := []handler.Handler{
        NewAuditLogCallback(*agentID),     // OnStart/OnEnd → 审计日志
        NewCostTrackerCallback(*agentID),  // OnEnd → Token 用量统计
        NewPerformanceCallback(*agentID),  // OnStart/OnEnd → 响应时间
    }

    // 6. 创建 Runner 并通过 stdin/stdout 与主进程通信
    runner := adk.NewRunner(ctx, adk.RunnerConfig{
        Agent:    agent,
        Handlers: callbackHandlers,
    })

    // 7. 主循环：读 stdin → Runner.Query → 写 stdout
    RunAgentWithIO(ctx, runner, chatModel, systemPrompt, os.Stdin, os.Stdout)
}

// RunAgentWithIO: 从主进程接收任务，调用 Eino Agent，返回结果
func RunAgentWithIO(ctx context.Context, runner *adk.Runner, chatModel ChatModel,
    systemPrompt string, in io.Reader, out io.Writer) {
    decoder := json.NewDecoder(in)
    encoder := json.NewEncoder(out)

    for {
        var msg AgentMessage
        if err := decoder.Decode(&msg); err != nil {
            break // stdin 关闭，子进程退出
        }

        switch msg.Type {
        case "task":
            // 执行 ReAct 循环
            iter := runner.Query(ctx, msg.Content)
            var result strings.Builder
            for {
                event, ok := iter.Next()
                if !ok { break }
                result.WriteString(event.Message.Content)
            }
            // 返回结果给主进程
            encoder.Encode(AgentResponse{
                Type:    "task_result",
                TaskID:  msg.TaskID,
                Content: result.String(),
            })

        case "meeting_invite":
            // 处理会议邀请
            HandleMeetingInvite(ctx, msg.MeetingID, msg.Agenda)

        case "steer":
            // CEO 新需求（steer 模式）
            iter := runner.Query(ctx, msg.Content)
            // ...处理新需求...
        }
    }
}
```

**子进程模型 vs goroutine 模型对比**：

| 维度 | Goroutine 模型 | 子进程模型 (Athena 选用) |
|------|---------------|------------------------|
| 隔离性 | 共享地址空间，一个 panic 可影响全局 | 独立进程空间，崩溃互不影响 |
| 资源控制 | 无法限制单个 goroutine 的 CPU/内存 | 可通过 cgroup/Job Object 限制资源 |
| 崩溃恢复 | `recover()` 只捕获 panic | 子进程退出 → supervisor 自动重启 |
| 通信开销 | 极低（共享内存） | 较低（stdin/stdout 或 Unix socket） |
| 调试 | 困难（所有 Agent 混在一个进程） | 清晰（每个 Agent 独立进程，独立日志） |

**注意**：管理层 Agent（AgentServer、HR、PM）可使用 goroutine 模型运行（它们数量少、逻辑简单、不需要强隔离），执行层 Agent（开发、测试等）使用子进程模型。

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
规则 10: 任何 Agent 可向 HR 提出招聘请求，HR 评估后直接招聘（无需招聘会议）
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
│  ├ 项目A │  │ 项目看板                              ││
│  └ 项目B │  │ ┌────────┐ ┌────────┐ ┌────────┐      ││
│          │  │ │ 目标    │ │ 进行中  │ │ 完成    │      ││
│ 👥 团队  │  │ │ ○ 任务1 │ │ ○ 任务3 │ │ ✅ 任务5 │     ││
│  ├ 开发   │  │ │ ○ 任务2 │ │ ○ 任务4 │ │ ✅ 任务6 │     ││
│  ├ 测试   │  │ └────────┘ └────────┘ └────────┘      ││
│  └ 设计   │  └──────────────────────────────────────┘│
│          │                                          │
│ 📊 黑板  │  ┌──────────────────────────────────────┐│
│  ├ 事实   │  │ Agent 状态                             ││
│  ├ 猜测   │  │ 🟢 a1b2c3-dev-1   开发中  [任务3]     ││
│  ├ 决议   │  │ 🟢 a1b2c3-test-1  测试中  [任务5]     ││
│  └ 进展   │  │ 🟡 a1b2c3-review-1 审查中  [会议2]    ││
│          │  │ ⚪ a1b2c3-design-1 空闲                ││
│ 🤝 会议  │  └──────────────────────────────────────┘│
│  ├ 进行中 │                                          │
│  └ 决议   │                                          │
│          │  ┌──────────────────────────────────────┐│
│ ⚙ 设置   │  │ [在这里输入你的需求，Athena 会处理...] ││
│          │  │                              [发送]   ││
└──────────┴──────────────────────────────────────────┘
```

### 核心 API

```go
// CEO输入（AgentServer 自动识别意图和项目）
POST /api/chat                              # 全局入口（AgentServer LLM匹配项目+意图）
POST /api/projects/{id}/chat               # 指定项目入口
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

// 裁员管理
GET  /api/layoff/plan                 # HR生成的裁员方案
POST /api/layoff/execute              # CEO确认执行裁员

// MCP工具管理
GET  /api/mcp/available               # 可用MCP工具列表
POST /api/mcp/install                 # 安装新MCP工具

// 实时通知
WS   /ws/projects/{id}/events
```

---

## 项目目录结构

这个目录只是参考, 实际项目不必与这个目录完全一致
```
athena/                              # 项目根目录 (D:/work/athena/)
├── README.md                        # 项目说明
├── PLAN.md                          # 本文件 — 项目规划文档
├── LICENSE
├── .gitignore
├── go.mod                           # Go 模块定义
├── go.sum                           # Go 依赖校验
├── main.go                          # 程序入口
│
├── cmd/                             # CLI 命令
│   └── athena/
│       └── main.go                  # CLI 入口 (start/stop/status)
│
├── config/                          # 配置文件
│   ├── athena.yaml                  # 主配置 (端口/LLM/上限等)
│   └── athena.example.yaml          # 配置示例
│
├── internal/                        # 内部包 (不可外部引用)
│   ├── server/                      # HTTP 服务
│   │   ├── server.go                # Gin 服务入口
│   │   ├── router.go                # 路由注册
│   │   └── middleware.go            # 中间件 (CORS/日志/鉴权)
│   │
│   ├── core/                        # 核心模块
│   │   ├── agent_server.go          # AgentServer (CEO秘书) — 含LLM意图识别与项目UUID匹配
│   │   ├── hr_agent.go              # HR Agent (招聘/裁员/工具分配)
│   │   ├── pm_agent.go              # 项目经理 Agent (需求拆解/任务分配)
│   │   ├── agent_runtime.go         # Agent 运行时 (子进程管理 + supervisor)
│   │   ├── agent_loop.go            # Agent 主循环 (Eino ChatModelAgent 集成 + stdin/stdout 通信)
│   │   ├── llm_client.go            # LLM 调用封装 (基于 Eino ChatModel，扩展 GLM 等)
│   │   ├── eino_tools.go            # Eino Tool 接口桥接 (Athena 工具 → Eino BaseTool)
│   │   ├── context_builder.go       # 上下文构建 (黑板+角色+个人记忆 → Eino System Prompt)
│   │
│   ├── blackboard/                  # 黑板系统
│   │   ├── board.go                 # 黑板核心逻辑 (读写/层级管理)
│   │   ├── fact_manager.go          # 事实管理 (确定/猜测/辅助知识)
│   │   ├── access_control.go        # 读写权限控制 (角色→层级矩阵)
│   │   └── search.go                # 全文搜索 (SQLite FTS5)
│   │
│   ├── meeting/                     # 会议系统
│   │   ├── manager.go               # 会议管理 (创建/邀请/关闭)
│   │   ├── database.go              # 每会议独立 SQLite 管理
│   │   └── resolution.go            # 决议生成与分发
│   │
│   ├── hr/                          # HR 子系统
│   │   ├── recruiter.go             # 招聘逻辑 (模板匹配/上下文注入)
│   │   ├── layoff.go                # 裁员逻辑 (方案生成/Agent销毁)
│   │   └── capacity.go              # 规模上限检查 (人数/资源)
│   │
│   ├── mcp/                         # MCP 工具集成
│   │   ├── registry.go              # MCP Server 注册表
│   │   ├── manager.go               # MCP 生命周期 (下载/配置/启动/停止)
│   │   ├── stdio_transport.go       # stdio 传输 (优先，不占端口)
│   │   └── sse_transport.go         # SSE 传输 (备选，占用端口)
│   │
│   ├── db/                          # 数据库层
│   │   ├── database.go              # SQLite 连接管理
│   │   ├── models.go                # 数据模型 (Agent/Project/Fact等)
│   │   └── migrations/              # 数据库迁移脚本
│   │       ├── 001_init.sql
│   │       └── 002_add_auxiliary.sql
│   │
│   ├── templates/                   # Agent 角色模板(可以被新增)
│   │   ├── ceo_secretary.yaml       # CEO秘书模板 (AgentServer)
│   │   ├── hr.yaml                  # HR 模板
│   │   ├── pm.yaml                  # 项目经理模板
│   │   ├── developer.yaml           # 开发模板
│   │   ├── tester.yaml              # 测试模板
│   │   ├── designer.yaml            # 设计模板
│   │   ├── reviewer.yaml            # 审查模板
│   │   ├── ops.yaml                 # 运维模板
│   │   └── doc.yaml                 # 文档模板
│   │
│   ├── tools/                       # 内置工具定义(可以被新增)
│   │   ├── base.go                  # 工具接口 (Tool interface)
│   │   ├── term.go                  # 命令行执行工具 (LLM 安全审查 + 执行)
│   │   ├── file_tools.go            # 文件读写编辑
│   │   ├── code_tools.go            # 代码搜索/执行
│   │   ├── test_tools.go            # 测试框架/断言
│   │   ├── git_tools.go             # Git 操作
│   │   ├── design_tools.go          # 架构/接口设计
│   │   └── review_tools.go          # 代码审查/Linter/Diff
│   │
│   ├── api/                         # API 路由处理
│   │   ├── projects.go              # 项目 CRUD
│   │   ├── agents.go                # Agent 管理/招聘/裁员
│   │   ├── blackboard.go            # 黑板读写
│   │   ├── chat.go                  # CEO 对话
│   │   ├── meetings.go              # 会议操作
│   │   ├── decisions.go             # CEO 抉择
│   │   ├── mcp.go                   # MCP 工具管理
│   │   └── websocket.go             # WebSocket 实时推送
│   │
│   └── prompts/                     # Prompt 模板 (soul.md 等效)
│       ├── ceo_secretary.md         # CEO秘书角色指令
│       ├── hr.md                    # HR 角色指令
│       ├── pm.md                    # 项目经理角色指令
│       ├── developer.md             # 开发角色指令
│       ├── tester.md                # 测试角色指令
│       ├── designer.md              # 设计角色指令
│       ├── reviewer.md              # 审查角色指令
│       ├── ops.md                   # 运维角色指令
│       └── doc.md                   # 文档角色指令
│
├── data/                            # 运行时数据 (gitignore)
│   ├── board/                       # 黑板数据库 (每项目独立 SQLite)
│   │   ├── board_a1b2c3.sqlite      # 项目 a1b2c3 的黑板
│   │   └── board_d4e5f6.sqlite      # 项目 d4e5f6 的黑板
│   ├── meetings/                    # 会议数据库 (每会议独立文件)
│   │   ├── meeting_m1.sqlite        # 会议 m1 数据
│   │   ├── meeting_m2.sqlite        # 会议 m2 数据
│   │   └── archived/                # 已关闭会议归档
│   │       └── meeting_m0.sqlite
│   ├── agents/                      # Agent 个人数据
│   │   ├── a1b2c3-dev-1/            # 开发 Agent: 项目a1b2c3第1个开发
│   │   │   ├── context.db           # 个人上下文 SQLite
│   │   │   ├── memory.md            # 个人工作记忆 (§分隔, 声明式事实)
│   │   │   ├── skills/              # 个人技能库 (从经验中形成)
│   │   │   │   ├── go-gin-setup/
│   │   │   │   │   └── SKILL.md
│   │   │   │   └── debug-timeout/
│   │   │   │       └── SKILL.md
│   │   │   └── working/             # 工作目录 (代码等)
│   │   └── a1b2c3-test-1/           # 测试 Agent: 项目a1b2c3第1个测试
│   │       ├── context.db
│   │       ├── memory.md
│   │       ├── skills/
│   │       └── working/
│   └── logs/                        # 运行日志
│       └── athena.log
│
├── mcp_servers/                     # 第三方 MCP Server (gitignore)
│   ├── mcp-filesystem/              # 文件系统 MCP
│   └── mcp-sqlite/                  # SQLite MCP
│
├── frontend/                        # Vue 3 前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── index.html
│   └── src/
│       ├── App.vue
│       ├── main.ts
│       ├── api/                     # API 调用封装
│       │   ├── projects.ts
│       │   ├── agents.ts
│       │   ├── blackboard.ts
│       │   ├── chat.ts
│       │   └── websocket.ts
│       ├── views/                   # 页面
│       │   ├── Dashboard.vue        # 仪表盘
│       │   ├── Project.vue          # 项目详情
│       │   ├── Blackboard.vue       # 黑板查看
│       │   ├── Meetings.vue         # 会议列表
│       │   └── Layoff.vue           # 裁员方案
│       ├── components/              # 组件
│       │   ├── ChatInput.vue        # CEO 对话输入框
│       │   ├── ProjectBoard.vue     # 项目看板
│       │   ├── AgentStatus.vue      # Agent 状态卡片
│       │   ├── FactList.vue         # 事实列表 (确定/猜测/辅助)
│       │   ├── MeetingList.vue      # 会议列表
│       │   └── DecisionPanel.vue    # CEO 抉择面板
│       └── stores/                  # Pinia 状态管理
│           ├── project.ts
│           └── agent.ts
│
├── tests/                           # 测试
│   ├── blackboard_test.go
│   ├── agent_runtime_test.go
│   ├── hr_agent_test.go
│   ├── meeting_test.go
│   ├── mcp_test.go
│   └── layoff_test.go
│
├── scripts/                         # 脚本
│   ├── setup.sh                     # 初始化
│   ├── dev.sh                       # 开发模式启动
│   └── build.sh                     # 构建
│
└── docs/                            # 文档
    ├── architecture.md              # 架构说明
    └── api.md                       # API 文档
```

### 路径用途说明

| 路径 | 用途 | 是否 gitignore |
|------|------|---------------|
| `data/board/board_*.sqlite` | 每项目独立黑板数据库 | ✅ 忽略 |
| `data/meetings/meeting_*.sqlite` | 每会议独立数据库 | ✅ 忽略 |
| `data/agents/*/context.db` | Agent 个人上下文 | ✅ 忽略 |
| `data/agents/*/memory.md` | Agent 个人工作记忆 | ✅ 忽略 |
| `data/agents/*/working/` | Agent 工作目录 | ✅ 忽略 |
| `data/logs/` | 运行日志 | ✅ 忽略 |
| `mcp_servers/` | 第三方 MCP Server | ✅ 忽略 |
| `internal/templates/` | Agent 角色模板 | ❌ 纳入版本控制 |
| `internal/prompts/` | 角色 Prompt | ❌ 纳入版本控制 |
| `config/athena.yaml` | 运行时配置 | ✅ 忽略 (example 不忽略) |

---

## 性能与成本基线

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 同时运行 Agent 数 | 10 人员工 | 10 个执行层 Agent 同时工作 |
| Agent 响应时间 | < 30s | 从任务分配到 Agent 开始执行 |
| 黑板读写延迟 | < 500ms | 单次黑板操作响应时间 |

> 上述为初始基线，后续根据实际运行数据调整。

---

## 开发路线图

> 💡 开发顺序遵循"最小闭环优先"：先做出"1 个项目经理 + 1 个开发 + 黑板"的最小可运行闭环，再扩展其他角色。

### Phase 1: 最小可运行 Demo

- [ ] Go 项目脚手架搭建 (Gin + SQLite + Vue 3 + **Eino 引入**)
- [ ] 数据库 Schema 实现与迁移脚本
- [ ] Agent 运行时核心 (supervisor + **Eino ChatModelAgent 子进程**)
- [ ] **Eino ChatModel 集成** (统一 LLM 调用，替换自研 LLM Client)
- [ ] **Eino Tool 接口桥接** (Athena 内置工具通过 Eino Tool 接口注册)
- [ ] 黑板系统实现 (board.go + channel 批量写入 + FTS5 搜索)
- [ ] 基础 API (项目 CRUD + WebSocket)
- [ ] 最简前端 (输入框 + 项目列表)
- [ ] **目标**：1 个 Agent + Eino ReAct 循环 + 1 个黑板读写 + 最简前端，验证 Eino + 自研分层架构

### Phase 2: 核心闭环 (PM + 开发 + 黑板)

- [ ] 项目经理 Agent 实现 (需求拆解 + 不确定性处理 + 验收交付 + Steer模式 + **需求回溯**)
- [ ] Developer Agent + 基础工具集 (含 term tool)
- [ ] AgentServer (CEO秘书) 实现（含 LLM 意图识别 + 项目UUID匹配）
- [ ] HR Agent 实现 (模板化 Agent 创建 + 裁员方案 + 模板扩展)
- [ ] 事实分级系统 (确定/猜测/待验证 + 交叉验证 + 置信度衰减)
- [ ] Agent 自省评分机制 (0-10 分 + 10 分推理流程写入黑板)
- [ ] 上下文注入与隔离机制 (一岗一项目)
- [ ] Agent 个人记忆系统 (memory.md, §分隔, 声明式事实)
- [ ] Agent 个人技能系统 (skills/, 从经验创建/使用中改进)
- [ ] 崩溃恢复与状态持久化 (**Eino Checkpoint** + 自研状态持久化，借鉴 Hermes session 保存机制)

### Phase 3: 会议系统与审查

- [ ] 会议系统实现 (5 步流程: 发起→讨论→决议→写入黑板→关闭)
- [ ] 会议讨论归档压缩 (保留决策依据，删除冗余发言)
- [ ] Reviewer Agent + 代码审查流程 (上下文隔离)
- [ ] 不确定问题上报链路 (Agent → 相关人 → PM → AgentServer → CEO)
- [ ] CEO抉择 API 与界面
- [ ] 裁员方案: Agent 列表 + 闲置时长，CEO 勾选删除

### Phase 4: 全角色 Agent + MCP工具 + 打磨

- [ ] Tester Agent + 测试工具集
- [ ] Designer Agent + 设计工具集
- [ ] Ops Agent + 部署工具集
- [ ] Doc Agent + 文档工具集
- [ ] MCP 工具集成 (registry + manager + stdio传输, Go SDK: https://modelcontextprotocol.io/docs/sdk)
- [ ] 完整 Web 管理界面
- [ ] Agent 状态监控 + 会议可视化
- [ ] 成本预警 (累计费用达 80% 通知 CEO，100% 暂停)
- [ ] 审计日志

### Phase 5: 高级特性 (后续)

- [ ] Agent 自我进化 (参考 Hermes Skills)
- [ ] DAG 任务编排
- [ ] 多项目并行支持 (一岗一项目，多项目多Agent)
- [ ] MCP 工具自动发现与互联网搜索安装
- [ ] 内部工具开发小组 (工具缺失时自研)
- [ ] 插件系统
- [ ] ABANDON 机制 (防止 Agent 陷入死循环)
- [ ] HR 互联网搜索辅助招聘
- [ ] 🔲 多模型策略细化 (不同角色用不同模型，降低成本)
- [ ] 🔲 Docker 打包 (沙箱化部署，安全隔离)

---

## 关键设计决策

### Q1: 为什么不用 CrewAI / MetaGPT / LangChainGo 作为底层框架？

**A**: 它们不满足核心需求——

**Python 框架**：
- CrewAI: Agent 共享上下文，无持久化，无动态创建
- MetaGPT: 严格 SOP 流水线，Agent 上下文不隔离
- 我们需要: 上下文隔离 + 动态招聘 + 黑板持久化 + 事实分级 + 会议沟通

**Go 框架**：
- LangChainGo: Python LangChain 的照搬移植，Go 里写起来不够地道；Agent 功能基础，无中断恢复；v0.1.x API 不稳定
- Google ADK-Go: Google Cloud 厂商锁定风险（Gemini 深度绑定）；A2A 协议有价值但 Athena 不需要跨部署 Agent 协作；社区生态还在建设
- **Eino ✅ 选用**: Go 原生设计，ReAct 内置，Checkpoint 中断恢复，DeepAgent 子Agent委派，无厂商锁定，中文社区强。详见「Go AI 框架选型」

**关键认知**：Eino 是 Athena 的**基础能力层**（LLM 调用 + Agent 内循环 + 工具接口），不是 Athena 的**替代品**。Athena 的核心创新（黑板、会议、HR、公司模型、事实分级）全部在 Eino 之上自研。

### Q2: 为什么选 SQLite 而不是 PostgreSQL？

**A**: 开箱即用原则。SQLite 零配置，单文件，用户(CEO)无需安装数据库服务。后续可通过替换 database.go 支持 PostgreSQL。

### Q3: 为什么后端用 Go 而不是 Python？

**A**: Go 的优势——
- 每个 Agent 独立子进程，天然进程隔离，崩溃互不影响
- 原生并发，supervisor 管理多个子进程开销低
- 单二进制部署，用户(CEO)无需安装 Python 环境
- 类型安全，编译期检查
- 性能更好，适合同时运行多个 Agent

### Q4: 如何防止 Agent 写入错误的"确定性事实"？

**A**: 三重保护——
1. Prompt 层: 反复强调只有 100% 确定的事实才能标记为"确定"
2. 审查层: 专职 Agent 审核黑板写入的事实
3. 升级层: 所有与该项目相关 Agent 可对"确定"事实提出质疑，并提出开会, 会议结论可以把"确定"降级为"猜测"

### Q5: HR Agent 如何决定招聘什么角色？

**A**: 专人专职 + HR 直接执行——
1. 任何 Agent 发现缺少某种能力 → 向 HR 提出招聘请求（附详细理由和岗位需求）
2. HR 评估请求 → 检查公司规模上限（人数/资源），超限上报 AgentServer 向CEO申请
3. HR 直接执行招聘：从模板库选择角色模板，自动配置工具和上下文
4. 模板不足时，HR 成立小组编写新模板
5. 可借助互联网搜索职位需求辅助定义
6. 新 Agent 上线后自动注册到项目和黑板，项目经理分配任务

### Q6: Agent 间沟通与上下文隔离如何平衡？

**A**: 双通道 + 辅助知识共享——
- **黑板通道**: 持久化知识（事实、进展、决议、辅助知识），所有 Agent 按权限访问
- **会议通道**: 实时沟通（讨论、答疑），过程不持久化，只保留决议
- **辅助知识共享**: 报错日志、关键判断依据等虽属"过程"，但对他人判断有直接帮助，通过黑板辅助层共享
- 个人工作记忆永远对其他 Agent 不可见

### Q7: Agent 缺少工具怎么办？

**A**: 三级解决——
1. **优先互联网搜索**：找到成熟工具 → 下载配置 → MCP 注册给对应 Agent（stdio优先，不占端口）
2. **公司内部开发**：搜索不到合适工具 → HR组织开发小组自行开发
3. **上报CEO**：工具卡点无法解决 → 告知CEO问题所在

### Q8: 公司资源不够时如何裁员？

**A**: HR出方案，CEO点选项——
1. HR 扫描已完成项目的 Agent、闲置超时 Agent → 生成裁员方案
2. CEO 在前端界面选择：裁员 / 搁置项目 / 增加上限
3. 执行裁员：supervisor 终止 Agent 子进程、清理上下文、从数据库注销

## 参考项目

| 项目 | 借鉴点 | 参考级别 |
|------|--------|---------|
| **MetaGPT** (FoundationAgents) | 公司角色模型、SOP 工作流、角色间文档传递 | ⭐ **主要架构参考** |
| **Hermes Agent** (NousResearch) | 记忆/技能自省系统、delegation 任务拆解、交互模式 | ⭐ **主要交互参考** |
| **Eino** (CloudWeGo/字节跳动) | Agent 内循环(ReAct)、LLM 统一接口(ChatModel)、Checkpoint 中断恢复、Tool 接口、事件流、回调切面 | ⭐ **基础框架（直接使用）** |
| **CrewAI** | Agent 角色定义 (role/goal/backstory)、YAML 配置解耦 | 辅助参考 |
| **Agent-Blackboard** (claudioed) | 黑板模式实现、MCP 持久化、语义搜索、本体约束 | 辅助参考 |
| **TCH Bytex** (Bytex) | 黑板 + DAG 架构、Agent 协调、上下文隔离 | 辅助参考 |
| **CHYing-agent** (yhy0) | 工具可见性隔离 (visibility)、ABANDON 死循环检测 | 辅助参考 |

---

## 借鉴设计详解

### 一、MetaGPT 架构借鉴（主要架构参考）

MetaGPT 的核心哲学是 **`Code = SOP(Team)`** — 将标准化流程注入团队协作。Athena 借鉴其：

| MetaGPT 模式 | Athena 对应 |
|-------------|-------------|
| SOP 流水线 (PM→Architect→PM→Engineer) | 公司工作流 (AgentServer→PM→各专业Agent→Review) |
| 角色间文档传递 (ProductManager 输出 → Architect 输入) | 黑板层级传递 (PM 拆解的需求 → 开发读取) |
| ProjectRepo 结构化输出 | 黑板数据库 + Agent 个人工作目录 |
| 角色定义 (role/goal/backstory) | Agent 模板 (YAML: role/system_prompt/tools/blackboard) |
| 发布-订阅消息模型 | 黑板读写 + 会议系统 (双通道) |

**Athena 的差异**：
- MetaGPT 是严格流水线，Athena 是按需沟通（只有遇到问题才开会）
- MetaGPT 角色固定，Athena 有 HR 动态招聘
- MetaGPT 无上下文隔离，Athena 严格隔离
- MetaGPT 无事实分级，Athena 有确定/猜测/辅助知识

### 二、Hermes 记忆系统借鉴（主要交互参考）

每个员工 Agent 都要像 Hermes 那样自我反省工作流程，形成记忆和技能。

#### 2.1 Agent 个人记忆 (memory.md)

借鉴 Hermes 的持久化记忆系统，每个 Agent 在自己的目录下维护 `memory.md`：

**存储位置**: `data/agents/{agent-name}/memory.md`

**格式** (借鉴 Hermes 的 `§` 分隔)：
```
项目A使用Go+Gin框架，数据库SQLite，REST风格API§
报错日志路径: /var/log/athena/project-a/，需要root权限读取§
CEO偏好简洁回复，不要冗长描述§
项目A的测试框架用的是testify，运行命令: go test ./... -v
```

**写入规则** (借鉴 Hermes)：
- 写**声明式事实**，不写指令 ✗ "总是用Go写代码" → ✓ "项目A使用Go语言"
- 优先保存**减少未来重复指导**的信息：用户偏好、环境细节、工具陷阱、项目约定
- **不保存**：任务进度（放黑板）、临时调试上下文、容易重新发现的事实
- 容量管理：超过上限时合并压缩旧条目
- 安全扫描：记忆注入系统 prompt，需防注入攻击

**触发时机**：
- Agent 发现环境事实（"项目用Go 1.22"）
- Agent 犯错后获得纠正（"不要用sudo运行Docker"）
- Agent 发现工具陷阱（"ChromaDB需要先create collection再add"）
- CEO 给出偏好（"回复简洁点"）

#### 2.2 Agent 个人技能 (skills/)

借鉴 Hermes 的 skill_manage 系统，Agent 从工作经验中形成可复用的技能：

**存储位置**: `data/agents/{agent-name}/skills/`

**技能结构** (借鉴 Hermes SKILL.md)：
```
data/agents/a1b2c3-dev-1/
├── memory.md              # 个人事实记忆
├── context.db             # 个人上下文 SQLite
├── working/               # 工作目录
└── skills/                # 个人技能库
    ├── go-gin-setup/
    │   └── SKILL.md       # Go Gin 项目搭建技能
    ├── sqlite-migration/
    │   └── SKILL.md       # SQLite 迁移技能
    └── debug-timeout/
        └── SKILL.md       # 超时问题调试技能
```

**技能创建触发** (借鉴 Hermes)：
1. 完成复杂任务后（5+ 工具调用）
2. 犯错后找到正确做法时
3. CEO 纠正方法后
4. 发现非平凡工作流时

**技能格式**：
```yaml
---
name: go-gin-setup
description: Go Gin 项目脚手架搭建，含路由/中间件/错误处理模板
tags: [go, gin, scaffold]
created_at: 2026-05-02
use_count: 0
---
# Go Gin 项目搭建

## 触发条件
新建 Go Web 项目时使用

## 步骤
1. go mod init
2. 引入 gin + sqlite driver
3. 创建 internal/server/ 结构
4. 配置中间件 (CORS/日志)
...

## 常见坑
- gin.Context 必须在 handler 内使用，不能传到 goroutine
- SQLite 需要启用 WAL 模式并发读写
```

#### 2.3 技能自省与维护 (Curator 机制)

借鉴 Hermes Curator，每个 Agent 定期自省和维护自己的技能库：

| 机制 | 说明 |
|------|------|
| **技能生命周期** | active → stale(30天未用) → archived(90天未用) → 可恢复 |
| **技能合并** | 发现近似重复技能时合并为一个 |
| **技能补丁** | 使用中发现技能过时/有误时立即修补 (patch) |
| **使用统计** | 记录每个技能的 use_count / view_count / last_used_at |
| **手动保护** | 重要技能可标记 pinned，不被自动归档 |

### 三、Hermes 任务拆解借鉴 → PM Agent 实现

Hermes 的 `delegate_task` 机制效果很好，Athena 的 PM Agent 借鉴其核心设计：

#### 3.1 Hermes delegate_task 工具详解

**工具参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `goal` | string | 必填 | 子Agent的任务目标 |
| `context` | string | 必填 | 子Agent所需的全部背景信息（**子Agent零历史，这是唯一上下文来源**） |
| `toolsets` | string[] | `["terminal","file","web"]` | 子Agent可用的工具集 |
| `role` | string | `"leaf"` | `"leaf"`=不能再委派；`"orchestrator"`=可生成自己的子Agent |
| `max_iterations` | int | 50 | 子Agent最大工具调用轮次 |
| `tasks` | array | — | 并行批量任务数组（最多 `max_concurrent_children` 个） |

**Hermes 配置项** (`config.yaml → delegation`)：

```yaml
delegation:
  model: ''                    # 子Agent可使用不同模型（空=继承父Agent）
  provider: ''
  inherit_mcp_toolsets: true   # 子Agent继承父Agent的MCP工具集
  max_iterations: 50           # 每个子Agent最大工具调用轮次
  child_timeout_seconds: 600   # 子Agent超时（10分钟，每次API/工具调用重置）
  max_concurrent_children: 3   # 最大并发子Agent数
  max_spawn_depth: 1           # 嵌套深度限制（1=扁平，2=允许子Agent再生成leaf，3=三层嵌套）
  orchestrator_enabled: true   # 全局开关：关闭则所有子Agent强制为leaf
  subagent_auto_approve: false # 子Agent工具调用是否自动批准
```

**子Agent权限限制**：

| 工具 | Leaf | Orchestrator | 原因 |
|------|:----:|:----------:|------|
| `delegate_task` | ❌ | ✅ | 防止递归委派失控 |
| `clarify` | ❌ | ❌ | 子Agent不能与用户交互 |
| `memory` | ❌ | ❌ | 子Agent不能写共享持久记忆 |
| `send_message` | ❌ | ❌ | 子Agent不能跨平台发消息 |
| `execute_code` | ❌ | ❌ | 子Agent应逐步推理 |

**子Agent生命周期**：

```
1. 父Agent调用 delegate_task(goal=..., context=..., toolsets=...)
2. 系统生成子Agent实例
   - 全新隔离对话（零历史）
   - 限制工具集
   - 独立终端会话
   - System Prompt 由 goal + context 构建
3. 子Agent执行任务
   - 受 max_iterations 限制（默认50轮）
   - 受 child_timeout_seconds 限制（默认600秒，API调用重置计时器）
4. 子Agent完成 → 返回结构化摘要（做了什么、发现了什么、修改了哪些文件、遇到什么问题）
5. 父Agent接收摘要，继续工作
```

**关键设计**：
- **同步阻塞**：`delegate_task` 在父Agent当前轮次内执行，阻塞直到所有子Agent完成
- **批量并行**：`tasks=[...]` 使用 ThreadPoolExecutor 并发执行，结果按输入顺序返回
- **中断传播**：父Agent被中断 → 所有子Agent被取消 → 返回 `status="interrupted"`
- **兄弟协调**：多个子Agent修改同一文件时，系统会发出 sibling modification 警告

#### 3.2 Hermes Subagent-Driven Development 技能

Hermes 的 `subagent-driven-development` 技能定义了一个**两阶段审查**的任务执行流程，Athena 的 PM Agent 直接借鉴：

**流程**：

```
1. 读取计划 → 提取所有任务（一次性读取，不让子Agent自己读计划文件）
2. 对每个任务：
   a. 派遣实现子Agent（delegate_task with goal + 完整context）
   b. 派遣规范审查子Agent → 检查是否符合原始spec
      - 不通过 → 修复 → 重新审查
   c. 派遣质量审查子Agent → 检查代码质量
      - 不通过 → 修复 → 重新审查
   d. 标记任务完成
3. 所有任务完成 → 派遣最终集成审查子Agent
4. 全量测试 + 提交
```

**任务粒度原则**：每个任务 = 2-5 分钟的专注工作。

| 太大 | 合适 |
|------|------|
| "实现用户认证系统" | "创建User模型，包含email和password字段" |
| | "添加密码哈希函数" |
| | "创建登录接口" |
| | "添加JWT token生成" |

**子Agent Context 打包要点**（Hermes 最佳实践）：

```
BAD: delegate_task(goal="修复那个错误")           ← 子Agent不知道"那个"是什么
GOOD: delegate_task(
    goal="修复 api/handlers.py 中的 TypeError",
    context="""
    api/handlers.py 第47行 TypeError:
    'NoneType' object has no attribute 'get'.
    process_request() 从 parse_body() 接收 dict，
    但 parse_body() 在 Content-Type 缺失时返回 None。
    项目路径: /home/user/myproject, Python 3.11。
    """
)                                                ← 子Agent拥有完整上下文
```

#### 3.3 Athena PM Agent 设计（借鉴 Hermes delegate_task）

**PM Agent 收到项目后的工作流**：

```
1. 需求分析
   - 识别需求中的矛盾和歧义
   - 评估可行性和优化空间
   - 输出：需求分析文档 → 写入黑板

2. 任务拆解（借鉴 Hermes 隐式拆解 + TASK.md 协议）
   - 将需求拆解为可独立执行的子任务
   - 每个任务 = 2-5 分钟专注工作（借鉴 subagent-driven-development 粒度）
   - 每个子任务包含：目标、约束、验收标准、依赖关系
   - 识别需要的角色和技能
   - 输出：任务列表 → 写入黑板

3. 人员评估
   - 对比现有 Agent 的角色和技能
   - 缺口角色 → 立刻与 HR 沟通招人
   - 输出：招聘请求 → 发给 HR

4. 任务分配（借鉴 Hermes delegate_task 的 context 打包）
   - 按依赖关系排序任务
   - 独立任务可并行分配（借鉴 max_concurrent_children=3）
   - 每个任务打包完整上下文（不让Agent自己读计划，PM提供全部信息）
   - 输出：agent_tasks 表记录

5. 两阶段审查（借鉴 subagent-driven-development）
   a. 规范审查 → 检查是否符合需求spec
      - 不通过 → Developer 修复 → 重新审查
   b. 质量审查（Reviewer Agent）→ 检查代码质量
      - 不通过 → Developer 修复 → 重新审查

6. 进度跟踪
   - 监控黑板上的进展更新
   - 发现阻塞 → 组织会议协调
   - 发现不确定 → 处理上报链路
   - 任务超时（借鉴 child_timeout_seconds）→ 上报 PM

7. 项目经理验收交付（PM 是最清楚需求的 Agent，负责最终验收）
   - **以最高标准验收**，避免提交给CEO后被返工
   - 验收维度：完善性、低 bug、高可运行、高性能
   - 验收依据：
     - 测试 Agent 提交的测试报告（md 格式 / html 格式）
     - 开发 Agent 提交的代码
     - Review Agent 的审查结论
     - 其他有助于验收的材料(比如类似项目的资料, 论文等)
   - 验收不通过 → 要求对应 Agent 整改 → 整改后重新验收 → **循环验证直至完善**
   - 任何不合适的地方都要求修改、返工、补充——宁可内部多轮验证，不可提交后被CEO返工
   - 验收通过 → 告知 AgentServer → AgentServer 告知CEO
   - CEO认为还需改进 → 新需求通过 AgentServer 下发给 PM → PM 修改需求并分配

8. 需求随时更新（steer 模式，借鉴 Hermes steer）
   - 执行层员工工作时，CEO可随时通过 AgentServer 向 PM 提出新要求
   - 这是**并行**的，不是串行——不会等当前任务全部完成才接收新需求
   - PM 收到新需求后，更新黑板上的项目目标和任务列表，实时通知相关 Agent
```

**PM Agent 的任务打包格式** (借鉴 Hermes delegate_task 的 goal+context 模式)：

```json
{
  "task_id": "task-001",
  "agent_id": "a1b2c3-dev-1",
  "goal": "实现用户认证模块，支持JWT登录",
  "context": {
    "task_from_plan": "创建 src/auth/jwt.go，实现 TokenGenerate 和 TokenValidate 函数",
    "project_goals": "构建Go后端API服务",
    "tech_stack": "Go 1.22 + Gin + SQLite",
    "constraints": ["使用golang-jwt库", "token有效期24h"],
    "dependencies": ["task-000: 数据库Schema已创建"],
    "acceptance_criteria": ["登录接口返回JWT", "中间件校验token"],
    "tdd_steps": [
      "1. 在 tests/auth/jwt_test.go 写失败测试",
      "2. 运行 go test ./tests/auth/ -v 确认失败",
      "3. 实现最小代码",
      "4. 运行 go test ./tests/auth/ -v 确认通过",
      "5. 运行 go test ./... -q 确认无回归"
    ]
  }
}
```

**PM Agent → 员工Agent 的"子Agent"限制** (借鉴 Hermes 权限隔离)：

| 能力 | 员工Agent | 说明 |
|------|:--------:|------|
| 接受新任务 | ✅ | 从 PM 获取 |
| 写黑板 | ✅ | 按角色权限 |
| 参加会议 | ✅ | 遇到问题主动发起 |
| 直接找CEO | ❌ | 通过 PM → AgentServer → CEO |
| 招人 | ❌ | 向 HR 申请，经会议审核 |
| 再委派子任务 | ❌ | 只有 PM 可以分配任务 |
| 写共享记忆 | ✅ | 写入个人 memory.md，不影响其他Agent |

#### 3.4 关键差异：Hermes delegate_task vs Athena PM 分配

| 维度 | Hermes delegate_task | Athena PM Agent |
|------|---------------------|-----------------|
| 拆解方式 | 隐式（Agent自行判断何时委派） | 显式（PM专门负责拆解和分配） |
| 上下文传递 | goal+context 字符串 | 黑板读取 + 任务打包 JSON（含 TDD 步骤） |
| 人员调度 | 固定工具集，无动态扩编 | HR 动态招聘，按需扩编 |
| 审查机制 | 无内置审查（需 skill 配合） | 内置两阶段审查（规范→质量） |
| 任务依赖 | 无显式依赖管理 | 依赖关系排序 + 阻塞检测 |
| 超时处理 | child_timeout_seconds，超时返回失败 | 任务超时 → 上报 PM → 可能开会协调 |
| 中断传播 | 父Agent中断→子Agent全部取消 | CEO可暂停项目→该项目所有Agent暂停 |
| 不确定性处理 | 无 | 上报链路 → PM → AgentServer → CEO |
| 结果收集 | 只返回结构化摘要 | 员工写入黑板 + 任务状态更新 |

### 四、Hermes Prompt 组装借鉴 → Athena Agent System Prompt

Hermes 的 System Prompt 采用 **10 层动态拼装**，Athena 借鉴此架构为每个员工 Agent 构建 system prompt：

#### 4.1 Hermes Prompt 10 层结构

| 层 | 内容 | 来源 | 缓存 |
|----|------|------|:----:|
| 1 | Agent 身份 | SOUL.md 或 DEFAULT_AGENT_IDENTITY | ✅ |
| 2 | 工具使用行为指导 | prompt_builder.py 硬编码 | ✅ |
| 3 | Honcho 静态块 | 第三方集成 | ✅ |
| 4 | 可选系统消息 | 配置/API 覆盖 | ✅ |
| 5 | 冻结 MEMORY 快照 | memory 工具持久化 | ✅ |
| 6 | 冻结 USER 快照 | 用户画像数据 | ✅ |
| 7 | 技能索引 | skills/ 目录扫描 | ✅ |
| 8 | 项目上下文文件 | .hermes.md → AGENTS.md → CLAUDE.md → .cursorrules | ✅ |
| 9 | 时间戳 + Session ID | 运行时生成 | ✅ |
| 10 | 平台提示 | CLI/Discord/Slack 等 | ✅ |

**关键设计**：
- **冻结快照**：MEMORY 和 USER 在会话启动时加载，会话内的写入不修改已构建的 prompt
- **安全扫描**：所有上下文文件经过注入攻击检测，截断上限 20K 字符（70/20 头尾比）
- **子Agent特殊处理**：`skip_context_files=True` → 使用 DEFAULT_AGENT_IDENTITY 替代 SOUL.md，跳过项目上下文

#### 4.2 Athena Agent System Prompt 组装（借鉴 Hermes）

| 层 | 内容 | Athena 来源 | 说明 |
|----|------|------------|------|
| 1 | Agent 身份 | `internal/prompts/{role}.md` | 角色专属 soul.md，替代 Hermes 的 SOUL.md |
| 2 | 公司级指令 | 黑板层级0（项目元信息） | 所有Agent共享 |
| 3 | 项目级指令 | 黑板层级1-2（事实+猜测） | 同项目Agent共享 |
| 4 | 会议决议 | 黑板层级5 | 相关Agent共享 |
| 5 | 冻结个人记忆快照 | `data/agents/{name}/memory.md` | §分隔声明式事实 |
| 6 | 技能索引 | `data/agents/{name}/skills/*/SKILL.md` | 个人技能列表 |
| 7 | 任务上下文 | PM 打包的 goal+context JSON | 当前任务的完整信息 |
| 8 | 时间戳 + AgentID | 运行时生成 | 会话标识 |

**Athena vs Hermes 差异**：
- Hermes 的 USER 快照 → Athena 无此层（CEO偏好通过 AgentServer 传递）
- Hermes 的项目上下文文件 → Athena 用黑板替代（更结构化）
- Hermes 的平台提示 → Athena 统一为 Web API 交互
- Athena 新增"任务上下文"层 → 来自 PM 的 delegate_task 式打包

---

*本文档是 Athena 项目的规划文档，将随开发进展持续更新。*
*最后更新: 2026-05-02 (v11: 引入 Eino 作为 Go AI 基础框架，技术栈从"全自研"改为"Eino基础层+Athena自研层"分层架构)*
