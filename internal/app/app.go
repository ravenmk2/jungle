// Package app is the composition root: it loads config, wires components,
// injects dependencies, and starts the HTTP server.
package app

import (
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/config"
	"github.com/ravenmk2/jungle/internal/handlers"
	"github.com/ravenmk2/jungle/internal/mcp"
	"github.com/ravenmk2/jungle/internal/service/workspace"
	"github.com/sirupsen/logrus"
)

// Run loads config.toml from configDir, applies dataDir/addr overrides,
// wires components, and serves HTTP. Empty dataDir/addr fall back to config.
func Run(configDir, dataDir, addr string) error {
	srvCfg, err := config.LoadServer(filepath.Join(configDir, "config.toml"))
	if err != nil {
		return err
	}
	if addr != "" {
		srvCfg.Server.Addr = addr
	}
	if dataDir != "" {
		srvCfg.Data.Dir = dataDir
	}
	if srvCfg.Data.Dir == "" {
		srvCfg.Data.Dir = "./data"
	}
	if lvl, err := logrus.ParseLevel(srvCfg.Log.Level); err == nil {
		logrus.SetLevel(lvl)
	}

	eng := echo.New()
	wsSvc := workspace.New(configDir, srvCfg.Data.Dir)
	handlers.New(eng, wsSvc)
	mcp.Mount(eng)

	logrus.Infof("jungle listening on %s", srvCfg.Server.Addr)
	return eng.Start(srvCfg.Server.Addr)
}
