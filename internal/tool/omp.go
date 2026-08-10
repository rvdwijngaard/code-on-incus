package tool

import (
	"path/filepath"
)

// OmpTool implements Tool for omp (https://omp.sh), a fork of pi.
// See: https://github.com/can1357/oh-my-pi
//
// omp is API-compatible with pi at the env-var level (PI_CODING_AGENT_DIR,
// PI_CONFIG_DIR, PI_CODING_AGENT_SESSION_DIR) but uses different default
// paths: ~/.omp/agent/ instead of ~/.pi/agent/. Credentials are stored in
// a SQLite database (agent.db) rather than auth.json, so omp and pi are
// NOT interchangeable credentials-wise — this is a separate tool type,
// not a pi binary replacement.
type OmpTool struct {
	permissionMode  string // "bypass" (default) or "interactive" — omp has no permission gate, so this is largely a no-op
	contextFilePath string // absolute path to sandbox context file inside container (set by SetAutoContextPath)
}

// NewOmp creates a new omp tool instance
func NewOmp() Tool { return &OmpTool{} }

func (o *OmpTool) Name() string { return "omp" }

func (o *OmpTool) Binary() string { return "omp" }

// ConfigDirName returns the config directory for omp.
// Omp stores config in ~/.omp/agent/ (controlled by PI_CONFIG_DIR).
func (o *OmpTool) ConfigDirName() string { return mustBundle("omp").ConfigDir }

func (o *OmpTool) SessionsDirName() string { return "sessions-omp" }

// BuildCommand builds the omp launch command.
// When resume is true, passes --continue to auto-resume the last session.
// Filesystem setup (symlink creation) is handled by PreLaunch(), keeping
// BuildCommand free of shell metacharacters.
func (o *OmpTool) BuildCommand(sessionID string, resume bool, resumeSessionID string) []string {
	cmd := []string{"omp"}
	if resume {
		// Always use --continue; omp manages its own session discovery
		// in <PI_CODING_AGENT_DIR>/sessions/. COI session UUIDs are only
		// used for metadata (slot/profile/persistent).
		cmd = append(cmd, "--continue")
	}
	return cmd
}

// PreLaunch implements ToolWithPreLaunch.
// When contextFilePath is set, creates the ~/.omp/agent/ directory and symlinks
// APPEND_SYSTEM.md to the sandbox context file. Omp appends APPEND_SYSTEM.md to
// its system prompt without replacing AGENTS.md, preserving the user's own context.
// The symlink is idempotent — no content duplication on container restart.
// Commands are returned as argv slices (no shell interpretation), preventing injection.
func (o *OmpTool) PreLaunch() [][]string {
	if o.contextFilePath == "" {
		return nil
	}
	// Derive home directory from the context file path (set as <home>/SANDBOX_CONTEXT.md)
	homeDir := filepath.Dir(o.contextFilePath)
	ompAgentDir := filepath.Join(homeDir, ".omp", "agent")
	linkTarget := filepath.Join(ompAgentDir, "APPEND_SYSTEM.md")

	return [][]string{
		{"mkdir", "-p", ompAgentDir},
		{"ln", "-sf", o.contextFilePath, linkTarget},
	}
}

// DiscoverSessionID returns "" because omp handles session discovery internally
// via --continue (sessions are stored as .jsonl files under the sessions dir).
func (o *OmpTool) DiscoverSessionID(stateDir string) string { return "" }

// GetSandboxSettings returns omp's settings to inject into settings.json.
// Omp has no permission bypass system like Claude or opencode.
// Context injection is handled via PreLaunch (symlinks APPEND_SYSTEM.md).
func (o *OmpTool) GetSandboxSettings() map[string]interface{} {
	return map[string]interface{}{}
}

// SetPermissionMode implements ToolWithPermissionMode.
// Omp has no permission gate, so this is stored but has no effect on
// BuildCommand or GetSandboxSettings.
func (o *OmpTool) SetPermissionMode(mode string) {
	o.permissionMode = mode
}

// SetAutoContextPath implements ToolWithAutoContextPath.
// Stores the absolute path to the sandbox context file so it can be
// referenced in omp's settings or command line.
// The path must be absolute; relative paths are rejected.
func (o *OmpTool) SetAutoContextPath(path string) {
	if !filepath.IsAbs(path) {
		return // reject relative paths silently; callers always provide absolute paths
	}
	o.contextFilePath = path
}

// EssentialConfigFiles implements ToolWithConfigDirFiles.
func (o *OmpTool) EssentialConfigFiles() []string {
	return mustBundle("omp").Files
}

// SandboxSettingsFileName implements ToolWithConfigDirFiles.
func (o *OmpTool) SandboxSettingsFileName() string { return mustBundle("omp").SandboxSettingsFile }

// StateConfigFileName implements ToolWithConfigDirFiles.
// Omp has no sibling state file.
func (o *OmpTool) StateConfigFileName() string { return mustBundle("omp").StateFile }

// AlwaysSetupConfig implements ToolWithConfigDirFiles.
// Omp needs config dir setup even without host config dir so that
// settings.json exists for sandbox context injection.
func (o *OmpTool) AlwaysSetupConfig() bool { return mustBundle("omp").AlwaysSetup }

// GetContainerEnv implements ToolWithContainerEnv.
// Redirects omp's session storage directory to the workspace mount so
// session data persists across ephemeral container recreations.
//
// omp exposes two distinct env vars for this: PI_CODING_AGENT_DIR overrides
// the *entire* agent dir (config + sessions + memory + agent.db); the
// separate PI_CODING_AGENT_SESSION_DIR overrides only session storage. We
// deliberately set the session-only one, so user config (settings.json,
// models.json, AGENTS.md) and the APPEND_SYSTEM.md context symlink stay at
// the default ~/.omp/agent/ path where setupCLIConfig writes them. Setting
// PI_CODING_AGENT_DIR would move everything to the workspace and orphan
// every host-seeded config file — a silent regression.
// Without this override, sessions live in ~/.omp/agent/sessions/ (inside
// the container) and are destroyed when the ephemeral container is deleted.
func (o *OmpTool) GetContainerEnv(workspacePath string) map[string]string {
	return map[string]string{
		"PI_CODING_AGENT_SESSION_DIR": filepath.Join(workspacePath, ".omp-agent"),
	}
}
