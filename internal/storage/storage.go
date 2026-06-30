// Package storage encapsulates data-dir path conventions, id allocation, and state I/O.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

// Paths resolves all on-disk paths under a data root.
type Paths struct {
	DataDir string
}

// New creates a Paths rooted at dataDir.
func New(dataDir string) *Paths { return &Paths{DataDir: dataDir} }

func (p *Paths) WorkspaceDir(ws string) string {
	return filepath.Join(p.DataDir, "ws-"+ws)
}
func (p *Paths) ProjectDir(ws, project string) string {
	return filepath.Join(p.WorkspaceDir(ws), "projects", project)
}
func (p *Paths) ServiceDir(ws, service string) string {
	return filepath.Join(p.WorkspaceDir(ws), "services", service)
}
func (p *Paths) BuildDir(ws, project, buildID string) string {
	return filepath.Join(p.ProjectDir(ws, project), "builds", buildID)
}
func (p *Paths) RunDir(ws, service, runID string) string {
	return filepath.Join(p.ServiceDir(ws, service), "runs", runID)
}
func (p *Paths) StateFile(ws string) string {
	return filepath.Join(p.WorkspaceDir(ws), "state.json")
}
func (p *Paths) BuildStateFile(ws, project, buildID string) string {
	return filepath.Join(p.BuildDir(ws, project, buildID), "state.json")
}
func (p *Paths) RunStateFile(ws, service, runID string) string {
	return filepath.Join(p.RunDir(ws, service, runID), "state.json")
}

// MavenSourceCacheDir returns the extraction cache dir for a Maven coordinate.
// <g> is the groupId with dots converted to path separators.
func (p *Paths) MavenSourceCacheDir(ws, g, a, v string) string {
	gPath := filepath.FromSlash(strings.ReplaceAll(g, ".", "/"))
	return filepath.Join(p.WorkspaceDir(ws), "maven-source-cache", gPath, a, v)
}

// WorkspaceState is the runtime state (state.json).
type WorkspaceState struct {
	CurrentProfile string `json:"current-profile"`
}

// BuildState is the per-build state (builds/<id>/state.json).
type BuildState struct {
	Status    string `json:"status"`    // running | success | failed
	StartedAt string `json:"startedAt"` // RFC3339
	EndedAt   string `json:"endedAt"`   // RFC3339
	ExitCode  int    `json:"exitCode"`
	Command   string `json:"command"`
}

// RunState is the per-run state (runs/<id>/state.json).
type RunState struct {
	Status    string `json:"status"` // running | stopped | exited | crashed
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Profile   string `json:"profile"`
	StartedAt string `json:"startedAt"` // RFC3339
	EndedAt   string `json:"endedAt"`   // RFC3339
	ExitCode  int    `json:"exitCode"`
}

// ReadState reads state.json for a workspace.
func (p *Paths) ReadState(ws string) (*WorkspaceState, error) {
	b, err := os.ReadFile(p.StateFile(ws))
	if err != nil {
		return nil, err
	}
	var s WorkspaceState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteState writes state.json for a workspace.
func (p *Paths) WriteState(ws string, s *WorkspaceState) error {
	return writeJSON(p.StateFile(ws), p.WorkspaceDir(ws), s)
}

// ReadBuildState reads a build's state.json.
func (p *Paths) ReadBuildState(ws, project, buildID string) (*BuildState, error) {
	b, err := os.ReadFile(p.BuildStateFile(ws, project, buildID))
	if err != nil {
		return nil, err
	}
	var s BuildState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteBuildState writes a build's state.json.
func (p *Paths) WriteBuildState(ws, project, buildID string, s *BuildState) error {
	return writeJSON(p.BuildStateFile(ws, project, buildID), p.BuildDir(ws, project, buildID), s)
}

// ReadRunState reads a run's state.json.
func (p *Paths) ReadRunState(ws, service, runID string) (*RunState, error) {
	b, err := os.ReadFile(p.RunStateFile(ws, service, runID))
	if err != nil {
		return nil, err
	}
	var s RunState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteRunState writes a run's state.json.
func (p *Paths) WriteRunState(ws, service, runID string, s *RunState) error {
	return writeJSON(p.RunStateFile(ws, service, runID), p.RunDir(ws, service, runID), s)
}

func writeJSON(path, dir string, v interface{}) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// AllocBuildID allocates the next 4-digit build-id for a project.
func (p *Paths) AllocBuildID(ws, project string) (string, error) {
	return p.allocID(p.ProjectDir(ws, project), "build-id")
}

// AllocRunID allocates the next 4-digit run-id for a service.
func (p *Paths) AllocRunID(ws, service string) (string, error) {
	return p.allocID(p.ServiceDir(ws, service), "run-id")
}

func (p *Paths) allocID(dir, name string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	idFile := filepath.Join(dir, name)
	lockFile := idFile + ".lock"
	if err := lockWait(lockFile); err != nil {
		return "", err
	}
	defer os.Remove(lockFile)

	b, _ := os.ReadFile(idFile)
	n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	n++
	if err := os.WriteFile(idFile, []byte(strconv.Itoa(n)), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n), nil
}

// lockWait acquires an exclusive lock file (portable: O_CREATE|O_EXCL with retry).
// NOTE: a crashed process may leave a stale lock; stale-detection by mtime is a future refinement.
func lockWait(path string) error {
	for i := 0; i < 100; i++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return apperrors.New(apperrors.InternalError, "id lock: "+err.Error())
		}
		time.Sleep(10 * time.Millisecond)
	}
	return apperrors.New(apperrors.InternalError, "id lock timeout")
}
