// Package handlers mounts /api/* RPC routes onto an Echo engine.
package handlers

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

// New mounts all /api routes onto eng and installs the unified error handler
// and panic-recovery middleware so every error (business, transport, panic)
// is rendered as a response envelope.
func New(eng *echo.Echo, wsSvc workspace.Service) {
	eng.Use(middleware.Recover())
	eng.HTTPErrorHandler = ErrorHandler
	eng.GET("/api/health", health)
	mountWorkspace(eng, wsSvc)
}
