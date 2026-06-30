package storage

import (
	"path/filepath"
	"testing"
)

func TestPaths(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	if got := p.ProjectDir("ws", "demo"); got != filepath.Join(p.DataDir, "ws-ws", "projects", "demo") {
		t.Fatalf("ProjectDir = %q", got)
	}
	if got := p.BuildDir("ws", "demo", "0001"); got != filepath.Join(p.ProjectDir("ws", "demo"), "builds", "0001") {
		t.Fatalf("BuildDir = %q", got)
	}
	if got := p.RunDir("ws", "svc", "0002"); got != filepath.Join(p.ServiceDir("ws", "svc"), "runs", "0002") {
		t.Fatalf("RunDir = %q", got)
	}
	if got := p.BuildStateFile("ws", "demo", "0001"); got != filepath.Join(p.BuildDir("ws", "demo", "0001"), "state.json") {
		t.Fatalf("BuildStateFile = %q", got)
	}
	if got := p.RunStateFile("ws", "svc", "0002"); got != filepath.Join(p.RunDir("ws", "svc", "0002"), "state.json") {
		t.Fatalf("RunStateFile = %q", got)
	}
	if got := p.MavenSourceCacheDir("ws", "com.example", "foo", "1.0"); got != filepath.Join(p.WorkspaceDir("ws"), "maven-source-cache", "com/example", "foo", "1.0") {
		t.Fatalf("MavenSourceCacheDir = %q", got)
	}
}

func TestAllocBuildID(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	id1, err := p.AllocBuildID("ws", "demo")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := p.AllocBuildID("ws", "demo")
	if err != nil {
		t.Fatal(err)
	}
	if id1 != "0001" || id2 != "0002" {
		t.Fatalf("ids = %q, %q", id1, id2)
	}
}

func TestWorkspaceState(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	if _, err := p.ReadState("ws"); err == nil {
		t.Fatal("expected error reading missing state")
	}
	if err := p.WriteState("ws", &WorkspaceState{CurrentProfile: "dev"}); err != nil {
		t.Fatal(err)
	}
	got, err := p.ReadState("ws")
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrentProfile != "dev" {
		t.Fatalf("profile = %q", got.CurrentProfile)
	}
}

func TestBuildRunState(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	bs := &BuildState{Status: "running", Command: "mvn package"}
	if err := p.WriteBuildState("ws", "demo", "0001", bs); err != nil {
		t.Fatal(err)
	}
	got, err := p.ReadBuildState("ws", "demo", "0001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "running" || got.Command != "mvn package" {
		t.Fatalf("build state = %+v", got)
	}
	rs := &RunState{Status: "running", PID: 123, Port: 8080, Profile: "dev"}
	if err := p.WriteRunState("ws", "svc", "0001", rs); err != nil {
		t.Fatal(err)
	}
	gotr, err := p.ReadRunState("ws", "svc", "0001")
	if err != nil {
		t.Fatal(err)
	}
	if gotr.PID != 123 || gotr.Port != 8080 || gotr.Profile != "dev" {
		t.Fatalf("run state = %+v", gotr)
	}
}
