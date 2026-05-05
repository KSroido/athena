package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func runTermForTest(t *testing.T, workspace string, input TermExecInput) (string, error) {
	t.Helper()
	termTool, err := NewTermExecTool(workspace)
	if err != nil {
		t.Fatalf("NewTermExecTool: %v", err)
	}
	args, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal input: %v", err)
	}
	return termTool.InvokableRun(context.Background(), string(args))
}

func TestTermExecToolDefaultWorkDirAndTimeout(t *testing.T) {
	workspace := t.TempDir()

	result, err := runTermForTest(t, workspace, TermExecInput{
		Command: "pwd; mkdir -p nested; echo ok > nested/out.txt; test -f nested/out.txt && echo exists",
		Timeout: 10,
	})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, workspace) || !strings.Contains(result, "exists") {
		t.Fatalf("unexpected result: %s", result)
	}
	if _, err := os.Stat(filepath.Join(workspace, "nested", "out.txt")); err != nil {
		t.Fatalf("expected output file in workspace: %v", err)
	}
}

func TestTermExecToolAllowsAbsolutePathInsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	workdir := filepath.Join(workspace, "inside")

	result, err := runTermForTest(t, workspace, TermExecInput{
		Command: "pwd; echo ok",
		WorkDir: workdir,
		Timeout: 5,
	})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, workdir) || !strings.Contains(result, "ok") {
		t.Fatalf("expected execution in absolute workspace subdir, got %s", result)
	}
}

func TestTermExecToolRejectsOutsideWorkDir(t *testing.T) {
	workspace := t.TempDir()

	_, err := runTermForTest(t, workspace, TermExecInput{
		Command: "pwd",
		WorkDir: "/tmp",
		Timeout: 5,
	})
	if err == nil || !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("expected outside workspace error, got %v", err)
	}
}

func TestTermExecToolRejectsSymlinkEscape(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	linkPath := filepath.Join(workspace, "outside-link")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	_, err := runTermForTest(t, workspace, TermExecInput{
		Command: "pwd",
		WorkDir: "outside-link",
		Timeout: 5,
	})
	if err == nil || !strings.Contains(err.Error(), "outside workspace after symlink resolution") {
		t.Fatalf("expected symlink escape rejection, got %v", err)
	}
}

func TestTermExecToolTimeout(t *testing.T) {
	workspace := t.TempDir()

	result, err := runTermForTest(t, workspace, TermExecInput{
		Command: "sleep 2; echo done",
		Timeout: 1,
	})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, `"exit_code":-1`) || !strings.Contains(result, `"timed_out":true`) {
		t.Fatalf("expected timeout exit_code -1 and timed_out true, got %s", result)
	}
	if !strings.Contains(result, "timeout after 1 seconds") {
		t.Fatalf("expected timeout message, got %s", result)
	}
}

func TestTermExecToolClampsTimeout(t *testing.T) {
	workspace := t.TempDir()

	result, err := runTermForTest(t, workspace, TermExecInput{
		Command: "echo ok",
		Timeout: 999,
	})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, "ok") || strings.Contains(result, `"timed_out":true`) {
		t.Fatalf("expected successful command with clamped timeout, got %s", result)
	}
}

func TestTruncateOutputKeepsUTF8Valid(t *testing.T) {
	input := strings.Repeat("测", 20)
	out := truncateOutput(input, 7)
	if !strings.Contains(out, "truncated") {
		t.Fatalf("expected truncation marker, got %q", out)
	}
	if !utf8.ValidString(out) {
		t.Fatalf("expected valid UTF-8, got %q", out)
	}
}

func runFileReadForTest(t *testing.T, workspace string, input FileReadInput) (string, error) {
	t.Helper()
	fileTool, err := NewFileReadTool(workspace)
	if err != nil {
		t.Fatalf("NewFileReadTool: %v", err)
	}
	args, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal input: %v", err)
	}
	return fileTool.InvokableRun(context.Background(), string(args))
}

func runFileWriteForTest(t *testing.T, workspace string, input FileWriteInput) (string, error) {
	t.Helper()
	fileTool, err := NewFileWriteTool(workspace)
	if err != nil {
		t.Fatalf("NewFileWriteTool: %v", err)
	}
	args, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal input: %v", err)
	}
	return fileTool.InvokableRun(context.Background(), string(args))
}

func TestFileReadWriteRejectPathEscape(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	_, err := runFileReadForTest(t, workspace, FileReadInput{Path: outsideFile})
	if err == nil || !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("expected read outside workspace rejection, got %v", err)
	}

	_, err = runFileWriteForTest(t, workspace, FileWriteInput{Path: filepath.Join("..", "escape.txt"), Content: "bad"})
	if err == nil || !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("expected write outside workspace rejection, got %v", err)
	}
}

func TestFileReadWriteRejectSymlinkEscape(t *testing.T) {
	workspace := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	readLink := filepath.Join(workspace, "read-link")
	if err := os.Symlink(outsideFile, readLink); err != nil {
		t.Fatalf("create read symlink: %v", err)
	}
	writeLink := filepath.Join(workspace, "write-link")
	if err := os.Symlink(outside, writeLink); err != nil {
		t.Fatalf("create write symlink: %v", err)
	}

	_, err := runFileReadForTest(t, workspace, FileReadInput{Path: "read-link"})
	if err == nil || !strings.Contains(err.Error(), "outside workspace after symlink resolution") {
		t.Fatalf("expected read symlink escape rejection, got %v", err)
	}

	_, err = runFileWriteForTest(t, workspace, FileWriteInput{Path: filepath.Join("write-link", "out.txt"), Content: "bad"})
	if err == nil || !strings.Contains(err.Error(), "outside workspace after symlink resolution") {
		t.Fatalf("expected write symlink escape rejection, got %v", err)
	}
}

func runPythonForTest(t *testing.T, workspace string, input PythonExecInput) (string, error) {
	t.Helper()
	pythonTool, err := NewPythonExecTool(workspace)
	if err != nil {
		t.Fatalf("NewPythonExecTool: %v", err)
	}
	args, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal input: %v", err)
	}
	return pythonTool.InvokableRun(context.Background(), string(args))
}

func TestPythonExecToolRunsInputCodeInsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	result, err := runPythonForTest(t, workspace, PythonExecInput{
		Code:    "from pathlib import Path\nPath('py-out.txt').write_text('ok')\nprint(6*7)",
		Timeout: 5,
	})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, "42") {
		t.Fatalf("expected Python output, got %s", result)
	}
	if data, err := os.ReadFile(filepath.Join(workspace, "py-out.txt")); err != nil || string(data) != "ok" {
		t.Fatalf("expected Python-created file, data=%q err=%v", string(data), err)
	}
}

func TestPythonExecToolRejectsOutsideWorkDir(t *testing.T) {
	workspace := t.TempDir()
	_, err := runPythonForTest(t, workspace, PythonExecInput{Code: "print('bad')", WorkDir: "/tmp", Timeout: 5})
	if err == nil || !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("expected outside workspace error, got %v", err)
	}
}

func TestPythonExecToolTimeout(t *testing.T) {
	workspace := t.TempDir()
	result, err := runPythonForTest(t, workspace, PythonExecInput{Code: "import time\ntime.sleep(2)", Timeout: 1})
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if !strings.Contains(result, `"timed_out":true`) || !strings.Contains(result, `"exit_code":-1`) {
		t.Fatalf("expected Python timeout, got %s", result)
	}
}

func TestDynamicPythonToolCreateAndRun(t *testing.T) {
	workspace := t.TempDir()
	createTool, err := NewPythonToolCreateTool(workspace)
	if err != nil {
		t.Fatalf("NewPythonToolCreateTool: %v", err)
	}
	code := "import json, sys\npayload=json.loads(sys.argv[1]) if len(sys.argv)>1 and sys.argv[1] else {}\nprint(json.dumps({'sum': payload.get('a',0)+payload.get('b',0)}))"
	args, _ := json.Marshal(PythonToolCreateInput{Name: "sum_tool", Description: "sum two numbers from JSON fields a and b", Code: code})
	result, err := createTool.InvokableRun(context.Background(), string(args))
	if err != nil {
		t.Fatalf("create dynamic tool: %v", err)
	}
	if !strings.Contains(result, "sum_tool") {
		t.Fatalf("unexpected create result: %s", result)
	}
	if _, err := os.Stat(filepath.Join(workspace, ".athena", "tools", "sum_tool.py")); err != nil {
		t.Fatalf("expected dynamic tool file: %v", err)
	}

	runner, err := NewDynamicPythonToolRunner(workspace)
	if err != nil {
		t.Fatalf("NewDynamicPythonToolRunner: %v", err)
	}
	runArgs, _ := json.Marshal(DynamicPythonToolInput{Name: "sum_tool", Input: `{"a":2,"b":5}`, Timeout: 5})
	runResult, err := runner.InvokableRun(context.Background(), string(runArgs))
	if err != nil {
		t.Fatalf("run dynamic tool: %v", err)
	}
	if !strings.Contains(runResult, `\"sum\": 7`) {
		t.Fatalf("expected dynamic tool output, got %s", runResult)
	}
	inventory := DynamicPythonToolInventory(workspace)
	if !strings.Contains(inventory, "sum_tool") || !strings.Contains(inventory, "sum two numbers") {
		t.Fatalf("expected inventory entry, got %s", inventory)
	}
}

func TestPromptPatchAppendsSoul(t *testing.T) {
	dataDir := t.TempDir()
	tool, err := NewPromptPatchTool(dataDir, "agent-1")
	if err != nil {
		t.Fatalf("NewPromptPatchTool: %v", err)
	}
	args, _ := json.Marshal(PromptPatchInput{Title: "测试规则", Rationale: "发现重复错误", Patch: "以后先运行最小复现。"})
	if _, err := tool.InvokableRun(context.Background(), string(args)); err != nil {
		t.Fatalf("prompt patch run: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dataDir, "agents", "agent-1", "soul.md"))
	if err != nil {
		t.Fatalf("read soul: %v", err)
	}
	if !strings.Contains(string(data), "测试规则") || !strings.Contains(string(data), "以后先运行最小复现") {
		t.Fatalf("expected soul patch, got %s", string(data))
	}
}
