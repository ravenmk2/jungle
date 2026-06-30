// Package database is the DB usecase (stub).
package database

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

type QueryOpts struct {
	MaxRows int // default 200 (spec 6.3/G7)
}

type QueryResult struct {
	Columns   []string        `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	Truncated bool            `json:"truncated"`
}

type Service interface {
	Reset(ctx context.Context, workspace, database string) error
	Query(ctx context.Context, workspace, database, sql string, opts QueryOpts) (*QueryResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Reset(ctx context.Context, workspace, database string) error {
	return apperrors.New(apperrors.DBResetFailed, "not implemented")
}
func (s *service) Query(ctx context.Context, workspace, database, sql string, opts QueryOpts) (*QueryResult, error) {
	return nil, apperrors.New(apperrors.DBQueryFailed, "not implemented")
}
