package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jrperin/copilot-go-proxy/internal/auth"
)

const (
	pidFile = "copilot-proxy.pid"
	logFile = "copilot-proxy.log"
)

func PIDFilePath() string {
	return filepath.Join(auth.DataDir(), pidFile)
}

func LogFilePath() string {
	return filepath.Join(auth.DataDir(), logFile)
}

func SavePID(pid int) error {
	if err := auth.EnsureDataDir(); err != nil {
		return err
	}
	return os.WriteFile(PIDFilePath(), []byte(strconv.Itoa(pid)), 0644)
}

func LoadPID() (int, error) {
	data, err := os.ReadFile(PIDFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("not running (no PID file)")
		}
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("corrupt PID file: %w", err)
	}

	return pid, nil
}

func RemovePID() error {
	return os.Remove(PIDFilePath())
}

func IsProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks if process exists without sending a signal
	err = proc.Signal(os.Signal(nil))
	return err == nil
}

func KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// FindRunningProxies finds all copilot-proxy processes via /proc (Linux)
func FindRunningProxies() []int {
	var pids []int

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return pids
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
		cmdline, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if strings.Contains(cmdStr, "copilot-proxy") && strings.Contains(cmdStr, "start") {
			pids = append(pids, pid)
		}
	}

	return pids
}
