// Package build is the build usecase (stub).
package build

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type RunOpts struct {
	Goal  string
	Clean bool
	Test  bool
}

// RunResult is the build result (spec 6.1: build-id + status).
type RunResult struct {
	BuildID string `json:"buildId"`
	Status  string `json:"status"` // success | failed
}

type Service interface {
	Run(ctx context.Context, workspace, project string, opts RunOpts) (*RunResult, error)
	LogSearch(ctx context.Context, workspace, project, buildID, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Run(ctx context.Context, workspace, project string, opts RunOpts) (*RunResult, error) {
	return nil, apperrors.New(apperrors.BuildFailed, "not implemented")
}
func (s *service) LogSearch(ctx context.Context, workspace, project, buildID, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
