package tool

import (
	"testing"
)

func TestPiTool_Basics(t *testing.T) {
	pi := NewPi()

	if pi.Name() != "pi" {
		t.Errorf("Name() = %q, want %q", pi.Name(), "pi")
	}
	if pi.Binary() != "pi" {
		t.Errorf("Binary() = %q, want %q", pi.Binary(), "pi")
	}
	if pi.ConfigDirName() != ".pi/agent" {
		t.Errorf("ConfigDirName() = %q, want %q", pi.ConfigDirName(), ".pi/agent")
	}
	if pi.SessionsDirName() != "sessions-pi" {
		t.Errorf("SessionsDirName() = %q, want %q", pi.SessionsDirName(), "sessions-pi")
	}
}

func TestPiTool_BuildCommand_NewSession(t *testing.T) {
	pi := NewPi()
	cmd := pi.BuildCommand("some-session-id", false, "")
	if len(cmd) != 1 || cmd[0] != "pi" {
		t.Errorf("BuildCommand(new) = %v, want [pi]", cmd)
	}
}

func TestPiTool_BuildCommand_Resume(t *testing.T) {
	pi := NewPi()
	cmd := pi.BuildCommand("", true, "")
	expected := []string{"pi", "--continue"}
	if len(cmd) != len(expected) {
		t.Fatalf("BuildCommand(resume) = %v, want %v", cmd, expected)
	}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("BuildCommand(resume)[%d] = %q, want %q", i, cmd[i], v)
		}
	}
}

func TestPiTool_BuildCommand_ResumeWithID(t *testing.T) {
	pi := NewPi()
	cmd := pi.BuildCommand("", true, "some-id")
	expected := []string{"pi", "--session", "some-id"}
	if len(cmd) != len(expected) {
		t.Fatalf("BuildCommand(resume with ID) = %v, want %v", cmd, expected)
	}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("BuildCommand(resume with ID)[%d] = %q, want %q", i, cmd[i], v)
		}
	}
}

func TestPiTool_DiscoverSessionID(t *testing.T) {
	pi := NewPi()
	id := pi.DiscoverSessionID("/some/path")
	if id != "" {
		t.Errorf("DiscoverSessionID() = %q, want %q", id, "")
	}
}

func TestPiTool_GetSandboxSettings(t *testing.T) {
	pi := NewPi()
	settings := pi.GetSandboxSettings()

	// Without contextFilePath, settings should be empty
	if len(settings) != 0 {
		t.Errorf("GetSandboxSettings() = %v, want empty map", settings)
	}
}

func TestPiTool_EssentialConfigFiles(t *testing.T) {
	pi := NewPi()
	tcf, ok := pi.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("PiTool does not implement ToolWithConfigDirFiles")
	}
	files := tcf.EssentialConfigFiles()
	expected := []string{"settings.json", "models.json", "auth.json"}
	if len(files) != len(expected) {
		t.Fatalf("EssentialConfigFiles() = %v, want %v", files, expected)
	}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("EssentialConfigFiles()[%d] = %q, want %q", i, f, expected[i])
		}
	}
}

func TestPiTool_SandboxSettingsFileName(t *testing.T) {
	pi := NewPi()
	tcf, ok := pi.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("PiTool does not implement ToolWithConfigDirFiles")
	}
	if tcf.SandboxSettingsFileName() != "settings.json" {
		t.Errorf("SandboxSettingsFileName() = %q, want %q", tcf.SandboxSettingsFileName(), "settings.json")
	}
}

func TestPiTool_StateConfigFileName(t *testing.T) {
	pi := NewPi()
	tcf, ok := pi.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("PiTool does not implement ToolWithConfigDirFiles")
	}
	if tcf.StateConfigFileName() != "" {
		t.Errorf("StateConfigFileName() = %q, want %q", tcf.StateConfigFileName(), "")
	}
}

func TestPiTool_AlwaysSetupConfig(t *testing.T) {
	pi := NewPi()
	tcf, ok := pi.(ToolWithConfigDirFiles)
	if !ok {
		t.Fatal("PiTool does not implement ToolWithConfigDirFiles")
	}
	if !tcf.AlwaysSetupConfig() {
		t.Error("AlwaysSetupConfig() = false, want true")
	}
}

func TestPiTool_RegistryLookup(t *testing.T) {
	pi, err := Get("pi")
	if err != nil {
		t.Fatalf("Get(\"pi\") returned error: %v", err)
	}
	if pi.Name() != "pi" {
		t.Errorf("Name() = %q, want %q", pi.Name(), "pi")
	}
}

func TestPiTool_ImplementsPermissionMode(t *testing.T) {
	pi := NewPi()

	twpm, ok := pi.(ToolWithPermissionMode)
	if !ok {
		t.Fatal("PiTool should implement ToolWithPermissionMode")
	}

	// Verify method works without panic
	twpm.SetPermissionMode("interactive")
}

func TestPiTool_ImplementsContainerEnv(t *testing.T) {
	pi := NewPi()

	twce, ok := pi.(ToolWithContainerEnv)
	if !ok {
		t.Fatal("PiTool should implement ToolWithContainerEnv")
	}

	env := twce.GetContainerEnv("/workspace")
	if env["PI_CODING_AGENT_SESSION_DIR"] != "/workspace/.pi-sessions" {
		t.Errorf("PI_CODING_AGENT_SESSION_DIR = %q, want %q", env["PI_CODING_AGENT_SESSION_DIR"], "/workspace/.pi-sessions")
	}
}

func TestPiTool_GetContainerEnv_CustomWorkspace(t *testing.T) {
	pi := &PiTool{}
	env := pi.GetContainerEnv("/home/user/project")
	if env["PI_CODING_AGENT_SESSION_DIR"] != "/home/user/project/.pi-sessions" {
		t.Errorf("PI_CODING_AGENT_SESSION_DIR = %q, want %q", env["PI_CODING_AGENT_SESSION_DIR"], "/home/user/project/.pi-sessions")
	}
}

func TestPiTool_SetAutoContextPath(t *testing.T) {
	pi := NewPi()

	acp, ok := pi.(ToolWithAutoContextPath)
	if !ok {
		t.Fatal("PiTool should implement ToolWithAutoContextPath")
	}

	// Verify method works without panic
	acp.SetAutoContextPath("/home/code/SANDBOX_CONTEXT.md")
}

func TestPiTool_GetSandboxSettings_WithAutoContext(t *testing.T) {
	pi := &PiTool{contextFilePath: "/home/code/SANDBOX_CONTEXT.md"}
	settings := pi.GetSandboxSettings()

	// Context injection is handled via AutoContextFile (AGENTS.md), not settings.json
	// GetSandboxSettings returns empty map for pi
	if len(settings) != 0 {
		t.Errorf("GetSandboxSettings() = %v, want empty map", settings)
	}
}

func TestPiTool_GetSandboxSettings_WithoutAutoContext(t *testing.T) {
	pi := &PiTool{} // No contextFilePath set
	settings := pi.GetSandboxSettings()

	// Should NOT have appendSystemPrompt field
	if _, ok := settings["appendSystemPrompt"]; ok {
		t.Error("GetSandboxSettings() should not have 'appendSystemPrompt' key when contextFilePath is empty")
	}
}

func TestPiTool_AutoContextFile(t *testing.T) {
	pi := NewPi()
	acf, ok := pi.(ToolWithAutoContextFile)
	if !ok {
		t.Fatal("PiTool should implement ToolWithAutoContextFile")
	}

	path := acf.AutoContextFile()
	if path != ".pi/agent/AGENTS.md" {
		t.Errorf("AutoContextFile() = %q, want %q", path, ".pi/agent/AGENTS.md")
	}
}

func TestListSupported_IncludesPi(t *testing.T) {
	supported := ListSupported()
	found := false
	for _, name := range supported {
		if name == "pi" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListSupported() = %v, does not include 'pi'", supported)
	}
}
