// Package mcp mounts the /mcp Streamable HTTP placeholder and defines the tool registry.
// Real MCP Go SDK wiring is deferred to the tool-wiring phase.
package mcp

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/apperrors"
	"github.com/ravenmk2/jungle/internal/common"
)

// Tool describes an MCP tool (name + handler to be wired later).
type Tool struct {
	Name string
}

// Registry holds registered tools.
type Registry struct{ tools map[string]Tool }

func NewRegistry() *Registry { return &Registry{tools: map[string]Tool{}} }

func (r *Registry) Register(t Tool) { r.tools[t.Name] = t }
func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// Mount attaches the /mcp placeholder route.
func Mount(eng *echo.Echo) {
	eng.Any("/mcp", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, common.Fail(apperrors.New(apperrors.InternalError, "mcp not implemented")))
	})
}
