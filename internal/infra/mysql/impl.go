// Package mysql is the MySQL adapter (stub).
package mysql

import "github.com/ravenmk2/jungle/internal/apperrors"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Reset(dsn, initSQL string) error {
	return apperrors.New(apperrors.DBResetFailed, "mysql reset not implemented")
}
func (a *Adapter) Query(dsn, sql string, maxRows int) (columns []string, rows [][]interface{}, truncated bool, err error) {
	return nil, nil, false, apperrors.New(apperrors.DBQueryFailed, "mysql query not implemented")
}
