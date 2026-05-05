//go:build windows

package tools

import "os/exec"

func setProcessGroup(cmd *exec.Cmd) {
	// Windows: no Setpgid, rely on exec.CommandContext cancellation
}

func killProcessGroup(pid int) {
	// Windows: CommandContext handles process kill on timeout
}
