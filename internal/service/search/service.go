// Package search is the search usecase (stub).
package search

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type Service interface {
	Docs(ctx context.Context, workspace, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
	MavenSource(ctx context.Context, workspace, dependency, project, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
	MavenRead(ctx context.Context, workspace, dependency, project, file, class string, rng [2]int) (interface{}, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Docs(ctx context.Context, workspace, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
func (s *service) MavenSource(ctx context.Context, workspace, dependency, project, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
func (s *service) MavenRead(ctx context.Context, workspace, dependency, project, file, class string, rng [2]int) (interface{}, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
