package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/common"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

func mountWorkspace(eng *echo.Echo, wsSvc workspace.Service) {
	g := eng.Group("/api/workspace")
	g.POST("/list", func(c *echo.Context) error {
		out, err := wsSvc.List(c.Request().Context())
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, common.OK(out))
	})
	g.POST("/get", func(c *echo.Context) error {
		var req struct {
			Workspace string `json:"workspace" validate:"required"`
		}
		if err := common.Bind(c, &req); err != nil {
			return err
		}
		out, err := wsSvc.Get(c.Request().Context(), req.Workspace)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, common.OK(out))
	})
	g.POST("/switch-profile", func(c *echo.Context) error {
		var req struct {
			Workspace string `json:"workspace" validate:"required"`
			Profile   string `json:"profile" validate:"required"`
		}
		if err := common.Bind(c, &req); err != nil {
			return err
		}
		if err := wsSvc.SwitchProfile(c.Request().Context(), req.Workspace, req.Profile); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, common.OK(map[string]string{"workspace": req.Workspace, "currentProfile": req.Profile}))
	})
}
