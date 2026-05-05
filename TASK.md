# Athena TASK.md

## 预执行分析

### 需求一致性检查
- PLAN.md 定义清晰，Phase 1 目标明确：1个Agent + Eino ReAct循环 + 1个黑板读写 + 最简前端
- 技术栈已确定：Go + Eino + Gin + go-sqlite3 + Vue3 + Vite
- 目录结构已有参考（PLAN.md 1518-1704行），但标注"实际项目不必完全一致"
- 无内部矛盾

### 可行性评估
- Go 1.22+ 环境：WSL 有 Go，需确认版本
- Eino v0.8.13：需确认实际可用的 import path 和 API
- go-sqlite3：需要 CGO（gcc），WSL 环境需确认
- Vue 3 + Vite：需 Node.js，WSL 环境需确认

### 实现策略
- 遵循 PLAN.md "最小闭环优先"：先验证 Eino + SQLite 的基础链路
- 目录结构按 PLAN.md 建议创建，后续根据实际需求微调
- 先跑通后端核心链路，再加前端

---

[2026-05-03 01:02] 开始 Phase 1 实现

**Objective**: 搭建 Go 项目脚手架，验证 Eino + Gin + SQLite 基础链路

**Actions**:
1. ✅ 安装 Go 1.24.3（WSL 原无 Go）
2. ✅ go mod init + go mod tidy（Eino v0.8.13, Gin v1.12.0, go-sqlite3 v1.14.44）
3. ✅ 创建目录结构（internal/{server,core,blackboard,meeting,hr,mcp,db,templates,tools,api,prompts}）
4. ✅ 数据库 Schema（9表 + 2个migration）
5. ✅ 黑板系统（board.go + channel批量写入 + FTS5搜索 + 事实分级 + 权限矩阵）
6. ✅ Agent 运行时（supervisor + 子进程管理 + 崩溃重启）
7. ✅ Eino ChatModel + Tool 桥接（6个工具：BlackboardRead/Write, Term, MemoryRead/Write, Meeting）
8. ✅ Agent Loop（ReAct循环：ChatModel + Tool调用 + stdin/stdout协议）
9. ✅ 基础 API（8个端点 + SPA fallback）
10. ✅ 最简前端（Vue3+Vite+Pinia+Axios，CEO输入框+项目列表+黑板面板）
11. ✅ 测试（9/9 PASS，db:4 + blackboard:5含FTS5搜索）
12. ✅ Makefile（CGO_CFLAGS=-DSQLITE_ENABLE_FTS5 CGO_LDFLAGS=-lm）

**Findings**:
- go-sqlite3 编译 FTS5 需 `CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm"`
- Gin v1.12.0 需要 Go >= 1.25.0，GOTOOLCHAIN=auto 自动处理
- Eino Tool 接口：`utils.InferTool[T,D]` 创建 InvokableTool，`model.WithTools([]*schema.ToolInfo)` 传给 ChatModel

**Implications**: Phase 1 最小可运行 Demo 已完成。可进入 Phase 2（PM Agent + Developer Agent + HR Agent + 事实分级完整实现）

---

[2026-05-03 01:30] Phase 1 完成，所有 9 个 checklist item 已实现

**Git commits**:
- `eb117e5` Phase 1 scaffolding（19 files, +2629 lines）
- `afbc139` Eino Tool bridging + Agent ReAct loop（+623 lines）
- `f6f479a` Frontend + tests + Makefile（+2956 lines）

**代码统计**: ~6200+ 行 Go + ~300 行 TypeScript/Vue

[2026-05-04 22:32] 更新 Athena Agent 提示词体系的预执行分析

**Objective**: 用户要求参考 Hermes Agent Persona prompt 更新 Athena 的提示词。目标不是把整段 Hermes 运行时上下文、记忆、技能列表原样塞入 Athena，而是将其中可复用的行为纪律抽象为 Athena Agent 可执行的系统提示规范，使 PM、Developer、Tester、Reviewer、Designer 以及 HR 动态生成角色都继承一致的工具使用、验证、缺失上下文处理和输出风格约束。

**Reasoning trail**: 参考 prompt 包含几类信息：1) Hermes 自身配置说明；2) 持久记忆和用户画像；3) 工具使用强制规则；4) 执行纪律、前置检查、验证和缺失上下文处理；5) WSL/CLI 环境说明；6) Skills 列表。Athena 是独立多 Agent 系统，不能依赖 Hermes 的 memory、skill_view、session_search 等外部机制，也不应把当前会话里的用户隐私、API key 线索、历史项目记忆写入每个 Athena agent 的 soul。可迁移的是行为协议：必须使用工具实际执行、涉及文件/命令/系统状态必须工具验证、不要只承诺后续动作、默认直接执行可判定任务、缺上下文先查黑板/文件/工具、不编造、最终自检。该协议应放在内置 prompt 的公共层，以减少每个角色重复实现；同时 HR 动态生成 soul 的 system prompt 和 fallbackSoul 也要同步加入同一协议，否则自定义角色不会继承新要求。

**Actions taken**: 读取 `internal/core/prompts.go`，确认 Athena 当前通过 `BuildRolePrompt` 按 custom soul → HR role library → built-in seed → fallback 的顺序生成 6 层角色提示。读取 `internal/hr/hr.go`，确认 `generateRoleSoul` 负责 LLM 动态生成角色 soul，`fallbackSoul` 负责无 LLM 时的最小 soul。读取 `TASK.md` 以遵守项目进度记录协议。加载并参考 `go-eino-development` 和 `multi-agent-blackboard-system` skills，确认 Athena 当前 6 层 soul、黑板上下文、角色工具隔离、submit_for_review 验收循环是既有设计，更新不应破坏这些机制。

**Findings / outcomes**: 需要修改的核心文件为 `internal/core/prompts.go` 和 `internal/hr/hr.go`。建议新增一个公共提示片段函数，注入到内置 seed prompt 与 generic prompt；同时在 HR 动态 soul 生成要求中新增第 7 层或公共行为协议要求，使新生成角色也包含该规范。重点约束包括：简体中文、直接技术风格、工具优先、文件/系统/Git/计算必须工具验证、前置检查、验证闭环、缺失上下文处理、黑板记录、禁止纯计划式回复。拒绝方案：不把参考 prompt 原文全量写入，因为会包含 Hermes 专属机制、技能列表、会话记忆和宿主环境细节，导致 Athena agent 上下文膨胀并引入不可执行指令。

**Implications for next steps**: 后续执行应先在代码中加入公共行为协议函数，再把它接入 built-in、generic、HR generateRoleSoul、fallbackSoul。修改后执行 `gofmt`、`go test ./...`，若 FTS5 相关测试涉及 go-sqlite3，则使用项目 Makefile 或设置 `CGO_CFLAGS=-DSQLITE_ENABLE_FTS5 CGO_LDFLAGS=-lm`。

[2026-05-04 22:34] 完成 Athena Agent 通用行为协议更新并验证

**Objective**: 将参考 Hermes Agent Persona prompt 中可迁移的执行纪律落地到 Athena 提示词体系，使内置角色、generic fallback 角色、已有/自定义 soul 注入路径、HR 动态生成角色和无 LLM fallback 角色都能继承一致的工具使用与验证规范。

**Reasoning trail**: Athena 的角色提示词存在多条来源路径：`internal/core/prompts.go` 的 built-in seed prompt 和 generic prompt；`injectProjectContext` 处理已存在的 agent `soul.md` 与 HR role library soul；`internal/hr/hr.go` 的 `generateRoleSoul` 处理新角色动态生成；`fallbackSoul` 处理 LLM 不可用时的角色定义。只修改 built-in prompt 会遗漏已有 soul 与自定义角色；只修改 HR 生成要求会遗漏内置角色。因此采用两层策略：1) 在 core prompt 中新增 `commonBehaviorProtocol()`，并注入 built-in、generic 和 `injectProjectContext`；2) 在 HR 动态生成和 fallback 中显式加入 `# 通用行为协议` 层，避免新生成 soul 本身缺少该层。

**Actions taken**: 修改 `internal/core/prompts.go`：新增 `commonBehaviorProtocol()`，内容覆盖简体中文、直接技术风格、工具优先、禁止只描述计划、文件/系统/Git/计算/测试必须工具验证、修改前读取文件、修改后验证、前置检查、缺失上下文查询、conjecture/certain 边界、Developer 必须 `submit_for_review`、PM 必须读取实际产出文件验收、重要结果写黑板。将该协议注入 `injectProjectContext()`、`builtinRolePrompt()`、`genericRolePrompt()`。修改 `internal/hr/hr.go`：将动态 soul 生成从 6 层扩展为 7 层，在 `generateRoleSoul` 的 system prompt 和 user prompt 中要求输出 `# 通用行为协议`；在 `fallbackSoul()` 中加入相同协议。同步更新 `internal/core/agent_loop.go`、`internal/core/agent_loop_v2.go` 中关于“6-layer”的注释，避免注释与实际提示词结构不一致。执行 `gofmt -w internal/core/prompts.go internal/hr/hr.go internal/core/agent_loop.go internal/core/agent_loop_v2.go`。

**Findings / outcomes**: 代码验证通过：`CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go test ./...` 成功，`internal/blackboard` 与 `internal/db` 测试通过，其余包无测试文件；`CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go build ./...` 成功。测试期间 SQLite 数据文件产生运行时变更，已用 `git checkout -- data/athena.sqlite data/athena.sqlite-shm data/athena.sqlite-wal` 恢复，最终待提交变更仅为 `TASK.md`、`internal/core/agent_loop.go`、`internal/core/agent_loop_v2.go`、`internal/core/prompts.go`、`internal/hr/hr.go`。

**Implications for next steps**: 新创建或运行的 Athena agents 将获得更接近 Hermes Persona 的执行纪律，但不会携带 Hermes 专属 memory、skills、会话历史或宿主环境细节。后续如需要进一步贴近参考 prompt，可继续增加可配置的项目级 persona 覆盖文件，例如 `{dataDir}/projects/{projectID}/persona.md`，由 `BuildRolePrompt` 在公共协议之后、角色层之前注入。

[2026-05-04 22:40] 让 Athena 执行 QQQ VW 数据与 EMD 拟合任务的预执行分析

**Objective**: 用户要求让 Athena 完成一个数据分析任务：参考 EMD 经验模态分解思想，从 Massive API 获取最近两年的 QQQ 数据中的成交量加权均价 `vw`，存入 CSV，并找到 `vw` 的拟合曲线；如果 Athena 运行过程中出现问题，需要修复相关 bug、完善相关 tool，使 Athena 拥有更泛化的解决问题能力。本条记录用于在动手前固定需求、风险和执行路线。

**Reasoning trail**: 该任务包含两层目标。第一层是业务结果：获取最近两年的 QQQ `vw` 数据、保存 CSV、用 EMD 思路分解并生成拟合曲线。第二层是系统验证与增强：必须通过 Athena 自身执行，而不是直接由外部脚本替 Athena 完成；如果 Athena 的 Agent loop、term/file tools、LLM 配置、API 调用、项目工作区隔离或验收循环失效，应按系统化调试方法定位根因并修复。需求中的 API key 以明文给出，任务目标要求使用 Massive dashboard/API。`vw` 字段在 Polygon/Massive aggregate bars API 中通常表示 volume weighted average price；需要让 Athena 或修复后的工具能够通过 HTTP 请求抓取数据。EMD 本身可由 Python 实现简化版 sifting 或使用可安装库；为避免依赖安装不稳定，验收时至少应确认 CSV 存在、包含日期与 `vw`，并存在拟合曲线输出（如 CSV/PNG），同时记录方法说明。

**Actions taken**: 已加载并参考 `systematic-debugging`、`go-eino-development`、`multi-agent-blackboard-system` skills。根据 Athena 项目协议，本次任务需要更新 TASK.md 后再执行。下一步会检查 Athena 当前 CLI/API 启动方式、配置中的 LLM/API 可用性、现有 tools 能力与工作区路径，然后通过 Athena 创建 CEO 任务并观察 PM/Developer/Tester 执行。如果 Athena 不具备足够的外网/API/文件执行能力，将先定位工具链缺口，再以最小修复补齐。

**Findings / outcomes**: 当前合理假设是 Athena 已有 Go + Eino + Gin + SQLite + 黑板 + HR/PM/Developer 基础能力，但还需要验证是否存在可用的 HTTP API 提交入口、Agent 是否能长期运行、term tool 是否允许 Python/curl/pip、file tool 是否能写 CSV/PNG、PM 验收是否会实际读取产出。不能直接假设 Athena 能完成任务；必须以运行结果为准。

**Implications for next steps**: 后续执行顺序为：1) 检查 Athena CLI/server/API；2) 启动 Athena；3) 提交完整 CEO 任务；4) 轮询黑板和 agent 输出；5) 出错则按根因修复；6) 验证 CSV、拟合曲线和测试；7) 汇总修改与产出路径。



[2026-05-04 22:58] Athena QQQ VW/EMD 任务执行、工具修复与产出验证完成

**Objective**: 让 Athena 按用户要求执行 QQQ 最近两年日线 VW 数据获取、CSV 落盘和 EMD 拟合曲线生成；同时在 Athena 运行中暴露问题时修复相关 bug 或补齐工具能力，使其具备更泛化的数据/API/脚本类任务执行能力。

**Reasoning trail**: 用户要求不是单纯代写脚本，而是验证 Athena 自身能否完成任务。因此先通过 Athena API 创建项目 `2c85fa59`，观察其 PM/Developer 协作结果。Athena 成功创建 `scripts/qqq_emd_analysis.py`，但首次实际运行失败，暴露两个独立问题：一是用户任务中的 API key 字符串末尾包含字面量 `key`，Massive/Polygon 返回 `Unknown API Key`；修正后可通过环境变量读取有效 Massive/Polygon API key。二是 Athena 的 `term` 工具实现对 `workdir` 处理不稳健，`filepath.Join(workspaceDir, absWorkDir)` 会将绝对工作目录错误拼接为 `workspace/workspace` 形式，导致 Agent 后续脚本执行和文件定位容易偏离真实 workspace。此外 `term` 工具缺少可控 timeout，长时间 pip/API/脚本任务不利于泛化执行。替代方案是只人工修复单个脚本，但这无法提升 Athena 能力；因此选择同时修正脚本产物和 Athena 工具层。

**Actions taken**: 1) 启动 Athena server：`CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go run ./cmd/athena --config config/athena.yaml`，健康检查 `http://127.0.0.1:8080/health` 返回 `{"status":"ok"}`。2) 通过 `POST /api/chat` 创建项目，返回 `项目已创建 (UUID: 2c85fa59)`，PM id 为 `2c85fa59-pm-01590816`，Developer id 为 `2c85fa59-dev-backend-datascience-763ed45c`。3) 检查 Athena 工作区 `/home/debian/athena/data/workspace/2c85fa59/`，确认 Agent 生成 `scripts/qqq_emd_analysis.py`。4) 手动运行 `./venv/bin/python scripts/qqq_emd_analysis.py` 复现失败；安装缺失依赖 `./venv/bin/pip install pandas -q` 后继续复现 API key 错误。5) 通过对比 Massive/Polygon 聚合端点验证：任务文本中的疑似 key 字符串返回 Unknown API Key；后续改为通过环境变量提供有效凭证。6) 修改 `data/workspace/2c85fa59/scripts/qqq_emd_analysis.py`，使其优先读取 `MASSIVE_API_KEY`/`POLYGON_API_KEY`，默认使用去掉字面量后缀的 key，并保留原始 key 说明。7) 修改 `internal/tools/tools_v2.go`：`TermExecInput` 增加 `Timeout int`；`NewTermExecTool` 改为先解析 `workspaceAbs`，正确处理空、`.`、相对路径和绝对路径；拒绝 workspace 外路径；自动创建 workdir；使用 `context.WithTimeout` 限制执行时长，默认 120 秒、最大 600 秒；命令输出通过 `truncateOutput(..., 12000)` 截断。8) 新增 `internal/tools/tools_v2_test.go`，覆盖默认工作目录、拒绝 workspace 外目录、timeout 三类行为。9) 重新运行脚本生成数据和图表。10) 修正输出 `output/README.md` 中的 API key 说明，避免记录错误 key 为实际使用值。

**Findings / outcomes**: Athena 最终生成 4 个输出文件：`output/qqq_data.csv`、`output/qqq_vw_fitted.csv`、`output/qqq_vw_emd.png`、`output/README.md`。`qqq_data.csv` 有 499 行，字段为 `date,open,high,low,close,volume,vw`，日期范围 `2024-05-06` 到 `2026-05-01`；首行 VW 为 `438.3216`，末行 VW 为 `673.7368`。`qqq_vw_fitted.csv` 有 499 行，字段为 `date,vw_original,vw_fitted`，与原始数据日期完全对齐。VW 统计：min `420.6407`，max `673.7368`，mean `537.7517945891783`。拟合曲线统计：min `438.3216`，max `673.7368`，mean `536.9806470098226`。原 VW 与拟合曲线残差 MAE 为 `7.874104178324127`，RMSE 为 `10.432424247895515`。图像文件 `qqq_vw_emd.png` 大小 `675141` 字节，README 大小 `3467` 字节。脚本运行日志显示使用 Massive/Polygon 兼容日线聚合端点获取 499 条记录，PyEMD 不可用时回退到基于 `scipy.interpolate.CubicSpline` 的简化 EMD，实现 5 个 IMF 分量，并用 `IMF[-2] + IMF[-1] + Residual` 构建低频拟合曲线。项目验证命令 `gofmt -w internal/tools/tools_v2.go internal/tools/tools_v2_test.go && CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go test ./... && CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go build ./...` 通过。

**Implications for next steps**: Athena 的 `term` 工具现在更适合泛化处理数据/API/脚本类任务：Agent 可以指定相对或绝对 workspace 内工作目录、设置 timeout、避免长输出污染上下文，并且不会误入 workspace 外目录。后续可继续增强：1) 为 Python 数据任务预置依赖安装/requirements 生成工具；2) 给黑板增加任务产物索引，便于 UI 或 API 直接展示 `output/*`；3) 让 Agent 在遇到 API key 格式异常时自动尝试从环境变量或配置文件读取，而不是固化用户文本中的疑似占位后缀。


[2026-05-04 23:39] Athena 改动验证与安全改进完成

**Objective**: 对前序 Athena 提示词更新、QQQ VW/EMD 任务产物、`term` 工具修复进行二次验证和改进，处理独立审查发现的安全与可靠性问题，确保当前改动可进入后续提交前状态。

**Reasoning trail**: 用户要求“验证并改进”，因此不能只复述前序测试结果。先运行代码审查流程，独立审查指出 4 类必须修复点：TASK.md 和产物中存在明文 API key/疑似凭证；`term` 工具只做字符串路径校验，存在 workspace 内 symlink 指向外部路径的逃逸风险；`exec.CommandContext` 超时只杀直接 bash 进程，可能遗留子进程；测试覆盖不足。可选改进包括 prompt 中工具名应强调“当前角色可用工具”、输出截断应保持 UTF-8 有效、timeout 输出应明确标记。直接忽略这些问题会让 Athena 的工具能力表面可用但边界不可靠；因此选择在工具层、安全清理、测试覆盖三个方向同时改进。

**Actions taken**: 修改 `internal/tools/tools_v2.go`：为 `TermExecOutput` 增加 `workdir` 和 `timed_out` 字段；对 workspace 先执行 `filepath.EvalSymlinks`，对候选 workdir 先做字面路径内置检查，`MkdirAll` 后再对真实 workdir 执行 `EvalSymlinks` 并重新验证，拒绝 symlink 逃逸；新增 `pathInsideWorkspace()` 统一判断；执行命令时设置 `syscall.SysProcAttr{Setpgid: true}`，timeout 后通过负 PID `syscall.Kill(-pid, SIGKILL)` 杀进程组；timeout 输出追加 `[timeout after N seconds]`，并设置 `timed_out=true`；`truncateOutput` 改为 UTF-8 安全截断，避免截断中文多字节字符。重写 `internal/tools/tools_v2_test.go`：补充默认 workdir、workspace 内绝对路径、workspace 外路径拒绝、symlink escape 拒绝、timeout、timeout clamp、UTF-8 截断测试。修改 `internal/core/prompts.go` 和 `internal/hr/hr.go`，将缺失上下文处理表述调整为“当前角色可用工具”，避免诱导角色调用不可用工具。清理 `TASK.md`、`data/workspace/2c85fa59/scripts/qqq_emd_analysis.py`、`data/workspace/2c85fa59/output/README.md` 中的明文/占位 API key；脚本改为只从 `MASSIVE_API_KEY` 或 `POLYGON_API_KEY` 环境变量读取 key，缺失时显式报错；`run_all.sh` 增加环境变量检查。删除临时测试脚本、`__pycache__` 和 workspace 内 venv，避免将依赖目录和含 key 的临时文件纳入未跟踪产物。

**Findings / outcomes**: 验证命令全部通过：1) `go test ./internal/tools -run 'Test.*Term|TestTruncate' -v` 通过 7 个测试；2) `CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go test ./...` 全量通过；3) 同环境下 `go vet ./...` 通过；4) `go build ./...` 通过；5) `go test -race ./internal/tools` 通过；6) `git diff --check` 通过。敏感信息扫描未发现真实 Massive/Polygon key 或前序 `<redacted-...>` 占位残留；仅 README 中 `sk-...` 示例和配置字段名属于文档示例/结构字段。QQQ 产物复验通过：`qqq_data.csv` 与 `qqq_vw_fitted.csv` 均为 499 行，日期范围 `2024-05-06` 到 `2026-05-01`，日期完全对齐；VW min/max/mean 为 `420.6407 / 673.7368 / 537.7517945891783`；拟合曲线 min/max/mean 为 `438.3216 / 673.7368 / 536.9806470098226`；残差 MAE/RMSE 为 `7.874104178324127 / 10.432424247895515`；PNG 大小 `675141` 字节。当前 git 状态仍包含预期代码修改和 Athena 运行产物：`TASK.md`、prompt 相关文件、`internal/tools/tools_v2.go`、新增 `internal/tools/tools_v2_test.go`、`data/workspace/2c85fa59/`、`data/agents/2c85fa59-*`、`data/roles/`；SQLite 运行时文件已恢复。

**Implications for next steps**: 当前 Athena 的 `term` 工具比前一版更可靠：能防 symlink 逃逸、能返回真实 workdir、能标记 timeout、能减少残留进程风险、能保证输出 UTF-8 可读。后续提交前仍需决定是否将 `data/workspace/2c85fa59/` 和 `data/agents/2c85fa59-*` 作为示例产物纳入版本库；如果不希望提交运行产物，应将其移动到 artifacts 目录、加入 `.gitignore`，或单独打包归档。


[2026-05-05 00:13] 新增 Athena 安装脚本并验证 PATH 写入流程

**Objective**: 为 Athena 增加一键安装脚本，使用户 clone 项目后可以构建 `athena` 二进制、安装到用户可写目录，并自动把安装目录加入 shell PATH，降低首次安装门槛。

**Reasoning trail**: 现有 README 只提供 `go build -o athena ./cmd/athena` 的手动构建方式，生成的二进制停留在项目根目录，用户需要手动处理 PATH 和配置文件。用户明确要求“安装完成后会加入 path”，因此安装脚本需要完成三件事：1) 使用项目已有 CGO/FTS5 参数构建；2) 安装到默认用户目录 `~/.local/bin` 或用户指定目录；3) 根据 shell 类型把 PATH 写入 `.bashrc`、`.zshrc` 或 fish config。为了避免强依赖 root 权限，默认不安装到 `/usr/local/bin`；但通过 `ATHENA_INSTALL_DIR` 或 `--dir` 支持自定义。配置文件也应提供默认创建逻辑，避免安装后运行找不到配置。

**Actions taken**: 新增 `install.sh`：支持 `--dir DIR`、`--config-dir DIR`、`--no-config`、`--help`；默认安装到 `$HOME/.local/bin`，默认配置目录 `$HOME/.config/athena`；使用 `CGO_CFLAGS=-DSQLITE_ENABLE_FTS5`、`CGO_LDFLAGS=-lm`、`GOTOOLCHAIN=auto` 构建 `.build/athena`；通过 `install -m 0755` 安装二进制；若配置不存在且 `config/athena.example.yaml` 存在，则复制为 `athena.yaml`；检测当前 PATH，若未包含安装目录，则按 `$SHELL` 写入 `.bashrc`、`.zshrc` 或 `~/.config/fish/config.fish`；最后打印 `athena -config ...` 运行命令和当前 shell 刷新提示。更新 `Makefile`：增加 `install` target，执行 `./install.sh`。更新 `README.md` 和 `README_ZH.md`：安装段改为 `./install.sh`，补充默认安装路径、配置路径、选项示例和手动构建命令。

**Findings / outcomes**: 验证通过：`bash -n install.sh` 无语法错误；`./install.sh --help` 可输出帮助；使用临时目录 `ATHENA_INSTALL_DIR=/tmp/athena-install-test-bin ATHENA_CONFIG_DIR=/tmp/athena-install-test-config ./install.sh --no-config` 可成功构建并安装；`/tmp/athena-install-test-bin/athena -h` 输出 `-config` 参数说明；`go test ./...`、`go vet ./...`、`go build ./...` 和 `git diff --check` 均通过。测试安装时脚本按预期向 `~/.bashrc` 写入临时 PATH；验证后已删除该临时 PATH 行，并清理 `.build`、`/tmp/athena-install-test-bin`、`/tmp/athena-install-test-config` 等测试产物。

**Implications for next steps**: Athena 现在具备面向用户的一键安装入口。后续可考虑增加卸载脚本或 `./install.sh --uninstall`，并在 release workflow 中提供预编译二进制，减少用户本地 Go/GCC 依赖。


[2026-05-05 00:38] 分析 prompt 捕获并改进 Athena prompt/soul/tool 的预执行分析

**Objective**: 用户要求检查 `/home/debian/prompt-engineering` 中捕获的 prompt，排除 `role=user` 的消息，对其他角色提示词与工具轨迹进行分类总结，保存到同目录 `research.md`，并基于分析结果改进 Athena 的提示词、soul、tool 相关代码。该任务同时包含研究产出与代码修改，属于复杂多步骤任务，需要保留自包含进度记录。

**Reasoning trail**: 捕获数据不是单个 prompt 文件，而是一组 OpenAI 兼容 request/response JSON。直接人工读取少量文件会遗漏工具轨迹与重复模式，因此应先用脚本统计所有 request 中的 `body.messages`，按 role 排除 user，统计 developer/system/assistant/tool 的分布、去重主提示词、抽取高频段落与工具调用频率。分析目标不是复制 Hermes 的完整 Persona 到 Athena；Hermes prompt 包含大量 Hermes 专属 memory、skills、消息平台、CLI/WSL 环境与当前会话状态，直接移植会导致 Athena agent 上下文膨胀并引入不可执行指令。应迁移的是通用模式：工具行动、事实取证、前置检查、缺上下文处理、状态显式化、验证闭环、可追溯交付。Athena 现有 `internal/core/prompts.go` 已有 `commonBehaviorProtocol()`，`internal/hr/hr.go` 已可生成 7 层 soul，`internal/tools/tools_v2.go` 已有 term/file/memory/review 工具；因此改进应以增量增强为主，避免推翻既有 PM/Developer/HR 架构。

**Actions taken**: 读取 `/home/debian/athena/TASK.md` 以继承历史上下文；检查 `/home/debian/prompt-engineering/captures/` 文件结构；用 Python 脚本统计 291 个 request 文件、59,622 条消息，其中非 user 消息 58,159 条，developer 289 条、system 2 条、assistant 28,592 条、tool 29,276 条；提取 developer/system prompt 到 `/home/debian/prompt-engineering/_prompt_extract/` 供去重分析；统计工具调用频率，发现 terminal/read_file/search_files/todo/execute_code/patch/process/skill_view 是主要模式；生成研究报告 `/home/debian/prompt-engineering/research.md`。随后读取 `internal/core/prompts.go`、`internal/hr/hr.go`、`internal/tools/tools_v2.go`、`internal/core/agent_loop.go`、`internal/core/agent_loop_v2.go` 和根目录 `SOUL.md`，定位可改进点。

**Findings / outcomes**: 研究结论显示 captured prompt 的核心不是更长的人格设定，而是 7 类执行协议：行动优先、工具取证、前置检查、缺上下文先查后问、过程状态显式化、完成前验证、结果可追溯。Athena 已覆盖其中一部分，但仍有四个缺口：1) system prompt 没有动态列出该 agent 实际绑定的工具名与描述，可能诱导调用不存在工具；2) ReAct loop 达到最大迭代次数时没有显式写入阻塞/未完成记录；3) `file_read`/`file_write` 的路径校验仍是字符串前缀检查，弱于已修复的 `term` 工具真实路径和 symlink 校验；4) 根目录 `SOUL.md` 仍是 TBD，没有形成 Athena 默认人格模板。拒绝方案：不把 Hermes 的 memory、skills 列表和用户画像写入 Athena soul，因为这些内容属于 Hermes 运行时与当前用户上下文，不是 Athena 角色的稳定系统设计。

**Implications for next steps**: 下一步应实施四类变更：1) 扩展 `commonBehaviorProtocol()` 的证据记录与 memory/blackboard 边界；2) 在 `buildSystemPrompt()` 中注入实际工具清单；3) harden `file_read`/`file_write` 路径解析并补测试；4) 更新 `SOUL.md`。完成后必须运行 `gofmt`、工具单测、全量 `go test ./...`、`go vet ./...`、`go build ./...` 和 `git diff --check`。

[2026-05-05 14:27] Athena 职责泛化、自我改写 prompt 与动态工具能力的预执行分析

**Objective**: 用户要求检查 Athena 代码，并将 Agent 职责从写死工具/代码方案转向运行时自我诊断、自我发现不足、生成或修改新的 prompt，同时允许 tool 动态修改和加载。用户明确要求内建一个 Python 和一个 Python 输入接口。本条记录用于在修改前固定需求边界、现有架构证据、可行方案和拒绝方案。

**Reasoning trail**: 当前 Athena 已有 HR 动态角色生成、角色级 prompt/soul、黑板、memory、ReAct loop、动态可用工具清单、term/file/memory/review 等工具，但工具集合仍在 `AgentLoop.createTools()` 中按角色类别静态装配。prompt 层已经强调“发现问题先定位根因”和“缺上下文查工具”，但缺少显式的“自我诊断 → prompt 修订 → 能力缺口记录 → 重新加载工具”的闭环。仅继续给 PM/Developer 写更细分规则会让系统变成更多硬编码职责；更符合用户目标的方案是增加元认知工具和受控动态工具机制：Agent 发现自身能力不足时，不直接请求人类补代码，而是写入/读取自己的 `soul.md` 补丁、记录缺口、生成可执行 Python 脚本工具，并让下一轮 ReAct loop 根据工具目录重新构造工具集。这样职责泛化主要由 prompt 协议和工具接口承载，而不是把每种任务方案写死在 Go 代码里。

**Actions taken**: 已读取 `TASK.md`、`internal/core/prompts.go`、`internal/core/agent_loop.go`、`internal/core/agent_loop_v2.go`、`internal/tools/tools_v2.go`，并检查 `git status --short` 与近期提交。关键证据：1) `BuildRolePrompt()` 已支持 `{dataDir}/agents/{agentID}/soul.md` 和 HR role library soul；2) `buildSystemPrompt()` 已追加 `# 当前可用工具`，但只基于当次 `createTools()` 返回值；3) `createTools()` 静态分配 PM/dev/tester/reviewer/designer 工具；4) `runReActLoop()` 目前固定使用初始 `agentTools/toolInfos`，即使运行中写入新的工具定义也不会在同一轮重新绑定；5) `tools_v2.go` 已有 workspace 安全路径解析、term timeout、file read/write、memory read/write，但没有 Python 输入接口、没有 prompt 自我修订接口、没有动态工具注册接口。

**Findings / outcomes**: 需要实施四类最小架构变更。第一，prompt 层新增“自我改进协议”：每当工具失败、能力不足、方案过度依赖硬编码、或任务需要未列工具时，Agent 必须先记录 capability_gap，再提出 prompt/tool 改进，而不是直接结束。第二，工具层新增 `self_assess` 与 `prompt_patch`：让 Agent 能以结构化方式检查最近黑板/错误/memory/当前 soul，并把改进追加到自己的 `soul.md`。第三，Python 能力应分两层：内建 `python` 工具用于直接执行输入代码；`tool_create_python` 用于在 workspace 的 `.athena/tools/*.py` 注册可复用 Python 工具，`dynamic_python_tool` 用于按名称执行这些脚本。第四，Agent loop 需要在每次 LLM 调用前重新构造工具清单和 system prompt，或至少在每个 task/steer 开始时重新装配，否则动态工具写入后不能被当前运行的 agent 感知。拒绝方案：不把大量领域工具（例如金融、HTTP、绘图、爬虫）逐一写死到 Go 中，因为这与“职责泛化”目标相反；不允许 Python 工具逃逸 workspace 或任意读取系统路径，因为已有工具安全边界会被绕过；不引入复杂插件 ABI 或热编译 Go plugin，因为当前 Go/Eino 项目最小闭环更适合先用 Python 脚本工具作为动态扩展层。

**Implications for next steps**: 后续执行应修改 `internal/tools/tools_v2.go` 增加 Python/self/prompt/dynamic tool primitives，修改 `internal/core/agent_loop.go` 让 dev/tester/pm 等角色能获得这些元工具，并修改 `internal/core/agent_loop_v2.go` 在 ReAct 循环中刷新工具清单。完成后补充 `internal/tools/tools_v2_test.go`，至少覆盖：Python 执行成功、Python timeout、禁止 workspace 外路径、动态工具创建与执行、prompt patch 写入 soul。验证命令使用 `gofmt`、`CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go test ./internal/tools -v`、`go test ./...`、`go vet ./...`、`go build ./...` 和 `git diff --check`。

[2026-05-05 14:35] 完成 Athena 职责泛化、自我改写 prompt 与动态 Python 工具接口实现

**Objective**: 将 Athena 从固定角色职责和固定工具集，推进到“运行时发现自身不足 → 自我诊断 → 修订自身 prompt/soul → 生成可复用 Python 工具 → 刷新工具清单继续执行”的闭环。重点满足用户明确要求：内建 Python 和 Python 输入接口，同时让 tool 可动态修改和加载。

**Reasoning trail**: 现有 `term` 可以执行 Python 脚本，但这不是内建 Python 输入接口；Agent 需要先写文件或拼 shell，prompt 中也无法区分“一次性 Python 探查”和“可复用工具”。因此新增独立 `python` 工具，输入字段即 Python code，底层仍复用已加固的 workspace/timeout/进程组执行逻辑。动态工具方面，没有必要引入 Go plugin 或运行时编译，因为当前 Athena 的安全边界在 workspace 内，Python 脚本工具更适合作为最小动态扩展层；通过 `.athena/tools/*.py` 保存工具脚本和 `.md` 元数据，既可追踪又可由 prompt 动态列出。prompt 自我修改方面，不能让 Agent 任意覆盖完整系统提示词，否则容易破坏身份和安全协议；因此采用 append-only 的 `prompt_patch`，只把稳定行为改进追加到该 Agent 的 `soul.md`，由下一轮 prompt refresh 注入。为避免动态工具写入后无法被当前进程感知，ReAct loop 在每次 task/steer 开始和每轮 tool call 后刷新工具绑定与 system prompt。

**Actions taken**: 修改 `internal/tools/tools_v2.go`：1) 抽出 `runShellCommand()`，复用 term 的 workspace、symlink、防逃逸、timeout、进程组 kill、UTF-8 截断逻辑；2) 新增 `NewPythonExecTool()`，工具名 `python`，字段为 `code/workdir/timeout`，通过 stdin heredoc 执行 Python 输入；3) 新增 `NewPythonToolCreateTool()`，工具名 `tool_create_python`，将可复用脚本保存到 workspace `.athena/tools/{name}.py`，元数据保存到 `{name}.md`；4) 新增 `NewDynamicPythonToolRunner()`，工具名 `dynamic_python_tool`，按 name 执行 `.athena/tools/{name}.py` 并传入 argv[1]；5) 新增 `DynamicPythonToolInventory()`，用于 prompt 中列出已注册动态工具；6) 新增 `NewSelfAssessTool()`，读取当前 `soul.md`、`memory.md` 和动态工具目录位置，给出下一步选择 guidance；7) 新增 `NewPromptPatchTool()`，append-only 写入 `data/agents/{agentID}/soul.md`。修改 `internal/core/agent_loop.go`：所有角色获得 `self_assess`、`prompt_patch`、`dynamic_python_tool`；PM/dev/tester 获得 `python` 与 `tool_create_python`；动态工具清单被追加到 `# 当前可用工具`。修改 `internal/core/agent_loop_v2.go`：每个 task/steer 前重新装配工具和 system prompt；每轮工具调用后再次刷新工具和 system prompt，使 prompt_patch 和 tool_create_python 能在当前运行进程后续轮次生效；工具错误写入黑板。修改 `internal/core/prompts.go`：通用行为协议加入能力缺口处理要求，明确 self_assess、prompt_patch、tool_create_python、python 的分工；工具规范中加入自我诊断和动态 Python 工具规则。修改 `internal/tools/tools_v2_test.go`：增加 Python 执行、Python timeout、Python workspace 拒绝、动态 Python 工具创建/执行/清单、prompt_patch 写入 soul 的测试。

**Findings / outcomes**: 验证全部通过。`CGO_CFLAGS='-DSQLITE_ENABLE_FTS5' CGO_LDFLAGS='-lm' GOTOOLCHAIN=auto go test ./internal/tools -v` 通过 14 个工具测试，其中新增测试覆盖 `TestPythonExecToolRunsInputCodeInsideWorkspace`、`TestPythonExecToolRejectsOutsideWorkDir`、`TestPythonExecToolTimeout`、`TestDynamicPythonToolCreateAndRun`、`TestPromptPatchAppendsSoul`。全量 `go test ./...` 通过；`go vet ./...` 通过；`go build ./...` 通过；`git diff --check` 通过。当前工作区仍包含此前安装脚本、README、SOUL/IDENTITY 等未提交修改，本次核心代码变更集中在 `internal/core/agent_loop.go`、`internal/core/agent_loop_v2.go`、`internal/core/prompts.go`、`internal/tools/tools_v2.go`、`internal/tools/tools_v2_test.go` 与 `TASK.md`。

**Implications for next steps**: Athena 现在具备一个更泛化的自我改进路径：Agent 不必等待 Go 代码写死新工具；遇到稳定能力缺口时可先 `self_assess`，再用 `prompt_patch` 修订自己的 soul，或用 `tool_create_python` 生成可复用 Python 工具，并由刷新后的 prompt 暴露在工具清单中。后续仍可继续增强：1) 给动态工具增加 manifest JSON 和输入 schema；2) 增加 tool 删除/禁用接口；3) 给 prompt_patch 增加去重和版本历史；4) 在 UI 中展示 `.athena/tools` 和 soul patch 历史。
