// Package maven is the Maven adapter: build, dependency:list/sources, sources.jar fetch (stub).
package maven

import "github.com/ravenmk2/jungle/internal/apperrors"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Build(repoDir string, goal string, clean, test bool) (int, error) {
	return 0, apperrors.New(apperrors.BuildFailed, "maven build not implemented")
}
func (a *Adapter) ResolveRunnableJar(moduleDir string) (string, error) {
	return "", apperrors.New(apperrors.RunFailed, "maven resolve not implemented")
}
func (a *Adapter) FetchSourcesJar(coord string) (string, error) {
	return "", apperrors.New(apperrors.SearchFailed, "maven fetch not implemented")
}
func (a *Adapter) FetchSourcesBatch(repoDir string) error {
	return apperrors.New(apperrors.SearchFailed, "maven dependency:sources not implemented")
}
func (a *Adapter) ListDependencies(repoDir string) ([]string, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "maven dep:list not implemented")
}
