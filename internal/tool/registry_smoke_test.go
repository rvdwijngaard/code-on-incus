package tool

import (
	"testing"
)

func TestRegistry_AllLoadedToolsHaveBasics(t *testing.T) {
	for _, name := range ListSupported() {
		t.Run(name, func(t *testing.T) {
			tool, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%q) failed: %v", name, err)
			}
			if tool.Name() != name {
				t.Errorf("Name() = %q, want %q", tool.Name(), name)
			}
			if tool.Binary() == "" {
				t.Errorf("Binary() = empty")
			}
			if tool.ConfigDirName() == "" {
				t.Errorf("ConfigDirName() = empty")
			}
		})
	}
}
