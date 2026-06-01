package process

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jrperin/copilot-go-proxy/internal/auth"
)

type StatusReport struct {
	Running       bool   `json:"running"`
	PID           int    `json:"pid,omitempty"`
	Port          int    `json:"port"`
	Authenticated bool   `json:"authenticated"`
	APIHealthy    bool   `json:"api_healthy"`
	LogPath       string `json:"log_path"`
	OrphanPIDs    []int  `json:"orphan_pids,omitempty"`
}

func GetStatus(port int) StatusReport {
	report := StatusReport{
		Port:          port,
		Authenticated: auth.IsAuthenticated(),
		LogPath:       LogFilePath(),
	}

	if pid, err := LoadPID(); err == nil {
		report.PID = pid
		report.Running = IsProcessAlive(pid)
	}

	// Detect orphan processes (running but no PID file)
	procs := FindRunningProxies()
	if !report.Running && len(procs) > 0 {
		report.OrphanPIDs = procs
		report.Running = true
		report.PID = procs[0]
	}

	if report.Running {
		report.APIHealthy = checkAPIHealth(port)
	}

	return report
}

func checkAPIHealth(port int) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		return false
	}

	status, _ := health["status"].(string)
	return status == "ok"
}

func Diagnose(port int) map[string]interface{} {
	result := map[string]interface{}{}

	// 1. Auth check
	result["authenticated"] = auth.IsAuthenticated()

	// 2. PID file
	pid, pidErr := LoadPID()
	result["pid_file_exists"] = pidErr == nil
	if pidErr == nil {
		result["pid"] = pid
		result["process_alive"] = IsProcessAlive(pid)
	}

	// 3. Running processes
	procs := FindRunningProxies()
	result["running_processes"] = procs

	// 4. Port check
	result["port"] = port
	result["api_healthy"] = checkAPIHealth(port)

	// 5. Log file
	logPath := LogFilePath()
	result["log_path"] = logPath
	if data, err := readLastLines(logPath, 10); err == nil {
		result["log_tail"] = data
	}

	return result
}

func readLastLines(path string, n int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := splitLines(string(data))
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result, nil
}

func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	// Remove trailing empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
