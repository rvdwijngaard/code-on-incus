package tool

import "path/filepath"

// PiTool implements Tool for pi (https://github.com/mariozechner/pi-coding-agent)
type PiTool struct {
	permissionMode  string // "bypass" (default) or "interactive" — pi has no permission gate, so this is largely a no-op
	contextFilePath string // absolute path to sandbox context file inside container (set by SetAutoContextPath)
}

// NewPi creates a new pi tool instance
func NewPi() Tool { return &PiTool{} }

func (p *PiTool) Name() string { return "pi" }

func (p *PiTool) Binary() string { return "pi" }

// ConfigDirName returns the config directory for pi.
// Pi stores config in ~/.pi/agent/ (controlled by PI_CODING_AGENT_DIR).
func (p *PiTool) ConfigDirName() string { return ".pi/agent" }

func (p *PiTool) SessionsDirName() string { return "sessions-pi" }

// BuildCommand builds the pi launch command.
// When resume is true, passes --continue to auto-resume the last session,
// or --session <id> if a specific session ID is provided.
func (p *PiTool) BuildCommand(sessionID string, resume bool, resumeSessionID string) []string {
	cmd := []string{"pi"}
	if resume {
		if resumeSessionID != "" {
			cmd = append(cmd, "--session", resumeSessionID)
		} else {
			cmd = append(cmd, "--continue")
		}
	}
	return cmd
}

// DiscoverSessionID returns "" because pi handles session discovery internally
// via --continue (sessions are stored as .jsonl files under the sessions dir).
func (p *PiTool) DiscoverSessionID(stateDir string) string { return "" }

// GetSandboxSettings returns pi's settings to inject into settings.json.
// Context injection is handled via AutoContextFile (writes to ~/.pi/agent/AGENTS.md).
// Pi has no permission bypass system like Claude or opencode.
func (p *PiTool) GetSandboxSettings() map[string]interface{} {
	return map[string]interface{}{}
}

// SetPermissionMode implements ToolWithPermissionMode.
// Pi has no permission gate, so this is stored but has no effect on
// BuildCommand or GetSandboxSettings.
func (p *PiTool) SetPermissionMode(mode string) {
	p.permissionMode = mode
}

// SetAutoContextPath implements ToolWithAutoContextPath.
// Stores the absolute path to the sandbox context file so it can be
// referenced in pi's settings or command line.
func (p *PiTool) SetAutoContextPath(path string) {
	p.contextFilePath = path
}

// EssentialConfigFiles implements ToolWithConfigDirFiles.
func (p *PiTool) EssentialConfigFiles() []string {
	return []string{"settings.json", "models.json", "auth.json"}
}

// SandboxSettingsFileName implements ToolWithConfigDirFiles.
func (p *PiTool) SandboxSettingsFileName() string { return "settings.json" }

// StateConfigFileName implements ToolWithConfigDirFiles.
// Pi has no sibling state file.
func (p *PiTool) StateConfigFileName() string { return "" }

// AlwaysSetupConfig implements ToolWithConfigDirFiles.
// Pi needs config dir setup even without host config dir so that
// settings.json exists for sandbox context injection.
func (p *PiTool) AlwaysSetupConfig() bool { return true }

// GetContainerEnv implements ToolWithContainerEnv.
// Redirects pi's session storage directory to the workspace mount so
// session data persists across ephemeral container recreations.
// Without this, sessions live in ~/.pi/agent/sessions/ (inside the container)
// and are destroyed when the ephemeral container is deleted.
func (p *PiTool) GetContainerEnv(workspacePath string) map[string]string {
	return map[string]string{
		"PI_CODING_AGENT_SESSION_DIR": filepath.Join(workspacePath, ".pi-sessions"),
	}
}

// AutoContextFile implements ToolWithAutoContextFile.
// Pi loads AGENTS.md from ~/.pi/agent/AGENTS.md (global) or from parent directories.
func (p *PiTool) AutoContextFile() string { return ".pi/agent/AGENTS.md" }
