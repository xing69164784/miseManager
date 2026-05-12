package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ToolInfo struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Versions  []string `json:"versions"`
	Default   string   `json:"default"`
	Source    string   `json:"source"`
	Installed bool     `json:"installed"`
}

type MiseToolVersion struct {
	Version          string                 `json:"version"`
	RequestedVersion string                 `json:"requested_version"`
	InstallPath      string                 `json:"install_path"`
	Source           map[string]interface{} `json:"source"`
	Installed        bool                   `json:"installed"`
	Active           bool                   `json:"active"`
}

type RegistryTool struct {
	Short       string   `json:"short"`
	Backends    []string `json:"backends"`
	Description string   `json:"description"`
	Aliases     []string `json:"aliases"`
}

type MiseService struct{}

var (
	logFile *os.File
	logMu   sync.Mutex
)

func newCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	return cmd
}

func getLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mise-gui.log")
}

func initLog() {
	path := getLogPath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	logFile = f
}

func writeLog(format string, args ...interface{}) {
	if logFile == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	logMu.Lock()
	logFile.WriteString(line)
	logFile.Sync()
	logMu.Unlock()
}

func (m *MiseService) GetLogs() (string, error) {
	path := getLogPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (m *MiseService) ClearLogs() (string, error) {
	path := getLogPath()
	err := os.WriteFile(path, []byte{}, 0644)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (m *MiseService) ListInstalledTools() ([]ToolInfo, error) {
	writeLog("[MiseService] 执行命令: mise ls --json")
	cmd := newCmd("mise", "ls", "--json")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run mise ls: %v, output: %s", err, out.String())
	}

	var rawTools map[string][]MiseToolVersion
	err = json.Unmarshal(out.Bytes(), &rawTools)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mise ls output: %v", err)
	}

	var tools []ToolInfo
	for name, versions := range rawTools {
		tool := ToolInfo{
			Name:      name,
			Versions:  make([]string, 0, len(versions)),
			Installed: true,
		}

		for _, v := range versions {
			tool.Versions = append(tool.Versions, v.Version)

			if v.Active {
				tool.Version = v.Version
				tool.Default = v.Version
			}

			if v.Source != nil {
				if path, ok := v.Source["path"].(string); ok && tool.Source == "" {
					tool.Source = path
				}
			}
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

func (m *MiseService) InstallTool(name string, version string) (string, error) {
	arg := name + "@" + version
	writeLog("[MiseService] 执行命令: mise install %s", arg)
	cmd := newCmd("mise", "install", arg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	writeLog("[MiseService] 命令执行完成: mise install %s, err=%v, output=%s", arg, err, out.String())
	if err != nil {
		return out.String(), fmt.Errorf("failed to install %s: %v", arg, err)
	}

	return out.String(), nil
}

func (m *MiseService) UninstallTool(name string, version string) (string, error) {
	writeLog("[MiseService] 执行命令: mise uninstall %s %s", name, version)
	cmd := newCmd("mise", "uninstall", name, version)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("failed to uninstall %s@%s: %v", name, version, err)
	}

	return out.String(), nil
}

func (m *MiseService) SetDefaultVersion(name string, version string) (string, error) {
	arg := name + "@" + version
	writeLog("[MiseService] 执行命令: mise use --global %s", arg)
	cmd := newCmd("mise", "use", "--global", arg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("failed to set default %s@%s: %v", name, version, err)
	}

	return out.String(), nil
}

func (m *MiseService) GetAvailableVersions(name string) ([]string, error) {
	writeLog("[MiseService] 执行命令: mise list-remote %s", name)
	cmd := newCmd("mise", "list-remote", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote versions for %s: %v", name, err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var versions []string
	for _, line := range lines {
		if line != "" {
			versions = append(versions, line)
		}
	}

	return versions, nil
}

func (m *MiseService) GetToolInfo(name string) (*ToolInfo, error) {
	tools, err := m.ListInstalledTools()
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if tool.Name == name {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", name)
}

func (m *MiseService) GetVersion() (string, error) {
	cmd := newCmd("mise", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get mise version: %v", err)
	}

	return strings.TrimSpace(out.String()), nil
}

func (m *MiseService) Reshim() (string, error) {
	cmd := newCmd("mise", "reshim")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("failed to reshim: %v", err)
	}

	return out.String(), nil
}

var (
	registryCache     []RegistryTool
	registryCacheTime time.Time
	registryCacheMu   sync.RWMutex
)

func (m *MiseService) SearchRegistry(query string) ([]RegistryTool, error) {
	writeLog("[MiseService] 执行命令: mise registry --json (query=%s)", query)
	registryCacheMu.RLock()
	cached := registryCache
	cacheTime := registryCacheTime
	registryCacheMu.RUnlock()

	if len(cached) == 0 || time.Since(cacheTime) > 10*time.Minute {
		cmd := newCmd("mise", "registry", "--json")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to run mise registry: %v", err)
		}

		var allTools []RegistryTool
		err = json.Unmarshal(out.Bytes(), &allTools)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mise registry output: %v", err)
		}

		registryCacheMu.Lock()
		registryCache = allTools
		registryCacheTime = time.Now()
		registryCacheMu.Unlock()
		cached = allTools
	}

	if query == "" {
		return cached[:min(len(cached), 50)], nil
	}

	var nameMatches []RegistryTool
	var aliasMatches []RegistryTool
	var descMatches []RegistryTool
	lowerQuery := strings.ToLower(query)
	for _, tool := range cached {
		if strings.Contains(strings.ToLower(tool.Short), lowerQuery) {
			nameMatches = append(nameMatches, tool)
		} else {
			matched := false
			for _, alias := range tool.Aliases {
				if strings.Contains(strings.ToLower(alias), lowerQuery) {
					aliasMatches = append(aliasMatches, tool)
					matched = true
					break
				}
			}
			if !matched && strings.Contains(strings.ToLower(tool.Description), lowerQuery) {
				descMatches = append(descMatches, tool)
			}
		}
	}

	filtered := make([]RegistryTool, 0, len(nameMatches)+len(aliasMatches)+len(descMatches))
	filtered = append(filtered, nameMatches...)
	filtered = append(filtered, aliasMatches...)
	filtered = append(filtered, descMatches...)

	if len(filtered) > 50 {
		filtered = filtered[:50]
	}

	return filtered, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *MiseService) MiseDoctor() (string, error) {
	writeLog("[MiseService] 执行命令: mise doctor")
	cmd := newCmd("mise", "doctor")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	writeLog("[MiseService] 命令执行完成: mise doctor, err=%v", err)
	if err != nil {
		return out.String(), fmt.Errorf("mise doctor failed: %v", err)
	}

	return out.String(), nil
}

func (m *MiseService) RunCustomCommand(args string) (string, error) {
	argList := strings.Fields(args)
	if len(argList) == 0 {
		return "", fmt.Errorf("empty command")
	}

	if argList[0] != "mise" {
		argList = append([]string{"mise"}, argList...)
	}

	writeLog("[MiseService] 执行自定义命令: %s", strings.Join(argList, " "))
	cmd := newCmd(argList[0], argList[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	writeLog("[MiseService] 自定义命令执行完成, err=%v, output length=%d", err, out.Len())
	if err != nil {
		return out.String(), fmt.Errorf("command failed: %v", err)
	}

	return out.String(), nil
}
