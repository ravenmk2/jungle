// Package run is the service-run usecase (stub).
package run

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type StartResult struct {
	RunID string `json:"runId"`
	Port  int    `json:"port"`
}

type Service interface {
	Start(ctx context.Context, workspace, service string) (*StartResult, error)
	Stop(ctx context.Context, workspace, service string) error
	Status(ctx context.Context, workspace, service string) (interface{}, error)
	LogSearch(ctx context.Context, workspace, service, runID, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Start(ctx context.Context, workspace, service string) (*StartResult, error) {
	return nil, apperrors.New(apperrors.RunFailed, "not implemented")
}
func (s *service) Stop(ctx context.Context, workspace, service string) error {
	return apperrors.New(apperrors.ServiceNotRunning, "not implemented")
}
func (s *service) Status(ctx context.Context, workspace, service string) (interface{}, error) {
	return nil, apperrors.New(apperrors.ServiceNotRunning, "not implemented")
}
func (s *service) LogSearch(ctx context.Context, workspace, service, runID, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
