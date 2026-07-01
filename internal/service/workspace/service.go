// Package workspace is the workspace/profile usecase.
package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ravenmk2/jungle/internal/apperrors"
	"github.com/ravenmk2/jungle/internal/config"
	"github.com/ravenmk2/jungle/internal/storage"
)

// View is the workspace view returned by Get.
type View struct {
	Name           string `json:"name"`
	CurrentProfile string `json:"currentProfile"`
	*config.WorkspaceConfig
}

// Service is the workspace usecase Port.
type Service interface {
	List(ctx context.Context) ([]string, error)
	Get(ctx context.Context, workspace string) (*View, error)
	SwitchProfile(ctx context.Context, workspace, profile string) error
}

type service struct {
	configDir string
	dataDir   string
}

// New creates a workspace Service.
func New(configDir, dataDir string) Service {
	return &service{configDir: configDir, dataDir: dataDir}
}

func (s *service) wsPath(name string) string {
	return filepath.Join(s.configDir, "workspaces", name+".toml")
}

func (s *service) List(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.configDir, "workspaces"))
	if err != nil {
		return nil, apperrors.New(apperrors.NotFound, "no workspaces")
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if name := strings.TrimSuffix(e.Name(), ".toml"); name != e.Name() {
			out = append(out, name)
		}
	}
	return out, nil
}

func (s *service) Get(ctx context.Context, workspace string) (*View, error) {
	cfg, err := config.LoadWorkspace(s.wsPath(workspace))
	if err != nil {
		return nil, apperrors.New(apperrors.NotFound, "workspace not found: "+workspace)
	}
	v := &View{Name: workspace, WorkspaceConfig: cfg}
	p := storage.New(s.dataDir)
	if st, err := p.ReadState(workspace); err == nil {
		v.CurrentProfile = st.CurrentProfile
	}
	return v, nil
}

func (s *service) SwitchProfile(ctx context.Context, workspace, profile string) error {
	if _, err := config.LoadWorkspace(s.wsPath(workspace)); err != nil {
		return apperrors.New(apperrors.NotFound, "workspace not found: "+workspace)
	}
	p := storage.New(s.dataDir)
	return p.WriteState(workspace, &storage.WorkspaceState{CurrentProfile: profile})
}
