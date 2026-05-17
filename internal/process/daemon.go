package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// StartDaemon starts the proxy as a background daemon process
func StartDaemon() error {
	// Check if already running
	if pid, err := LoadPID(); err == nil && IsProcessAlive(pid) {
		return fmt.Errorf("already running (PID %d)", pid)
	}

	// Clean up stale PID
	RemovePID()

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	// Open log file
	logF, err := os.OpenFile(LogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	cmd := exec.Command(execPath, "start", "--foreground")
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // create new session
	}

	if err := cmd.Start(); err != nil {
		logF.Close()
		return fmt.Errorf("starting process: %w", err)
	}

	pid := cmd.Process.Pid

	// Detach from child
	cmd.Process.Release()

	// Save PID
	if err := SavePID(pid); err != nil {
		return fmt.Errorf("saving PID: %w", err)
	}

	return nil
}

// StopDaemon stops the running daemon
func StopDaemon() error {
	pid, err := LoadPID()
	if err != nil {
		return fmt.Errorf("not running: %w", err)
	}

	if !IsProcessAlive(pid) {
		RemovePID()
		return fmt.Errorf("stale PID file removed (process %d not found)", pid)
	}

	if err := KillProcess(pid); err != nil {
		return fmt.Errorf("killing process %d: %w", pid, err)
	}

	RemovePID()
	return nil
}
