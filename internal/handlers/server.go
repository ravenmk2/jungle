// Package handlers mounts /api/* RPC routes onto an Echo engine.
package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

// New mounts all /api routes onto eng.
func New(eng *echo.Echo, wsSvc workspace.Service) {
	eng.GET("/api/health", health)
	mountWorkspace(eng, wsSvc)
}
