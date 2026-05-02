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
