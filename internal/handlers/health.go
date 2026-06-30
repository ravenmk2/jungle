package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/common"
)

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, common.OK(map[string]string{"status": "ok"}))
}
