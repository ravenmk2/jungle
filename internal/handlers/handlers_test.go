package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/common"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

func setupEng(t *testing.T) *echo.Echo {
	t.Helper()
	configDir := t.TempDir()
	wsDir := filepath.Join(configDir, "workspaces")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, "demo.toml"), []byte(`[java]
version = 8
home = "/jdk"
[profiles]
items = ["dev", "staging"]
[databases."db1"]
host = "127.0.0.1"
port = 3306
db = "demo"
user = "root"
password = "pw"
init-sql = "/init.sql"
[services."svc1"]
project = "demo"
module = "app"
work-dir = "/wd"
port = 8080
database = "db1"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	New(e, workspace.New(configDir, t.TempDir()))
	return e
}

func TestHealth(t *testing.T) {
	e := setupEng(t)
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var env common.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.Success {
		t.Fatalf("envelope = %+v", env)
	}
}

func TestWorkspaceList(t *testing.T) {
	e := setupEng(t)
	req := httptest.NewRequest("POST", "/api/workspace/list", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got struct {
		Success bool            `json:"success"`
		Data    []string        `json:"data"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if !got.Success || len(got.Data) != 1 || got.Data[0] != "demo" {
		t.Fatalf("list = %+v body=%s", got, rec.Body.String())
	}
}

func TestWorkspaceGetAndSwitch(t *testing.T) {
	e := setupEng(t)
	// get
	req := httptest.NewRequest("POST", "/api/workspace/get", strings.NewReader(`{"workspace":"demo"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d", rec.Code)
	}
	// switch-profile
	req = httptest.NewRequest("POST", "/api/workspace/switch-profile", strings.NewReader(`{"workspace":"demo","profile":"dev"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("switch status = %d", rec.Code)
	}
	var env common.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.Success {
		t.Fatalf("switch envelope = %+v", env)
	}
}

func TestWorkspaceGetValidation(t *testing.T) {
	e := setupEng(t)
	// missing workspace field -> VALIDATION_ERROR envelope (still HTTP 200)
	req := httptest.NewRequest("POST", "/api/workspace/get", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var env common.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %+v", env)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("business error status = %d, want 200", rec.Code)
	}
}

func TestNotFoundEnvelope(t *testing.T) {
	e := setupEng(t)
	req := httptest.NewRequest("GET", "/api/no-such-route", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var env common.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Success || env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND envelope, got %+v", env)
	}
}

func TestPanicEnvelope(t *testing.T) {
	e := setupEng(t)
	e.GET("/api/boom", func(c *echo.Context) error { panic("boom") })
	req := httptest.NewRequest("GET", "/api/boom", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var env common.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Success || env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR envelope, got %+v", env)
	}
}

func TestWorkspaceGetShape(t *testing.T) {
	e := setupEng(t)
	req := httptest.NewRequest("POST", "/api/workspace/get", strings.NewReader(`{"workspace":"demo"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "password") || strings.Contains(body, "Password") {
		t.Fatalf("password leaked: %s", body)
	}
	for _, want := range []string{`"currentProfile"`, `"initSql"`, `"workDir"`, `"databases"`, `"services"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
	for _, bad := range []string{`"InitSQL"`, `"WorkDir"`, `"CurrentProfile"`, `"Password"`, `"Databases"`} {
		if strings.Contains(body, bad) {
			t.Fatalf("PascalCase leak %q in %s", bad, body)
		}
	}
}
