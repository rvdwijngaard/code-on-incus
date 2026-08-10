package tool

import (
	"testing"
)

func TestOmpTool_Basics(t *testing.T) {
	omp := NewOmp()

	if omp.Name() != "omp" {
		t.Errorf("Name() = %q, want %q", omp.Name(), "omp")
	}
	if omp.Binary() != "omp" {
		t.Errorf("Binary() = %q, want %q", omp.Binary(), "omp")
	}
	if omp.ConfigDirName() != ".omp/agent" {
		t.Errorf("ConfigDirName() = %q, want %q", omp.ConfigDirName(), ".omp/agent")
	}
	if omp.SessionsDirName() != "sessions-omp" {
		t.Errorf("SessionsDirName() = %q, want %q", omp.SessionsDirName(), "sessions-omp")
	}
}

func TestOmpTool_BuildCommand_NewSession(t *testing.T) {
	omp := NewOmp()
	cmd := omp.BuildCommand("some-session-id", false, "")
	if len(cmd) != 1 || cmd[0] != "omp" {
		t.Errorf("BuildCommand(new) = %v, want [omp]", cmd)
	}
}

func TestOmpTool_BuildCommand_Resume(t *testing.T) {
	omp := NewOmp()
	cmd := omp.BuildCommand("", true, "")
	expected := []string{"omp", "--continue"}
	if len(cmd) != len(expected) {
		t.Fatalf("BuildCommand(resume) = %v, want %v", cmd, expected)
	}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("BuildCommand(resume)[%d] = %q, want %q", i, cmd[i], v)
		}
	}
}

func TestOmpTool_BuildCommand_ResumeWithID(t *testing.T) {
	omp := NewOmp()
	// Even when a resumeSessionID is provided, omp always uses --continue
	// because it manages its own session discovery in its agent dir.
	cmd := omp.BuildCommand("", true, "some-id")
	expected := []string{"omp", "--continue"}
	if len(cmd) != len(expected) {
		t.Fatalf("BuildCommand(resume with ID) = %v, want %v", cmd, expected)
	}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("BuildCommand(resume with ID)[%d] = %q, want %q", i, cmd[i], v)
		}
	}
}

func TestOmpTool_DiscoverSessionID(t *testing.T) {
	omp := NewOmp()
	id := omp.DiscoverSessionID("/some/path")
	if id != "" {
		t.Errorf("DiscoverSessionID() = %q, want %q", id, "")
	}
}

func TestOmpTool_GetSandboxSettings(t *testing.T) {
	omp := NewOmp()
	settings := omp.GetSandboxSettings()

	// Without contextFilePath, settings should be empty
	if len(settings) != 0 {
		t.Errorf("GetSandboxSettings() = %v, want empty map", settings)
	}
}

func TestOmpTool_EssentialConfigFiles(t *testing.T) {
	omp := NewOmp()
	tcf, ok := omp.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("OmpTool does not implement ToolWithConfigDirFiles")
	}
	files := tcf.EssentialConfigFiles()
	// omp stores credentials in agent.db (SQLite), not auth.json — so we
	// do NOT include auth.json here. See catalog.toml comment.
	expected := []string{"settings.json", "models.json", "AGENTS.md"}
	if len(files) != len(expected) {
		t.Fatalf("EssentialConfigFiles() = %v, want %v", files, expected)
	}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("EssentialConfigFiles()[%d] = %q, want %q", i, f, expected[i])
		}
	}
}

func TestOmpTool_SandboxSettingsFileName(t *testing.T) {
	omp := NewOmp()
	tcf, ok := omp.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("OmpTool does not implement ToolWithConfigDirFiles")
	}
	if tcf.SandboxSettingsFileName() != "settings.json" {
		t.Errorf("SandboxSettingsFileName() = %q, want %q", tcf.SandboxSettingsFileName(), "settings.json")
	}
}

func TestOmpTool_StateConfigFileName(t *testing.T) {
	omp := NewOmp()
	tcf, ok := omp.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("OmpTool does not implement ToolWithConfigDirFiles")
	}
	if tcf.StateConfigFileName() != "" {
		t.Errorf("StateConfigFileName() = %q, want %q", tcf.StateConfigFileName(), "")
	}
}

func TestOmpTool_AlwaysSetupConfig(t *testing.T) {
	omp := NewOmp()
	tcf, ok := omp.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("OmpTool does not implement ToolWithConfigDirFiles")
	}
	if !tcf.AlwaysSetupConfig() {
		t.Error("AlwaysSetupConfig() = false, want true")
	}
}

func TestOmpTool_RegistryLookup(t *testing.T) {
	omp, err := Get("omp")
	if err != nil {
		t.Fatalf("Get(\"omp\") returned error: %v", err)
	}
	if omp.Name() != "omp" {
		t.Errorf("Name() = %q, want %q", omp.Name(), "omp")
	}
}

func TestOmpTool_ImplementsPermissionMode(t *testing.T) {
	omp := NewOmp()

	twpm, ok := omp.(ToolWithPermissionMode)
	if !ok {
		t.Fatal("OmpTool should implement ToolWithPermissionMode")
	}

	// Verify method works without panic
	twpm.SetPermissionMode("interactive")
}

func TestOmpTool_ImplementsContainerEnv(t *testing.T) {
	omp := NewOmp()

	twce, ok := omp.(ToolWithContainerEnv)
	if !ok {
		t.Fatal("OmpTool should implement ToolWithContainerEnv")
	}

	env := twce.GetContainerEnv("/workspace")
	// omp exposes two env vars with distinct semantics. We must set the
	// session-only one (PI_CODING_AGENT_SESSION_DIR); setting the agent-dir
	// one (PI_CODING_AGENT_DIR) would orphan every host-seeded config file.
	if env["PI_CODING_AGENT_SESSION_DIR"] != "/workspace/.omp-agent" {
		t.Errorf("PI_CODING_AGENT_SESSION_DIR = %q, want %q", env["PI_CODING_AGENT_SESSION_DIR"], "/workspace/.omp-agent")
	}
	if _, set := env["PI_CODING_AGENT_DIR"]; set {
		t.Errorf("env contains PI_CODING_AGENT_DIR; that var overrides the entire agent dir and would orphan host-seeded config")
	}
}

func TestOmpTool_GetContainerEnv_CustomWorkspace(t *testing.T) {
	omp := &OmpTool{}
	env := omp.GetContainerEnv("/home/user/project")
	if env["PI_CODING_AGENT_SESSION_DIR"] != "/home/user/project/.omp-agent" {
		t.Errorf("PI_CODING_AGENT_SESSION_DIR = %q, want %q", env["PI_CODING_AGENT_SESSION_DIR"], "/home/user/project/.omp-agent")
	}
}

func TestOmpTool_GetSandboxSettings_WithAutoContext(t *testing.T) {
	omp := &OmpTool{contextFilePath: "/home/code/SANDBOX_CONTEXT.md"}
	settings := omp.GetSandboxSettings()

	// Context injection is handled via PreLaunch (symlinks APPEND_SYSTEM.md), not settings.json
	if len(settings) != 0 {
		t.Errorf("GetSandboxSettings() = %v, want empty map", settings)
	}
}

func TestOmpTool_BuildCommand_WithContext(t *testing.T) {
	omp := &OmpTool{contextFilePath: "/home/code/SANDBOX_CONTEXT.md"}
	cmd := omp.BuildCommand("some-session-id", false, "")
	// BuildCommand should be clean — no shell setup, that's in PreLaunch
	expected := []string{"omp"}
	if len(cmd) != len(expected) {
		t.Fatalf("BuildCommand(with context) = %v, want %v", cmd, expected)
	}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("BuildCommand(with context)[%d] = %q, want %q", i, cmd[i], v)
		}
	}
}

func TestOmpTool_PreLaunch_WithContext(t *testing.T) {
	omp := &OmpTool{contextFilePath: "/home/code/SANDBOX_CONTEXT.md"}
	cmds := omp.PreLaunch()
	if len(cmds) != 2 {
		t.Fatalf("PreLaunch() returned %d commands, want 2", len(cmds))
	}
	// First command: mkdir -p
	expectedMkdir := []string{"mkdir", "-p", "/home/code/.omp/agent"}
	if len(cmds[0]) != len(expectedMkdir) {
		t.Fatalf("PreLaunch()[0] = %v, want %v", cmds[0], expectedMkdir)
	}
	for i, v := range expectedMkdir {
		if cmds[0][i] != v {
			t.Errorf("PreLaunch()[0][%d] = %q, want %q", i, cmds[0][i], v)
		}
	}
	// Second command: ln -sf
	expectedLn := []string{"ln", "-sf", "/home/code/SANDBOX_CONTEXT.md", "/home/code/.omp/agent/APPEND_SYSTEM.md"}
	if len(cmds[1]) != len(expectedLn) {
		t.Fatalf("PreLaunch()[1] = %v, want %v", cmds[1], expectedLn)
	}
	for i, v := range expectedLn {
		if cmds[1][i] != v {
			t.Errorf("PreLaunch()[1][%d] = %q, want %q", i, cmds[1][i], v)
		}
	}
}

func TestOmpTool_PreLaunch_WithoutContext(t *testing.T) {
	omp := &OmpTool{}
	cmds := omp.PreLaunch()
	if cmds != nil {
		t.Errorf("PreLaunch() = %v, want nil", cmds)
	}
}

func TestOmpTool_PreLaunch_SpecialChars(t *testing.T) {
	// Paths with spaces, quotes, etc. are passed as argv elements — no shell interpretation
	omp := &OmpTool{contextFilePath: "/home/code/path with spaces/SANDBOX_CONTEXT.md"}
	cmds := omp.PreLaunch()
	if len(cmds) != 2 {
		t.Fatalf("PreLaunch() returned %d commands, want 2", len(cmds))
	}
	// The path should appear verbatim, not quoted
	if cmds[1][2] != "/home/code/path with spaces/SANDBOX_CONTEXT.md" {
		t.Errorf("PreLaunch()[1][2] = %q, want path with spaces verbatim", cmds[1][2])
	}
}

func TestOmpTool_SetAutoContextPath_RejectsRelative(t *testing.T) {
	omp := &OmpTool{}
	// Verify interface implementation
	var _ ToolWithAutoContextPath = omp

	omp.SetAutoContextPath("relative/path/SANDBOX_CONTEXT.md")
	if omp.contextFilePath != "" {
		t.Errorf("SetAutoContextPath accepted relative path, contextFilePath = %q", omp.contextFilePath)
	}
}

func TestOmpTool_SetAutoContextPath_AcceptsAbsolute(t *testing.T) {
	omp := &OmpTool{}
	omp.SetAutoContextPath("/home/code/SANDBOX_CONTEXT.md")
	if omp.contextFilePath != "/home/code/SANDBOX_CONTEXT.md" {
		t.Errorf("SetAutoContextPath did not store path, contextFilePath = %q", omp.contextFilePath)
	}
}

func TestOmpTool_ImplementsPreLaunch(t *testing.T) {
	omp := NewOmp()
	_, ok := omp.(ToolWithPreLaunch)
	if !ok {
		t.Fatal("OmpTool should implement ToolWithPreLaunch")
	}
}

func TestOmpTool_DoesNotImplementAutoContextFile(t *testing.T) {
	omp := NewOmp()
	_, ok := omp.(ToolWithAutoContextFile)
	if ok {
		t.Error("OmpTool should NOT implement ToolWithAutoContextFile")
	}
}

func TestListSupported_IncludesOmp(t *testing.T) {
	supported := ListSupported()
	found := false
	for _, name := range supported {
		if name == "omp" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListSupported() = %v, does not include 'omp'", supported)
	}
}
