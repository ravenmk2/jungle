// Package process is the background-process adapter (stub).
package process

import "github.com/ravenmk2/jungle/internal/apperrors"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Start(javaHome, jar, workDir string, port int, profile string) (pid int, err error) {
	return 0, apperrors.New(apperrors.RunFailed, "process start not implemented")
}
func (a *Adapter) Stop(pid int) error {
	return apperrors.New(apperrors.ServiceNotRunning, "process stop not implemented")
}
