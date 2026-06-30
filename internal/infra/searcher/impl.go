// Package searcher is the ripgrep adapter (stub).
package searcher

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Search(ctx context.Context, roots []string, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "searcher not implemented")
}
