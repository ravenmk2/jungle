package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T, configDir, dataDir string) Service {
	t.Helper()
	wsDir := filepath.Join(configDir, "workspaces")
	os.MkdirAll(wsDir, 0o755)
	os.WriteFile(filepath.Join(wsDir, "demo.toml"), []byte(`[java]
version = 8
home = "/jdk"
[maven]
home = "/mvn"
repo = "/repo"
[profiles]
items = ["dev", "staging"]
`), 0o644)
	return New(configDir, dataDir)
}

func TestList(t *testing.T) {
	s := setup(t, t.TempDir(), t.TempDir())
	got, err := s.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "demo" {
		t.Fatalf("List = %+v", got)
	}
}

func TestGetAndSwitch(t *testing.T) {
	s := setup(t, t.TempDir(), t.TempDir())
	v, err := s.Get(nil, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if v.CurrentProfile != "" {
		t.Fatalf("initial profile = %q, want empty", v.CurrentProfile)
	}
	if err := s.SwitchProfile(nil, "demo", "dev"); err != nil {
		t.Fatal(err)
	}
	v, _ = s.Get(nil, "demo")
	if v.CurrentProfile != "dev" {
		t.Fatalf("after switch profile = %q", v.CurrentProfile)
	}
}

func TestListEmpty(t *testing.T) {
	configDir := t.TempDir()
	os.MkdirAll(filepath.Join(configDir, "workspaces"), 0o755)
	s := New(configDir, t.TempDir())
	got, err := s.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("List = nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("List = %+v, want empty", got)
	}
}
