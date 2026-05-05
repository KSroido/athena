package hr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ksroido/athena/internal/db"
)

// AgentStarter is the interface for starting agents (injected by core)
type AgentStarter interface {
	StartAgentFromHR(agent *db.Agent, projectID string) error
}

// LLMCaller is the interface for calling LLM (injected by core).
// The adapter wrapping core.LLMClient is defined in agent_manager.go
// to avoid circular imports between hr and core packages.
type LLMCaller interface {
	ChatWithSystem(ctx context.Context, system, user string) (*db.LLMResponse, error)
}

// HR manages agent hiring, role generation, and role library.
//
// Role resolution order (per project):
//  1. Project-level roles: {dataDir}/roles/{projectID}/{role}.json — confirmed fit, direct use
//  2. Global roles: ~/.athena/roles/{role}.json + fitness check
//     - Fit → copy to project-level for future use
//     - Not fit → LLM regenerates with project context, save to project-level
//  3. Built-in seed templates: SeedTemplates map (generic, no fitness check needed)
//  4. LLM dynamic generation: HR generates soul on-the-fly, save to project-level
type HR struct {
	mainDB   *db.DB
	starter  AgentStarter
	llm      LLMCaller
	dataDir  string
	rolesDir string // ~/.athena/roles/ — global role library (reference + reuse)
}

// New creates a new HR instance
func New(mainDB *db.DB, starter AgentStarter, dataDir string) *HR {
	homeDir, _ := os.UserHomeDir()
	rolesDir := filepath.Join(homeDir, ".athena", "roles")

	h := &HR{
		mainDB:   mainDB,
		starter:  starter,
		dataDir:  dataDir,
		rolesDir: rolesDir,
	}

	// Ensure global roles directory exists and seed it
	h.initRolesDir()

	return h
}

// SetLLM injects the LLM client for dynamic role generation and fitness checks
func (h *HR) SetLLM(llm LLMCaller) {
	h.llm = llm
}

// RolesDir returns the global roles directory path
func (h *HR) RolesDir() string {
	return h.rolesDir
}

// ---------------------------------------------------------------------------
// Role Template
// ---------------------------------------------------------------------------

// RoleTemplate defines a role template for creating agents.
type RoleTemplate struct {
	Role        string   `json:"role"`        // e.g. "dev.backend.finance"
	Name        string   `json:"name"`        // e.g. "金融/量化开发工程师"
	Category    string   `json:"category"`    // e.g. "dev" — determines tool set
	Description string   `json:"description"` // Human-readable description
	Domain      string   `json:"domain"`      // Generation context: what project/domain this soul was created for
	Tools       []string `json:"tools"`       // tool names (auto-filled from category if empty)
	Soul        string   `json:"soul"`        // Full role soul prompt
}

// SeedTemplates are the built-in seed templates shipped with Athena.
// These are generic and project-agnostic — no fitness check needed.
var SeedTemplates = map[string]RoleTemplate{
	"pm": {
		Role: "pm", Name: "项目经理", Category: "pm",
		Description: "需求拆解、任务分配、验收交付。唯一有权招聘和分配任务的角色。",
	},
	"dev.frontend": {
		Role: "dev.frontend", Name: "前端开发工程师", Category: "dev",
		Description: "Web/移动端界面开发：HTML/CSS/JS、React/Vue、Canvas、响应式布局、浏览器兼容。",
	},
	"dev.backend": {
		Role: "dev.backend", Name: "后端开发工程师", Category: "dev",
		Description: "通用后端开发（API、业务逻辑、服务端架构）。如需求涉及特定领域应招聘对应专家。",
	},
	"dev.fullstack": {
		Role: "dev.fullstack", Name: "全栈开发工程师", Category: "dev",
		Description: "前后端均可，适合小型项目或原型阶段。大型项目应拆分为前端+后端专家。",
	},
	"tester": {
		Role: "tester", Name: "测试工程师", Category: "tester",
		Description: "功能测试、集成测试、回归测试。编写测试用例，执行测试，出具报告。",
	},
	"reviewer": {
		Role: "reviewer", Name: "代码审查员", Category: "reviewer",
		Description: "独立代码审查：正确性/健壮性/性能/安全性/可维护性。上下文与开发隔离。",
	},
	"designer": {
		Role: "designer", Name: "UI/UX设计师", Category: "designer",
		Description: "交互设计、用户体验、设计规范、组件库。",
	},
}

// CategoryToolMap defines which tools each top-level category gets.
var CategoryToolMap = map[string][]string{
	"pm":       {"blackboard_read", "blackboard_write", "assign_task", "hr_request", "file_read", "file_write", "memory_read", "memory_write"},
	"dev":      {"blackboard_read", "blackboard_write", "term", "file_read", "file_write", "submit_for_review", "memory_read", "memory_write"},
	"tester":   {"blackboard_read", "blackboard_write", "term", "file_read", "file_write", "memory_read", "memory_write"},
	"reviewer": {"blackboard_read", "blackboard_write", "file_read", "memory_read", "memory_write"},
	"designer": {"blackboard_read", "blackboard_write", "file_read", "file_write", "memory_read", "memory_write"},
}

// ---------------------------------------------------------------------------
// Roles Directories
// ---------------------------------------------------------------------------

// globalRolePath returns the path for a role in the global library
func (h *HR) globalRolePath(role string) string {
	return filepath.Join(h.rolesDir, role+".json")
}

// projectRolePath returns the path for a role in the project-level library
func (h *HR) projectRolePath(projectID, role string) string {
	return filepath.Join(h.dataDir, "roles", projectID, role+".json")
}

// projectRolesDir returns the project-level roles directory
func (h *HR) projectRolesDir(projectID string) string {
	return filepath.Join(h.dataDir, "roles", projectID)
}

// initRolesDir ensures the global roles directory exists and writes seed templates
func (h *HR) initRolesDir() {
	if err := os.MkdirAll(h.rolesDir, 0755); err != nil {
		log.Printf("[hr] warning: failed to create roles dir %s: %v", h.rolesDir, err)
		return
	}

	// Write seed templates as JSON files (only if file doesn't exist — user edits are preserved)
	for _, tmpl := range SeedTemplates {
		path := h.globalRolePath(tmpl.Role)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			tmplCopy := tmpl
			tmplCopy.Tools = GetToolsForCategory(tmpl.Category)
			h.writeRoleFile(path, &tmplCopy)
		}
	}

	// Write README
	readmePath := filepath.Join(h.rolesDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readmeContent := "# Athena Role Library\n\n" +
			"This directory contains global role templates for Athena agents.\n" +
			"These serve as a reference library — HR checks fitness before reusing them for a project.\n\n" +
			"## Role Resolution\n\n" +
			"1. Project-level: {dataDir}/roles/{projectID}/{role}.json (confirmed fit)\n" +
			"2. Global: this directory + fitness check (LLM evaluates if soul matches project)\n" +
			"3. Seed templates (generic, no fitness check needed)\n" +
			"4. LLM dynamic generation (when no match found)\n\n" +
			"## Adding Global Roles\n\n" +
			"Global roles are reference templates. HR will check if a global role's soul\n" +
			"fits the current project before reusing it. If not, a project-specific variant\n" +
			"is generated and saved to the project-level directory.\n"
		_ = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	}
}

// writeRoleFile writes a RoleTemplate to a JSON file
func (h *HR) writeRoleFile(path string, tmpl *RoleTemplate) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Printf("[hr] warning: failed to create dir for role file %s: %v", path, err)
		return
	}
	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		log.Printf("[hr] warning: failed to marshal role template: %v", err)
		return
	}
	_ = os.WriteFile(path, data, 0644)
}

// readRoleFile reads a RoleTemplate from a JSON file
func (h *HR) readRoleFile(path string) (*RoleTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tmpl RoleTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parse role file %s: %w", path, err)
	}
	return &tmpl, nil
}

// LoadAllGlobalRoles loads all role templates from ~/.athena/roles/
func (h *HR) LoadAllGlobalRoles() map[string]RoleTemplate {
	roles := make(map[string]RoleTemplate)
	entries, err := os.ReadDir(h.rolesDir)
	if err != nil {
		return roles
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(h.rolesDir, entry.Name())
		tmpl, err := h.readRoleFile(path)
		if err != nil {
			log.Printf("[hr] warning: failed to load role file %s: %v", path, err)
			continue
		}
		if len(tmpl.Tools) == 0 {
			tmpl.Tools = GetToolsForCategory(tmpl.Category)
		}
		roles[tmpl.Role] = *tmpl
	}
	return roles
}

// LoadProjectRoles loads all role templates from project-level directory
func (h *HR) LoadProjectRoles(projectID string) map[string]RoleTemplate {
	roles := make(map[string]RoleTemplate)
	dir := h.projectRolesDir(projectID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return roles
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		tmpl, err := h.readRoleFile(path)
		if err != nil {
			log.Printf("[hr] warning: failed to load project role file %s: %v", path, err)
			continue
		}
		if len(tmpl.Tools) == 0 {
			tmpl.Tools = GetToolsForCategory(tmpl.Category)
		}
		roles[tmpl.Role] = *tmpl
	}
	return roles
}

// SaveProjectRole writes a role template to the project-level directory
func (h *HR) SaveProjectRole(projectID string, tmpl *RoleTemplate) error {
	if tmpl.Role == "" {
		return fmt.Errorf("role ID cannot be empty")
	}
	if tmpl.Category == "" {
		tmpl.Category = InferCategory(tmpl.Role)
	}
	if len(tmpl.Tools) == 0 {
		tmpl.Tools = GetToolsForCategory(tmpl.Category)
	}
	path := h.projectRolePath(projectID, tmpl.Role)
	h.writeRoleFile(path, tmpl)
	log.Printf("[hr] saved project role %s for project %s to %s", tmpl.Role, projectID, path)
	return nil
}

// SaveGlobalRole writes a role template to the global library
func (h *HR) SaveGlobalRole(tmpl *RoleTemplate) error {
	if tmpl.Role == "" {
		return fmt.Errorf("role ID cannot be empty")
	}
	if tmpl.Category == "" {
		tmpl.Category = InferCategory(tmpl.Role)
	}
	if len(tmpl.Tools) == 0 {
		tmpl.Tools = GetToolsForCategory(tmpl.Category)
	}
	path := h.globalRolePath(tmpl.Role)
	h.writeRoleFile(path, tmpl)
	log.Printf("[hr] saved global role %s to %s", tmpl.Role, path)
	return nil
}

// ---------------------------------------------------------------------------
// Role Resolution (with fitness check)
// ---------------------------------------------------------------------------

// ResolvedRole is the result of role resolution
type ResolvedRole struct {
	Template      RoleTemplate
	Source        string // "project", "global_fit", "global_unfit_regen", "seed", "llm_generated", "fallback"
	FitnessPassed bool   // Whether fitness check passed (true for project/seed/llm, checked for global)
}

// ResolveRoleForProject looks up a role template with project-aware fitness checking.
//
// Resolution order:
//  1. Project-level roles — already confirmed fit, direct use
//  2. Global roles + fitness check — LLM evaluates if soul matches project context
//     - Fit → copy to project-level, use directly
//     - Not fit → LLM regenerates with project context, save to project-level
//  3. Seed templates — generic, no fitness check needed
//  4. Not found → LLM generates from scratch
func (h *HR) ResolveRoleForProject(role, projectID, reason string) (*ResolvedRole, error) {
	category := InferCategory(role)

	// 1. Project-level roles (already confirmed fit)
	if projRoles := h.LoadProjectRoles(projectID); projRoles != nil {
		if tmpl, ok := projRoles[role]; ok {
			return &ResolvedRole{
				Template:      tmpl,
				Source:        "project",
				FitnessPassed: true,
			}, nil
		}
	}

	// 2. Global roles + fitness check
	if globalRoles := h.LoadAllGlobalRoles(); globalRoles != nil {
		if tmpl, ok := globalRoles[role]; ok {
			// Seed roles in global dir don't need fitness check (they're generic)
			if _, isSeed := SeedTemplates[role]; isSeed {
				return &ResolvedRole{
					Template:      tmpl,
					Source:        "seed",
					FitnessPassed: true,
				}, nil
			}

			// Custom global role — check fitness
			fit, err := h.checkRoleFitness(tmpl, reason)
			if err != nil {
				log.Printf("[hr] fitness check failed (LLM unavailable), assuming not fit: %v", err)
				fit = false
			}

			if fit {
				// Fit → copy to project-level for future use
				_ = h.SaveProjectRole(projectID, &tmpl)
				return &ResolvedRole{
					Template:      tmpl,
					Source:        "global_fit",
					FitnessPassed: true,
				}, nil
			}

			// Not fit → need to regenerate with project context
			return &ResolvedRole{
				Template:      tmpl,
				Source:        "global_unfit_regen",
				FitnessPassed: false,
			}, nil
		}
	}

	// 3. Seed templates (not in global dir yet, but defined in code)
	if tmpl, ok := SeedTemplates[role]; ok {
		result := tmpl
		result.Tools = GetToolsForCategory(tmpl.Category)
		return &ResolvedRole{
			Template:      result,
			Source:        "seed",
			FitnessPassed: true,
		}, nil
	}

	// 4. Not found
	return &ResolvedRole{
		Template: RoleTemplate{
			Role:     role,
			Category: category,
			Tools:    GetToolsForCategory(category),
		},
		Source:        "not_found",
		FitnessPassed: false,
	}, nil
}

// ---------------------------------------------------------------------------
// Fitness Check
// ---------------------------------------------------------------------------

// checkRoleFitness uses LLM to evaluate whether a global role's soul fits the
// current project's needs. Returns true if the soul's domain expertise matches
// the project context described in `reason`.
func (h *HR) checkRoleFitness(tmpl RoleTemplate, projectReason string) (bool, error) {
	// No soul to check — always unfit
	if tmpl.Soul == "" {
		return false, nil
	}

	// No project context provided — assume fit (can't evaluate without context)
	if projectReason == "" {
		return true, nil
	}

	// No LLM available — assume fit (can't evaluate without LLM)
	if h.llm == nil {
		log.Printf("[hr] no LLM available for fitness check, assuming global role %s is fit", tmpl.Role)
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	system := `你是 Athena 系统的角色适配性评估器。你需要判断一个全局角色库中的角色定义（soul）是否适配当前项目的需求。

判断标准：
- soul 中的专业领域、工作流程、约束是否与项目需求匹配
- soul 是否针对某个特定行业/场景定制，而项目需求是另一个行业/场景
- 通用角色（如通用后端开发、通用前端开发）默认适配

只输出 JSON：{"fit": true} 或 {"fit": false, "reason": "不适配原因"}`

	// Truncate soul to avoid excessive token usage
	soulPreview := tmpl.Soul
	if len(soulPreview) > 2000 {
		soulPreview = soulPreview[:2000] + "\n...（已截断）"
	}

	user := fmt.Sprintf(`角色ID: %s
角色名称: %s
角色Domain: %s
角色描述: %s

Soul 内容摘要:
%s

当前项目招聘原因: %s

请判断此角色的 soul 是否适配当前项目需求。`, tmpl.Role, tmpl.Name, tmpl.Domain, tmpl.Description, soulPreview, projectReason)

	resp, err := h.llm.ChatWithSystem(ctx, system, user)
	if err != nil {
		return false, fmt.Errorf("LLM fitness check: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	// Extract JSON from response
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}

	// Simple parsing — just look for "fit": true/false
	if strings.Contains(content, `"fit": true`) || strings.Contains(content, `"fit":true`) {
		log.Printf("[hr] fitness check PASSED for role %s (project reason: %s)", tmpl.Role, truncateStr(projectReason, 50))
		return true, nil
	}

	log.Printf("[hr] fitness check FAILED for role %s (project reason: %s)", tmpl.Role, truncateStr(projectReason, 50))
	return false, nil
}

// ---------------------------------------------------------------------------
// Role Catalog
// ---------------------------------------------------------------------------

// RoleCatalog returns all available roles (project + global + seed) as a formatted string
func (h *HR) RoleCatalog(projectID string) string {
	allRoles := make(map[string]RoleTemplate)

	// Seed templates
	for k, v := range SeedTemplates {
		allRoles[k] = v
	}

	// Global roles (override seeds)
	for k, v := range h.LoadAllGlobalRoles() {
		allRoles[k] = v
	}

	// Project roles (override global)
	if projectID != "" {
		for k, v := range h.LoadProjectRoles(projectID) {
			allRoles[k] = v
		}
	}

	// Group by category
	groups := map[string][]RoleTemplate{}
	for _, tmpl := range allRoles {
		groups[tmpl.Category] = append(groups[tmpl.Category], tmpl)
	}
	for cat := range groups {
		sort.Slice(groups[cat], func(i, j int) bool {
			return groups[cat][i].Role < groups[cat][j].Role
		})
	}

	categories := []string{"pm", "dev", "tester", "reviewer", "designer"}
	var sb strings.Builder

	for _, cat := range categories {
		templates, ok := groups[cat]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("### %s\n", categoryDisplayName(cat)))
		for _, t := range templates {
			sb.WriteString(fmt.Sprintf("- `%s` — %s: %s\n", t.Role, t.Name, t.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("**注意：以上仅为已注册角色。你可以指定任意角色ID（如 `dev.backend.finance`），HR会自动生成对应的专业soul。**\n")
	sb.WriteString("**HR会检查全局角色库中的角色是否适配当前项目，不适配时会自动重新生成。**\n")

	return sb.String()
}

func categoryDisplayName(cat string) string {
	switch cat {
	case "pm":
		return "项目管理"
	case "dev":
		return "开发"
	case "tester":
		return "测试"
	case "reviewer":
		return "审查"
	case "designer":
		return "设计"
	default:
		return cat
	}
}

// InferCategory determines the top-level category from a role string
func InferCategory(role string) string {
	switch {
	case role == "pm":
		return "pm"
	case strings.HasPrefix(role, "dev.") || role == "dev":
		return "dev"
	case strings.HasPrefix(role, "tester"):
		return "tester"
	case role == "reviewer":
		return "reviewer"
	case strings.HasPrefix(role, "designer"):
		return "designer"
	default:
		return "dev"
	}
}

// GetToolsForCategory returns the tool set for a given category
func GetToolsForCategory(category string) []string {
	if tools, ok := CategoryToolMap[category]; ok {
		return tools
	}
	return CategoryToolMap["dev"]
}

// ---------------------------------------------------------------------------
// HireRequest & Hire
// ---------------------------------------------------------------------------

// HireRequest is a request to hire a new agent.
type HireRequest struct {
	Role       string `json:"role"`       // Role ID (e.g. "dev.backend.finance")
	Speciality string `json:"speciality"` // Extra speciality hint for LLM soul generation
	ProjectID  string `json:"project_id"`
	Reason     string `json:"reason"` // Why this role is needed (used in soul generation + fitness check)
}

// Hire creates and starts a new agent with project-aware role resolution.
func (h *HR) Hire(req *HireRequest) (*db.Agent, error) {
	// 1. Check company size limit
	maxAgents := 100
	var count int
	h.mainDB.DB().QueryRow("SELECT COUNT(*) FROM agents WHERE status != 'offline'").Scan(&count)
	if count >= maxAgents {
		return nil, fmt.Errorf("公司人数已达上限 (%d/%d)，请联系CEO扩容", count, maxAgents)
	}

	// 2. Resolve role with fitness check
	resolved, err := h.ResolveRoleForProject(req.Role, req.ProjectID, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("resolve role: %w", err)
	}

	tmpl := resolved.Template
	category := tmpl.Category
	if category == "" {
		category = InferCategory(req.Role)
	}

	// 3. Determine soul content
	var soulContent string
	needsCustomSoul := false

	switch resolved.Source {
	case "project", "global_fit", "seed":
		// Role confirmed fit — use existing soul (or built-in prompt for seeds)
		soulContent = tmpl.Soul

	case "global_unfit_regen":
		// Global role doesn't fit this project → regenerate with project context
		needsCustomSoul = true
		log.Printf("[hr] global role %s not fit for project %s, regenerating soul", req.Role, req.ProjectID)

	case "not_found":
		// No matching role anywhere → generate from scratch
		needsCustomSoul = true
	}

	// 4. Generate custom soul if needed
	if needsCustomSoul {
		soulContent, err = h.generateRoleSoul(req.Role, req.Speciality, category, req.Reason)
		if err != nil {
			log.Printf("[hr] LLM soul generation failed, using fallback: %v", err)
			soulContent = h.fallbackSoul(req.Role, req.Speciality, category)
		}

		// Save to project-level directory (project-specific)
		saveTmpl := &RoleTemplate{
			Role:        req.Role,
			Name:        extractRoleName(soulContent, req.Role),
			Category:    category,
			Description: req.Reason,
			Domain:      req.Reason, // Record the project context this soul was generated for
			Tools:       GetToolsForCategory(category),
			Soul:        soulContent,
		}
		_ = h.SaveProjectRole(req.ProjectID, saveTmpl)

		// Also update global library (as reference for other projects)
		globalTmpl := *saveTmpl
		_ = h.SaveGlobalRole(&globalTmpl)

		// Update tmpl for DB record
		tmpl = *saveTmpl
	}

	// 5. Check duplicate role in project
	var existingCount int
	h.mainDB.DB().QueryRow(
		"SELECT COUNT(*) FROM agents a JOIN project_members pm ON a.id = pm.agent_id WHERE pm.project_id = ? AND a.role = ? AND a.status != 'offline'",
		req.ProjectID, tmpl.Role,
	).Scan(&existingCount)
	if existingCount > 0 {
		return nil, fmt.Errorf("项目 %s 已有 %s 角色的 Agent", req.ProjectID, tmpl.Role)
	}

	// 6. Create agent record
	if len(tmpl.Tools) == 0 {
		tmpl.Tools = GetToolsForCategory(category)
	}

	agentID := fmt.Sprintf("%s-%s-%s", req.ProjectID, strings.ReplaceAll(tmpl.Role, ".", "-"), uuid.New().String()[:8])
	agent := &db.Agent{
		ID:        agentID,
		Name:      tmpl.Name,
		Role:      tmpl.Role,
		Status:    "idle",
		Tools:     toolsToJSON(tmpl.Tools),
		Model:     "default",
		CreatedBy: "hr",
		CreatedAt: time.Now(),
	}

	_, err = h.mainDB.DB().Exec(`
		INSERT INTO agents (id, name, role, status, tools, model, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Name, agent.Role, agent.Status, agent.Tools, agent.Model, agent.CreatedBy, agent.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert agent: %w", err)
	}

	// 7. Add to project members
	_, err = h.mainDB.DB().Exec(`
		INSERT INTO project_members (id, project_id, agent_id, role, joined_at)
		VALUES (?, ?, ?, ?, ?)
	`, uuid.New().String()[:8], req.ProjectID, agent.ID, agent.Role, time.Now())
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}

	// 8. Create agent data directory
	agentDir := filepath.Join(h.dataDir, "agents", agent.ID)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		log.Printf("[hr] warning: failed to create agent dir %s: %v", agentDir, err)
	}

	// 9. Write soul file
	soulPath := filepath.Join(agentDir, "soul.md")
	if soulContent != "" {
		_ = os.WriteFile(soulPath, []byte(soulContent), 0644)
	} else {
		_ = os.WriteFile(soulPath, []byte(fmt.Sprintf("# %s\n\n内置角色，使用系统提示词。\n", tmpl.Name)), 0644)
	}

	// 10. Initialize memory.md
	memoryPath := filepath.Join(agentDir, "memory.md")
	_ = os.WriteFile(memoryPath, []byte(fmt.Sprintf("# %s 个人记忆\n\n", tmpl.Name)), 0644)

	// 11. Start agent goroutine
	if err := h.starter.StartAgentFromHR(agent, req.ProjectID); err != nil {
		log.Printf("[hr] failed to start agent %s: %v", agent.ID, err)
	}

	log.Printf("[hr] hired %s (%s, role=%s, source=%s) for project %s",
		tmpl.Name, agent.ID, tmpl.Role, resolved.Source, req.ProjectID)
	return agent, nil
}

// ---------------------------------------------------------------------------
// Dynamic Soul Generation
// ---------------------------------------------------------------------------

// generateRoleSoul uses LLM to generate a role soul prompt for a custom role
func (h *HR) generateRoleSoul(role, speciality, category, reason string) (string, error) {
	if h.llm == nil {
		return "", fmt.Errorf("LLM not available for soul generation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	system := "你是 Athena 系统的 HR 角色设计师。你的任务是为一个新 Agent 生成完整的角色定义（soul）。\n\n" +
		"soul 必须严格遵循以下7层结构，每层都不可省略：\n\n" +
		"# 身份\n- 明确声明角色名称、专业领域、所属项目\n\n" +
		"# 通用行为协议\n- 使用简体中文，输出直接、技术化、可执行，避免隐喻和文学化表述\n" +
		"- 需要行动时必须调用可用工具执行，不要只描述计划或承诺后续动作\n" +
		"- 涉及文件内容、目录结构、系统状态、命令结果、Git状态、计算、测试、构建、端口和进程时，必须用工具验证\n" +
		"- 修改文件前先读取相关文件，修改后执行验证命令或读取结果确认\n" +
		"- 缺少上下文时优先使用当前角色可用工具读黑板、文件、记忆或执行查询；仍缺失则写黑板请求澄清\n" +
		"- 不确定结论标记为 conjecture；重要结论、进展、错误和验证结果必须写入黑板并附证据\n\n" +
		"# 核心原则\n- 5条以内，是该角色的行为底线和决策准则\n- 必须体现该角色的专业特性（区别于通用角色）\n\n" +
		"# 工作流程\n- 分阶段的步骤，每步包含具体操作\n- 必须包含与其他角色的协作点（何时读黑板、何时提交验收等）\n\n" +
		"# 工具使用规范\n- 列出该角色可用工具及使用场景\n- 开发类角色必须包含 submit_for_review\n\n" +
		"# 约束\n- 角色不可逾越的边界\n- 至少包含：禁止编造事实\n\n" +
		"# 自检清单\n- 完成任务前的逐项确认\n- 必须与原则和工作流程对应\n\n" +
		"要求：\n1. 专业性：原则和工作流程必须体现该角色的专业深度，不是泛泛而谈\n" +
		"2. 可操作性：每个步骤必须具体到可执行，不用模糊表述\n" +
		"3. 协作性：明确何时与谁协作，通过什么工具\n" +
		"4. 只输出 soul 内容本身，不要输出任何解释或元信息"

	user := fmt.Sprintf("请为以下角色生成 soul：\n\n- 角色ID: %s\n- 专业方向: %s\n- 角色大类: %s\n- 项目需求: %s\n\n请生成完整的7层 soul。", role, speciality, categoryDisplayName(category), reason)

	resp, err := h.llm.ChatWithSystem(ctx, system, user)
	if err != nil {
		return "", fmt.Errorf("LLM generate soul: %w", err)
	}

	return resp.Content, nil
}

// fallbackSoul generates a minimal soul without LLM
func (h *HR) fallbackSoul(role, speciality, category string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 身份\n\n你是 Athena 系统中的 **%s** Agent。\n", role))
	if speciality != "" {
		sb.WriteString(fmt.Sprintf("你的专业方向是：%s\n", speciality))
	}
	sb.WriteString(fmt.Sprintf("角色大类：%s\n\n", categoryDisplayName(category)))

	sb.WriteString("# 通用行为协议\n\n")
	sb.WriteString("- 使用简体中文，输出直接、技术化、可执行，避免隐喻和文学化表述\n")
	sb.WriteString("- 需要行动时必须调用可用工具执行，不要只描述计划或承诺后续动作\n")
	sb.WriteString("- 涉及文件内容、目录结构、系统状态、命令结果、Git状态、计算、测试、构建、端口和进程时，必须用工具验证\n")
	sb.WriteString("- 修改文件前先读取相关文件，修改后执行验证命令或读取结果确认\n")
	sb.WriteString("- 缺少上下文时优先使用当前角色可用工具读黑板、文件、记忆或执行查询；仍缺失则写黑板请求澄清\n")
	sb.WriteString("- 不确定结论标记为 conjecture；重要结论、进展、错误和验证结果必须写入黑板并附证据\n\n")

	sb.WriteString("# 核心原则\n\n")
	sb.WriteString("- 专业专注：只处理自己专业领域内的问题\n")
	sb.WriteString("- 事实驱动：所有结论基于实际验证，不确定的标记为 conjecture\n")
	sb.WriteString("- 协作优先：遇到非本领域问题，通过黑板请求其他 Agent 协助\n")
	sb.WriteString("- 产出可见：每完成一个阶段，写入黑板记录进展\n\n")

	sb.WriteString("# 工作流程\n\n")
	sb.WriteString("1. 读取黑板，理解任务要求和验收标准\n")
	sb.WriteString("2. 执行专业领域内的开发/测试/设计工作\n")
	sb.WriteString("3. 使用工具完成具体操作\n")
	if category == "dev" {
		sb.WriteString("4. 使用 submit_for_review 提交验收\n")
		sb.WriteString("5. 收到整改要求后逐一修复\n")
	} else {
		sb.WriteString("4. 将结果写入黑板\n")
	}
	sb.WriteString("\n")

	sb.WriteString("# 工具使用规范\n\n")
	for _, t := range GetToolsForCategory(category) {
		sb.WriteString(fmt.Sprintf("- %s\n", t))
	}
	sb.WriteString("\n")

	sb.WriteString("# 约束\n\n")
	sb.WriteString("- 禁止编造事实\n")
	if category == "dev" {
		sb.WriteString("- 禁止提交未完成的半成品\n")
		sb.WriteString("- term 命令不得包含危险操作\n")
	}
	sb.WriteString("\n")

	sb.WriteString("# 自检清单\n\n")
	sb.WriteString("1. 是否完全理解了任务要求？\n")
	sb.WriteString("2. 产出是否覆盖了所有验收标准？\n")
	if category == "dev" {
		sb.WriteString("3. 是否使用 submit_for_review 提交了验收？\n")
	}

	return sb.String()
}

// extractRoleName tries to extract the role name from soul content (first **...** line)
func extractRoleName(soul, fallback string) string {
	for _, line := range strings.Split(soul, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			name := strings.Trim(line, "*")
			if name != "" {
				return name
			}
		}
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// AvailableRolesList returns all role IDs (project + global + seed) as a sorted slice
func (h *HR) AvailableRolesList(projectID string) []string {
	allRoles := make(map[string]bool)
	for k := range SeedTemplates {
		allRoles[k] = true
	}
	for k := range h.LoadAllGlobalRoles() {
		allRoles[k] = true
	}
	if projectID != "" {
		for k := range h.LoadProjectRoles(projectID) {
			allRoles[k] = true
		}
	}
	roles := make([]string, 0, len(allRoles))
	for r := range allRoles {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	return roles
}

// ListCompany returns all agents
func (h *HR) ListCompany() ([]*db.Agent, error) {
	rows, err := h.mainDB.DB().Query(
		"SELECT id, name, role, status, model, created_by, created_at FROM agents ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*db.Agent
	for rows.Next() {
		a := &db.Agent{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Role, &a.Status, &a.Model, &a.CreatedBy, &a.CreatedAt); err != nil {
			continue
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// Fire removes an agent
func (h *HR) Fire(agentID string) error {
	_, err := h.mainDB.DB().Exec("UPDATE agents SET status = 'offline' WHERE id = ?", agentID)
	return err
}

// toolsToJSON converts a tool list to JSON string
func toolsToJSON(tools []string) string {
	if len(tools) == 0 {
		return "[]"
	}
	result := "["
	for i, t := range tools {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("\"%s\"", t)
	}
	result += "]"
	return result
}

// truncateStr truncates a string
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
