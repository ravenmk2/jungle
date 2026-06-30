---
date: 2026-06-30
---

# Jungle 基础结构脚手架 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use wf-executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 Jungle 项目的可构建、可运行骨架：完整目录结构 + 核心共享层（config/storage/common/apperrors）+ workspace 最小可用端点 + 其余模块 stub + 组装根与入口，满足 `go build ./...` 通过、`/api/health` 与 `/api/workspace/*` 可调用。

**Architecture:** 端口/适配器分层。`internal/service`（usecase）不依赖 transport/infra；`internal/app` 为组装根负责组件初始化与静态 DI；`internal/handlers` 与 `internal/mcp` 为薄传输适配器；`internal/infra` 提供外部适配器 stub。HTTP 为 RPC-style（POST-only、统一信封）。

**Tech Stack:** Go 1.26、Echo v4、pelletier/go-toml/v2、go-playground/validator/v10、logrus。

## Global Constraints

- Module path: `github.com/ravenmk2/jungle`
- Go 版本: 1.26
- 配置格式: 人工编辑 = TOML（`config/config.toml` + `config/workspaces/*.toml`），程序写入 = JSON（`state.json`）
- HTTP API 风格: POST-only（`/api/health` 例外用 GET）、统一信封 `{success,data,error}`、kebab-case 路径、业务结果统一 200
- 除 `/api/health` 与 `/api/workspace/list` 外，所有端点请求体必须带 `workspace` 字段
- 错误码 `UPPER_SNAKE_CASE`
- 依赖铁律: `service` 不 import `handlers`/`mcp`/`infra`；`app` 是唯一组装根；`apperrors` 为各层可 import 的叶子包
- 数据目录: `./data/ws-{workspace}/...`，builds/runs 复数目录名，4 位 id 从 0001 起
- `.gitignore`: 忽略 `config/config.toml`、`config/workspaces/`、`data/`；保留 `config/config-sample.toml`
- 不在本计划范围: build/run/database/search 的完整实现、MCP Go SDK 接线（/mcp 仅占位路由 + Tool/Registry 接口）
- 提交: 每个 Task 末的 commit 步骤按 git-commit 规范执行（提交消息需经用户确认）

## File Structure

| 文件 | 职责 |
| --- | --- |
| `go.mod` / `go.sum` | 模块与依赖 |
| `.gitignore` | 忽略运行时配置与数据 |
| `config/config-sample.toml` | 服务级配置模板（tracked） |
| `Makefile` | build/run/test/vet/fmt |
| `.golangci.yml` | lint 配置 |
| `README.md` | 项目说明 |
| `AGENTS.md` | lint/test 命令 |
| `cmd/jungle/main.go` | 入口：解析 flag，调用 app |
| `internal/app/app.go` | 组装根：配置加载 + DI + HTTP Server 初始化 |
| `internal/apperrors/errors.go` | 类型化错误与错误码 |
| `internal/config/config.go` | TOML structs + loader（server + workspace） |
| `internal/storage/storage.go` | 路径助手 + id 分配器 + state.json 读写 |
| `internal/common/envelope.go` | 响应信封 |
| `internal/common/bind.go` | 请求解码 + 校验 + workspace 解析 |
| `internal/common/error_map.go` | error → 信封映射 |
| `internal/service/common/ports.go` | 跨切面 Port 接口（Searcher 等） |
| `internal/service/workspace/service.go` | workspace usecase（最小实现） |
| `internal/service/{build,run,database,search}/*.go` | usecase stub（ErrNotImplemented） |
| `internal/infra/{maven,mysql,process,searcher}/*.go` | 适配器 stub |
| `internal/handlers/server.go` | Echo 启动 + 路由挂载 |
| `internal/handlers/health.go` | `/api/health` |
| `internal/handlers/workspace.go` | `/api/workspace/*` |
| `internal/mcp/mcp.go` | `/mcp` 占位路由 + Tool/Registry 接口 |
| `internal/*/doc.go` | 各包职责说明占位 |

---

### Task 1: 项目初始化与目录骨架

**Files:**
- Create: `go.mod`, `.gitignore`, `config/config-sample.toml`, `Makefile`, `.golangci.yml`
- Create: `cmd/jungle/doc.go` 及 `internal/{app,apperrors,config,storage,common,handlers,mcp,infra,infra/maven,infra/mysql,infra/process,infra/searcher,service,service/common,service/build,service/run,service/database,service/search,service/workspace}/doc.go`
- Modify: `README.md`, `AGENTS.md`

**Interfaces:**
- Produces: 可被 `go build ./...` 编译的空骨架。

- [ ] **Step 1: 初始化 module**

Run:
```bash
go mod init github.com/ravenmk2/jungle
go get github.com/labstack/echo/v4@latest
go get github.com/pelletier/go-toml/v2@latest
go get github.com/go-playground/validator/v10@latest
go get github.com/sirupsen/logrus@latest
```
Expected: `go.mod` 生成，含 4 个 require。

- [ ] **Step 2: 写 `.gitignore`**

```gitignore
# 运行时配置与数据
config/config.toml
config/workspaces/
data/
```

- [ ] **Step 3: 写 `config/config-sample.toml`**

```toml
[server]
addr = "127.0.0.1:7788"

[data]
dir = "./data"

[log]
level = "info"
```

- [ ] **Step 4: 写 `Makefile`**

```makefile
.PHONY: build run test vet fmt lint

build:
	go build ./...

run:
	go run ./cmd/jungle

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run
```

- [ ] **Step 5: 写 `.golangci.yml`**

```yaml
run:
  go: "1.26"
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
```

- [ ] **Step 6: 各包 `doc.go` 占位**

为以下每个包各写一个 `doc.go`。**关键：`package` 名必须等于末级目录名**（照抄示例的 `package app` 到 apperrors 会编译失败）。示例：

`internal/app/doc.go`:
```go
// Package app is the composition root: component initialization, static DI, HTTP server init.
package app
```

`internal/apperrors/doc.go`:
```go
// Package apperrors provides typed errors and stable error codes.
package apperrors
```

`internal/storage/doc.go`:
```go
// Package storage encapsulates data-dir path conventions, id allocation, and state I/O.
package storage
```

为以下每个包各写一个 `doc.go`（包名 = 末级目录名）：`internal/app`、`internal/apperrors`、`internal/config`、`internal/storage`、`internal/common`、`internal/handlers`、`internal/mcp`、`internal/infra/maven`、`internal/infra/mysql`、`internal/infra/process`、`internal/infra/searcher`、`internal/service/common`、`internal/service/build`、`internal/service/run`、`internal/service/database`、`internal/service/search`、`internal/service/workspace`。

> `internal/infra` 与 `internal/service` 本身不建包（仅作父目录，子包各自成包）。`cmd/jungle` 在 Task 10 写 `main.go`（`package main`），此处不建 doc.go。

- [ ] **Step 7: 更新 `README.md`**

```markdown
# Jungle

为 AI Coding 设计的 Java 服务辅助开发环境。提供数据库操作、项目构建、服务运行、文档与 Maven 源码检索，支持 HTTP API 与 MCP 接入。

## 开发

```bash
make build     # 编译
make run       # 运行（默认读 ./config/config.toml）
make test      # 测试
make vet       # go vet
make lint      # golangci-lint
```

配置：复制 `config/config-sample.toml` 为 `config/config.toml` 并按需修改；workspace 配置置于 `config/workspaces/<name>.toml`。

详见 `docs/specs/0001-jungle-foundation.md`。
```

- [ ] **Step 8: 更新 `AGENTS.md`**

```markdown
# AGENTS.md

## 常用命令

- 构建: `make build`
- 运行: `make run`
- 测试: `make test`
- 静态检查: `make vet`
- Lint: `make lint`
- 格式化: `make fmt`
```

- [ ] **Step 9: 验证编译并提交**

Run:
```bash
go build ./...
```
Expected: 无输出（成功）。

```bash
git add -A
git commit -m "chore: 初始化项目骨架与目录结构"
```

---

### Task 2: apperrors（类型化错误与错误码）

**Files:**
- Create: `internal/apperrors/errors.go`
- Test: `internal/apperrors/errors_test.go`

**Interfaces:**
- Produces: `apperrors.Code`、`apperrors.Error`、`apperrors.Detail`、`apperrors.New(code,msg)`、`apperrors.NewWithDetails(code,msg,details)`、全部错误码常量。

- [ ] **Step 1: 写失败测试**

```go
package apperrors

import "testing"

func TestNewAndError(t *testing.T) {
	e := New(NotFound, "workspace not found")
	if e.Code != NotFound {
		t.Fatalf("code = %s, want %s", e.Code, NotFound)
	}
	if e.Error() != "workspace not found" {
		t.Fatalf("Error() = %q", e.Error())
	}
}

func TestNewWithDetails(t *testing.T) {
	d := []Detail{{Code: "REQUIRED", Message: "workspace is required", Target: "workspace"}}
	e := NewWithDetails(ValidationError, "validation failed", d)
	if len(e.Details) != 1 || e.Details[0].Target != "workspace" {
		t.Fatalf("details = %+v", e.Details)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/apperrors/...`
Expected: FAIL（`New`/`NotFound` 等未定义）。

- [ ] **Step 3: 写实现**

```go
// Package apperrors provides typed errors and stable error codes.
package apperrors

// Code is a stable, UPPER_SNAKE error code.
type Code string

const (
	ValidationError   Code = "VALIDATION_ERROR"
	NotFound          Code = "NOT_FOUND"
	InternalError     Code = "INTERNAL_ERROR"
	BuildFailed       Code = "BUILD_FAILED"
	RunFailed         Code = "RUN_FAILED"
	ServiceNotRunning Code = "SERVICE_NOT_RUNNING"
	DBQueryFailed     Code = "DB_QUERY_FAILED"
	DBResetFailed     Code = "DB_RESET_FAILED"
	SearchFailed      Code = "SEARCH_FAILED"
	AmbiguousTarget   Code = "AMBIGUOUS_TARGET"
)

// Detail is a field-level error.
type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

// Error is a typed error carrying a stable code and optional details.
type Error struct {
	Code    Code
	Message string
	Details []Detail
}

func (e *Error) Error() string { return e.Message }

// New creates a typed error.
func New(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

// NewWithDetails creates a typed error with field-level details.
func NewWithDetails(code Code, msg string, details []Detail) *Error {
	return &Error{Code: code, Message: msg, Details: details}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/apperrors/...`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/apperrors
git commit -m "feat(apperrors): 添加类型化错误与错误码"
```

---

### Task 3: config（TOML 加载）

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces: `config.ServerConfig`、`config.WorkspaceConfig` 及子结构、`config.LoadServer(path)`、`config.LoadWorkspace(path)`。

- [ ] **Step 1: 写失败测试**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServer(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(p, []byte(`[server]
addr = "127.0.0.1:7788"
[data]
dir = "./data"
[log]
level = "info"
`), 0o644)
	cfg, err := LoadServer(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != "127.0.0.1:7788" {
		t.Fatalf("addr = %q", cfg.Server.Addr)
	}
	if cfg.Data.Dir != "./data" {
		t.Fatalf("data.dir = %q", cfg.Data.Dir)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("log.level = %q", cfg.Log.Level)
	}
}

func TestLoadWorkspace(t *testing.T) {
	p := filepath.Join(t.TempDir(), "demo.toml")
	os.WriteFile(p, []byte(`[java]
version = 8
home = "/jdk"
[maven]
home = "/mvn"
repo = "/repo"
[docs]
dirs = ["a", "b"]
[projects."demo"]
repo = "/proj"
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
[profiles]
items = ["dev", "staging"]
`), 0o644)
	cfg, err := LoadWorkspace(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Java.Version != 8 || cfg.Java.Home != "/jdk" {
		t.Fatalf("java = %+v", cfg.Java)
	}
	if cfg.Maven.Repo != "/repo" {
		t.Fatalf("maven.repo = %q", cfg.Maven.Repo)
	}
	if len(cfg.Docs.Dirs) != 2 || cfg.Docs.Dirs[1] != "b" {
		t.Fatalf("docs.dirs = %+v", cfg.Docs.Dirs)
	}
	if cfg.Projects["demo"].Repo != "/proj" {
		t.Fatalf("projects = %+v", cfg.Projects)
	}
	if cfg.Databases["db1"].InitSQL != "/init.sql" {
		t.Fatalf("db.init-sql = %q", cfg.Databases["db1"].InitSQL)
	}
	if cfg.Services["svc1"].Database != "db1" || cfg.Services["svc1"].WorkDir != "/wd" {
		t.Fatalf("service = %+v", cfg.Services["svc1"])
	}
	if len(cfg.Profiles.Items) != 2 {
		t.Fatalf("profiles = %+v", cfg.Profiles.Items)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/config/...`
Expected: FAIL（类型/函数未定义）。

- [ ] **Step 3: 写实现**

```go
// Package config loads TOML configuration (server-level and per-workspace).
package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// ServerConfig is the jungle binary's own config (config.toml).
type ServerConfig struct {
	Server ServerSection `toml:"server"`
	Data   DataSection   `toml:"data"`
	Log    LogSection    `toml:"log"`
}

type ServerSection struct {
	Addr string `toml:"addr"`
}

type DataSection struct {
	Dir string `toml:"dir"`
}

type LogSection struct {
	Level string `toml:"level"`
}

// WorkspaceConfig is a per-workspace config (config/workspaces/<name>.toml).
type WorkspaceConfig struct {
	Java      JavaSection         `toml:"java"`
	Maven     MavenSection        `toml:"maven"`
	Docs      DocsSection         `toml:"docs"`
	Projects  map[string]Project  `toml:"projects"`
	Databases map[string]Database `toml:"databases"`
	Services  map[string]Service  `toml:"services"`
	Profiles  ProfilesSection     `toml:"profiles"`
}

type JavaSection struct {
	Version int    `toml:"version"`
	Home    string `toml:"home"`
}

type MavenSection struct {
	Home string `toml:"home"`
	Repo string `toml:"repo"`
}

type DocsSection struct {
	Dirs []string `toml:"dirs"`
}

type Project struct {
	Repo string `toml:"repo"`
}

type Database struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DB       string `toml:"db"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	InitSQL  string `toml:"init-sql"`
}

type Service struct {
	Project  string `toml:"project"`
	Module   string `toml:"module"`
	WorkDir  string `toml:"work-dir"`
	Port     int    `toml:"port"`
	Database string `toml:"database"`
}

type ProfilesSection struct {
	Items []string `toml:"items"`
}

// LoadServer reads and parses config.toml.
func LoadServer(path string) (*ServerConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadWorkspace reads and parses a workspace toml.
func LoadWorkspace(path string) (*WorkspaceConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg WorkspaceConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/config/...`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/config
git commit -m "feat(config): 添加 server/workspace TOML 加载"
```

---

### Task 4: storage（路径 + id 分配 + state 读写）

**Files:**
- Create: `internal/storage/storage.go`
- Test: `internal/storage/storage_test.go`

**Interfaces:**
- Produces: `storage.Paths`、`storage.WorkspaceState`、`storage.BuildState`、`storage.RunState`、`storage.New(dataDir)`、`Paths.ProjectDir/ServiceDir/BuildDir/RunDir/StateFile/BuildStateFile/RunStateFile/MavenSourceCacheDir`、`Paths.AllocBuildID/AllocRunID`、`Paths.ReadState/WriteState/ReadBuildState/WriteBuildState/ReadRunState/WriteRunState`。

- [ ] **Step 1: 写失败测试**

```go
package storage

import (
	"path/filepath"
	"testing"
)

func TestPaths(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	if got := p.ProjectDir("ws", "demo"); got != filepath.Join(p.DataDir, "ws-ws", "projects", "demo") {
		t.Fatalf("ProjectDir = %q", got)
	}
	if got := p.BuildDir("ws", "demo", "0001"); got != filepath.Join(p.ProjectDir("ws", "demo"), "builds", "0001") {
		t.Fatalf("BuildDir = %q", got)
	}
	if got := p.RunDir("ws", "svc", "0002"); got != filepath.Join(p.ServiceDir("ws", "svc"), "runs", "0002") {
		t.Fatalf("RunDir = %q", got)
	}
	if got := p.BuildStateFile("ws", "demo", "0001"); got != filepath.Join(p.BuildDir("ws", "demo", "0001"), "state.json") {
		t.Fatalf("BuildStateFile = %q", got)
	}
	if got := p.RunStateFile("ws", "svc", "0002"); got != filepath.Join(p.RunDir("ws", "svc", "0002"), "state.json") {
		t.Fatalf("RunStateFile = %q", got)
	}
	if got := p.MavenSourceCacheDir("ws", "com.example", "foo", "1.0"); got != filepath.Join(p.WorkspaceDir("ws"), "maven-source-cache", "com/example", "foo", "1.0") {
		t.Fatalf("MavenSourceCacheDir = %q", got)
	}
}

func TestAllocBuildID(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	id1, err := p.AllocBuildID("ws", "demo")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := p.AllocBuildID("ws", "demo")
	if err != nil {
		t.Fatal(err)
	}
	if id1 != "0001" || id2 != "0002" {
		t.Fatalf("ids = %q, %q", id1, id2)
	}
}

func TestWorkspaceState(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	if _, err := p.ReadState("ws"); err == nil {
		t.Fatal("expected error reading missing state")
	}
	if err := p.WriteState("ws", &WorkspaceState{CurrentProfile: "dev"}); err != nil {
		t.Fatal(err)
	}
	got, err := p.ReadState("ws")
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrentProfile != "dev" {
		t.Fatalf("profile = %q", got.CurrentProfile)
	}
}

func TestBuildRunState(t *testing.T) {
	p := New(filepath.Join(t.TempDir(), "data"))
	bs := &BuildState{Status: "running", Command: "mvn package"}
	if err := p.WriteBuildState("ws", "demo", "0001", bs); err != nil {
		t.Fatal(err)
	}
	got, err := p.ReadBuildState("ws", "demo", "0001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "running" || got.Command != "mvn package" {
		t.Fatalf("build state = %+v", got)
	}
	rs := &RunState{Status: "running", PID: 123, Port: 8080, Profile: "dev"}
	if err := p.WriteRunState("ws", "svc", "0001", rs); err != nil {
		t.Fatal(err)
	}
	gotr, err := p.ReadRunState("ws", "svc", "0001")
	if err != nil {
		t.Fatal(err)
	}
	if gotr.PID != 123 || gotr.Port != 8080 || gotr.Profile != "dev" {
		t.Fatalf("run state = %+v", gotr)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/storage/...`
Expected: FAIL。

- [ ] **Step 3: 写实现**

```go
// Package storage encapsulates data-dir path conventions, id allocation, and state.json I/O.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

// Paths resolves all on-disk paths under a data root.
type Paths struct {
	DataDir string
}

// New creates a Paths rooted at dataDir.
func New(dataDir string) *Paths { return &Paths{DataDir: dataDir} }

func (p *Paths) WorkspaceDir(ws string) string {
	return filepath.Join(p.DataDir, "ws-"+ws)
}
func (p *Paths) ProjectDir(ws, project string) string {
	return filepath.Join(p.WorkspaceDir(ws), "projects", project)
}
func (p *Paths) ServiceDir(ws, service string) string {
	return filepath.Join(p.WorkspaceDir(ws), "services", service)
}
func (p *Paths) BuildDir(ws, project, buildID string) string {
	return filepath.Join(p.ProjectDir(ws, project), "builds", buildID)
}
func (p *Paths) RunDir(ws, service, runID string) string {
	return filepath.Join(p.ServiceDir(ws, service), "runs", runID)
}
func (p *Paths) StateFile(ws string) string {
	return filepath.Join(p.WorkspaceDir(ws), "state.json")
}
func (p *Paths) BuildStateFile(ws, project, buildID string) string {
	return filepath.Join(p.BuildDir(ws, project, buildID), "state.json")
}
func (p *Paths) RunStateFile(ws, service, runID string) string {
	return filepath.Join(p.RunDir(ws, service, runID), "state.json")
}

// MavenSourceCacheDir returns the extraction cache dir for a Maven coordinate.
// <g> is the groupId with dots converted to path separators.
func (p *Paths) MavenSourceCacheDir(ws, g, a, v string) string {
	gPath := filepath.FromSlash(strings.ReplaceAll(g, ".", "/"))
	return filepath.Join(p.WorkspaceDir(ws), "maven-source-cache", gPath, a, v)
}

// WorkspaceState is the runtime state (state.json).
type WorkspaceState struct {
	CurrentProfile string `json:"current-profile"`
}

// BuildState is the per-build state (builds/<id>/state.json).
type BuildState struct {
	Status    string `json:"status"`     // running | success | failed
	StartedAt string `json:"startedAt"`  // RFC3339
	EndedAt   string `json:"endedAt"`    // RFC3339
	ExitCode  int    `json:"exitCode"`
	Command   string `json:"command"`
}

// RunState is the per-run state (runs/<id>/state.json).
type RunState struct {
	Status    string `json:"status"`     // running | stopped | exited | crashed
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Profile   string `json:"profile"`
	StartedAt string `json:"startedAt"`  // RFC3339
	EndedAt   string `json:"endedAt"`     // RFC3339
	ExitCode  int    `json:"exitCode"`
}

// ReadState reads state.json for a workspace.
func (p *Paths) ReadState(ws string) (*WorkspaceState, error) {
	b, err := os.ReadFile(p.StateFile(ws))
	if err != nil {
		return nil, err
	}
	var s WorkspaceState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteState writes state.json for a workspace.
func (p *Paths) WriteState(ws string, s *WorkspaceState) error {
	return writeJSON(p.StateFile(ws), p.WorkspaceDir(ws), s)
}

// ReadBuildState reads a build's state.json.
func (p *Paths) ReadBuildState(ws, project, buildID string) (*BuildState, error) {
	b, err := os.ReadFile(p.BuildStateFile(ws, project, buildID))
	if err != nil {
		return nil, err
	}
	var s BuildState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteBuildState writes a build's state.json.
func (p *Paths) WriteBuildState(ws, project, buildID string, s *BuildState) error {
	return writeJSON(p.BuildStateFile(ws, project, buildID), p.BuildDir(ws, project, buildID), s)
}

// ReadRunState reads a run's state.json.
func (p *Paths) ReadRunState(ws, service, runID string) (*RunState, error) {
	b, err := os.ReadFile(p.RunStateFile(ws, service, runID))
	if err != nil {
		return nil, err
	}
	var s RunState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteRunState writes a run's state.json.
func (p *Paths) WriteRunState(ws, service, runID string, s *RunState) error {
	return writeJSON(p.RunStateFile(ws, service, runID), p.RunDir(ws, service, runID), s)
}

func writeJSON(path, dir string, v interface{}) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// AllocBuildID allocates the next 4-digit build-id for a project.
func (p *Paths) AllocBuildID(ws, project string) (string, error) {
	return p.allocID(p.ProjectDir(ws, project), "build-id")
}

// AllocRunID allocates the next 4-digit run-id for a service.
func (p *Paths) AllocRunID(ws, service string) (string, error) {
	return p.allocID(p.ServiceDir(ws, service), "run-id")
}

func (p *Paths) allocID(dir, name string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	idFile := filepath.Join(dir, name)
	lockFile := idFile + ".lock"
	if err := lockWait(lockFile); err != nil {
		return "", err
	}
	defer os.Remove(lockFile)

	b, _ := os.ReadFile(idFile)
	n, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	n++
	if err := os.WriteFile(idFile, []byte(strconv.Itoa(n)), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n), nil
}

// lockWait acquires an exclusive lock file (portable: O_CREATE|O_EXCL with retry).
// NOTE: a crashed process may leave a stale lock; stale-detection by mtime is a future refinement.
func lockWait(path string) error {
	for i := 0; i < 100; i++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return apperrors.New(apperrors.InternalError, "id lock: "+err.Error())
		}
		time.Sleep(10 * time.Millisecond)
	}
	return apperrors.New(apperrors.InternalError, "id lock timeout")
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/storage/...`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/storage
git commit -m "feat(storage): 添加路径助手、id 分配与 state 读写"
```

---

### Task 5: common（信封 + 绑定 + 错误映射）

**Files:**
- Create: `internal/common/envelope.go`, `internal/common/bind.go`, `internal/common/error_map.go`
- Test: `internal/common/common_test.go`

**Interfaces:**
- Consumes: `apperrors`。
- Produces: `common.Envelope`、`common.OK(data)`、`common.Fail(err)`、`common.Bind(c, req)`、`common.MapError(err)`。

- [ ] **Step 1: 写失败测试**

```go
package common

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/apperrors"
)

func TestOKAndFail(t *testing.T) {
	ok := OK(map[string]int{"n": 1})
	if !ok.Success || ok.Data == nil || ok.Error != nil {
		t.Fatalf("OK = %+v", ok)
	}
	fail := Fail(apperrors.New(apperrors.NotFound, "x"))
	if fail.Success || fail.Error == nil || fail.Error.Code != "NOT_FOUND" {
		t.Fatalf("Fail = %+v", fail)
	}
}

func TestMapError(t *testing.T) {
	e := MapError(apperrors.New(apperrors.BuildFailed, "boom"))
	if e.Error.Code != "BUILD_FAILED" {
		t.Fatalf("code = %s", e.Error.Code)
	}
	e2 := MapError(errFoo)
	if e2.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("unknown err should map to INTERNAL_ERROR, got %s", e2.Error.Code)
	}
}

var errFoo = echo.NewHTTPError(500, "foo")

func TestBindRequired(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var got struct {
		Workspace string `json:"workspace" validate:"required"`
	}
	if err := Bind(c, &got); err == nil {
		t.Fatal("expected validation error for missing workspace")
	}
}

func TestBindOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"workspace":"ws"}`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var got struct {
		Workspace string `json:"workspace" validate:"required"`
	}
	if err := Bind(c, &got); err != nil {
		t.Fatal(err)
	}
	if got.Workspace != "ws" {
		t.Fatalf("workspace = %q", got.Workspace)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/common/...`
Expected: FAIL。

- [ ] **Step 3: 写 envelope.go**

```go
// Package common provides transport-shared helpers: response envelope, request binding, and error mapping.
package common

import "github.com/ravenmk2/jungle/internal/apperrors"

// Envelope is the unified RPC-style response envelope.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorBody  `json:"error"`
}

// ErrorBody is the error part of the envelope.
type ErrorBody struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []apperrors.Detail `json:"details,omitempty"`
}

// OK builds a success envelope.
func OK(data interface{}) Envelope {
	return Envelope{Success: true, Data: data}
}

// Fail builds a failure envelope from a typed error.
func Fail(err *apperrors.Error) Envelope {
	return Envelope{Success: false, Error: &ErrorBody{
		Code:    string(err.Code),
		Message: err.Message,
		Details: err.Details,
	}}
}
```

- [ ] **Step 4: 写 error_map.go**

```go
package common

import (
	"errors"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

// MapError maps any error to an envelope. Typed errors keep their code;
// unknown errors become INTERNAL_ERROR.
func MapError(err error) Envelope {
	var te *apperrors.Error
	if errors.As(err, &te) {
		return Fail(te)
	}
	return Fail(apperrors.New(apperrors.InternalError, "internal error"))
}
```

- [ ] **Step 5: 写 bind.go**

```go
package common

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/apperrors"
)

var validate = validator.New()

// Bind decodes the JSON request body into req and validates it.
// Returns a typed ValidationError (with field-level details) on failure.
func Bind(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return apperrors.New(apperrors.ValidationError, "invalid request body")
	}
	if err := validate.Struct(req); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) *apperrors.Error {
	details := []apperrors.Detail{}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			details = append(details, apperrors.Detail{
				Code:    fe.Tag(),
				Message: "field " + fe.Field() + " failed " + fe.Tag(),
				Target:  fe.Field(),
			})
		}
	}
	return apperrors.NewWithDetails(apperrors.ValidationError, "validation failed", details)
}
```

- [ ] **Step 6: 运行测试确认通过**

Run: `go test ./internal/common/...`
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/common
git commit -m "feat(common): 添加响应信封、请求绑定与错误映射"
```

---

### Task 6: service/common Ports + workspace usecase + 其余 stub

**Files:**
- Create: `internal/service/common/ports.go`
- Create: `internal/service/workspace/service.go`
- Test: `internal/service/workspace/service_test.go`
- Create: `internal/service/{build,run,database,search}/service.go`（stub）

**Interfaces:**
- Consumes: `config`、`storage`、`apperrors`。
- Produces: `service/common.Searcher`/`SearchOpts`(含 `Context{A,B,C}`)/`Match`/`SearchResult`、`workspace.Service`（List/Get/SwitchProfile）、`build.Service`+`RunResult`、`run.Service`+`StartResult`、`database.Service`+`QueryOpts`/`QueryResult`、`search.Service`；各 stub 的 `New` 返回对应 `ErrNotImplemented` 错误码。

- [ ] **Step 1: 写 service/common/ports.go**

```go
// Package common defines cross-cutting Port interfaces shared by multiple usecases and infra.
package common

import "context"

// SearchOpts are shared search options (used by docs/maven/log search).
type SearchOpts struct {
	Literal    bool
	IgnoreCase bool
	Glob       string
	Type       string
	Context    ContextLines
	MaxCount   int
	Raw        bool
}

// ContextLines configures surrounding lines (spec 6.4.1: context {A,B,C}).
type ContextLines struct {
	A int // before
	B int // after
	C int // both (shorthand for A=B=C)
}

// Match is a single search hit.
type Match struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	LineContent string `json:"lineContent"`
	Type       string `json:"type"` // "match" | "context"
}

// SearchResult is the truncated (non-paged) search response.
type SearchResult struct {
	Items     []Match `json:"items"`
	Total     int     `json:"total"`
	Truncated bool    `json:"truncated"`
}

// Searcher is the Port for full-text search (implemented by infra/searcher).
type Searcher interface {
	Search(ctx context.Context, roots []string, query string, opts SearchOpts) (*SearchResult, error)
}
```

- [ ] **Step 2: 写 workspace 失败测试**

```go
package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func setup(t *testing.T, configDir, dataDir string) *Service {
	t.Helper()
	wsDir := filepath.Join(configDir, "workspaces")
	os.MkdirAll(wsDir, 0o755)
	os.WriteFile(filepath.Join(wsDir, "demo.toml"), []byte(`[java]
version = 8
home = "/jdk"
[maven]
home = "/mvn"
repo = "/repo"
[profiles]
items = ["dev", "staging"]
`), 0o644)
	return New(configDir, dataDir)
}

func TestList(t *testing.T) {
	s := setup(t, t.TempDir(), t.TempDir())
	got, err := s.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "demo" {
		t.Fatalf("List = %+v", got)
	}
}

func TestGetAndSwitch(t *testing.T) {
	s := setup(t, t.TempDir(), t.TempDir())
	v, err := s.Get(nil, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if v.CurrentProfile != "" {
		t.Fatalf("initial profile = %q, want empty", v.CurrentProfile)
	}
	if err := s.SwitchProfile(nil, "demo", "dev"); err != nil {
		t.Fatal(err)
	}
	v, _ = s.Get(nil, "demo")
	if v.CurrentProfile != "dev" {
		t.Fatalf("after switch profile = %q", v.CurrentProfile)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/service/workspace/...`
Expected: FAIL。

- [ ] **Step 4: 写 workspace/service.go**

```go
// Package workspace is the workspace/profile usecase.
package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ravenmk2/jungle/internal/apperrors"
	"github.com/ravenmk2/jungle/internal/config"
	"github.com/ravenmk2/jungle/internal/storage"
)

// View is the workspace view returned by Get.
type View struct {
	Name           string `json:"name"`
	CurrentProfile string `json:"currentProfile"`
	*config.WorkspaceConfig
}

// Service is the workspace usecase Port.
type Service interface {
	List(ctx context.Context) ([]string, error)
	Get(ctx context.Context, workspace string) (*View, error)
	SwitchProfile(ctx context.Context, workspace, profile string) error
}

type service struct {
	configDir string
	dataDir   string
}

// New creates a workspace Service.
func New(configDir, dataDir string) Service {
	return &service{configDir: configDir, dataDir: dataDir}
}

func (s *service) wsPath(name string) string {
	return filepath.Join(s.configDir, "workspaces", name+".toml")
}

func (s *service) List(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.configDir, "workspaces"))
	if err != nil {
		return nil, apperrors.New(apperrors.NotFound, "no workspaces")
	}
	var out []string
	for _, e := range entries {
		if name := strings.TrimSuffix(e.Name(), ".toml"); name != e.Name() {
			out = append(out, name)
		}
	}
	return out, nil
}

func (s *service) Get(ctx context.Context, workspace string) (*View, error) {
	cfg, err := config.LoadWorkspace(s.wsPath(workspace))
	if err != nil {
		return nil, apperrors.New(apperrors.NotFound, "workspace not found: "+workspace)
	}
	v := &View{Name: workspace, WorkspaceConfig: cfg}
	p := storage.New(s.dataDir)
	if st, err := p.ReadState(workspace); err == nil {
		v.CurrentProfile = st.CurrentProfile
	}
	return v, nil
}

func (s *service) SwitchProfile(ctx context.Context, workspace, profile string) error {
	if _, err := config.LoadWorkspace(s.wsPath(workspace)); err != nil {
		return apperrors.New(apperrors.NotFound, "workspace not found: "+workspace)
	}
	p := storage.New(s.dataDir)
	return p.WriteState(workspace, &storage.WorkspaceState{CurrentProfile: profile})
}
```

- [ ] **Step 5: 写其余 usecase stub**

`internal/service/build/service.go`:
```go
// Package build is the build usecase (stub).
package build

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type RunOpts struct {
	Goal  string
	Clean bool
	Test  bool
}

// RunResult is the build result (spec 6.1: build-id + status).
type RunResult struct {
	BuildID string `json:"buildId"`
	Status  string `json:"status"` // success | failed
}

type Service interface {
	Run(ctx context.Context, workspace, project string, opts RunOpts) (*RunResult, error)
	LogSearch(ctx context.Context, workspace, project, buildID, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Run(ctx context.Context, workspace, project string, opts RunOpts) (*RunResult, error) {
	return nil, apperrors.New(apperrors.BuildFailed, "not implemented")
}
func (s *service) LogSearch(ctx context.Context, workspace, project, buildID, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
```

`internal/service/run/service.go`:
```go
// Package run is the service-run usecase (stub).
package run

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type StartResult struct {
	RunID string `json:"runId"`
	Port  int    `json:"port"`
}

type Service interface {
	Start(ctx context.Context, workspace, service string) (*StartResult, error)
	Stop(ctx context.Context, workspace, service string) error
	Status(ctx context.Context, workspace, service string) (interface{}, error)
	LogSearch(ctx context.Context, workspace, service, runID, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Start(ctx context.Context, workspace, service string) (*StartResult, error) {
	return nil, apperrors.New(apperrors.RunFailed, "not implemented")
}
func (s *service) Stop(ctx context.Context, workspace, service string) error {
	return apperrors.New(apperrors.ServiceNotRunning, "not implemented")
}
func (s *service) Status(ctx context.Context, workspace, service string) (interface{}, error) {
	return nil, apperrors.New(apperrors.ServiceNotRunning, "not implemented")
}
func (s *service) LogSearch(ctx context.Context, workspace, service, runID, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
```

`internal/service/database/service.go`:
```go
// Package database is the DB usecase (stub).
package database

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
)

type QueryOpts struct {
	MaxRows int // default 200 (spec 6.3/G7)
}

type QueryResult struct {
	Columns   []string        `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	Truncated bool            `json:"truncated"`
}

type Service interface {
	Reset(ctx context.Context, workspace, database string) error
	Query(ctx context.Context, workspace, database, sql string, opts QueryOpts) (*QueryResult, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Reset(ctx context.Context, workspace, database string) error {
	return apperrors.New(apperrors.DBResetFailed, "not implemented")
}
func (s *service) Query(ctx context.Context, workspace, database, sql string, opts QueryOpts) (*QueryResult, error) {
	return nil, apperrors.New(apperrors.DBQueryFailed, "not implemented")
}
```

`internal/service/search/service.go`:
```go
// Package search is the search usecase (stub).
package search

import (
	"context"

	"github.com/ravenmk2/jungle/internal/apperrors"
	svc "github.com/ravenmk2/jungle/internal/service/common"
)

type Service interface {
	Docs(ctx context.Context, workspace, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
	MavenSource(ctx context.Context, workspace, dependency, project, query string, opts svc.SearchOpts) (*svc.SearchResult, error)
	MavenRead(ctx context.Context, workspace, dependency, project, file, class string, rng [2]int) (interface{}, error)
}

type service struct{}

func New() Service { return &service{} }

func (s *service) Docs(ctx context.Context, workspace, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
func (s *service) MavenSource(ctx context.Context, workspace, dependency, project, query string, opts svc.SearchOpts) (*svc.SearchResult, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
func (s *service) MavenRead(ctx context.Context, workspace, dependency, project, file, class string, rng [2]int) (interface{}, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "not implemented")
}
```

- [ ] **Step 6: 运行测试确认通过**

Run: `go test ./internal/service/...`
Expected: PASS（workspace 测试通过，stub 无测试）。

- [ ] **Step 7: 提交**

```bash
git add internal/service
git commit -m "feat(service): 添加 Searcher Port、workspace 最小实现与其余 stub"
```

---

### Task 7: infra stubs

**Files:**
- Create: `internal/infra/{maven,mysql,process,searcher}/impl.go`

**Interfaces:**
- Consumes: `service/common`（Port）。
- Produces: 各适配器 struct + `New()`，方法返回 ErrNotImplemented。

- [ ] **Step 1: 写四个 stub**

`internal/infra/searcher/impl.go`:
```go
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
```

`internal/infra/maven/impl.go`:
```go
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
// FetchSourcesBatch runs `mvn dependency:sources` for a project (spec 6.4.3 project scope).
func (a *Adapter) FetchSourcesBatch(repoDir string) error {
	return apperrors.New(apperrors.SearchFailed, "maven dependency:sources not implemented")
}
func (a *Adapter) ListDependencies(repoDir string) ([]string, error) {
	return nil, apperrors.New(apperrors.SearchFailed, "maven dep:list not implemented")
}
```

`internal/infra/mysql/impl.go`:
```go
// Package mysql is the MySQL adapter (stub).
package mysql

import "github.com/ravenmk2/jungle/internal/apperrors"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Reset(dsn, initSQL string) error {
	return apperrors.New(apperrors.DBResetFailed, "mysql reset not implemented")
}
func (a *Adapter) Query(dsn, sql string, maxRows int) (columns []string, rows [][]interface{}, truncated bool, err error) {
	return nil, nil, false, apperrors.New(apperrors.DBQueryFailed, "mysql query not implemented")
}
```

`internal/infra/process/impl.go`:
```go
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
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/infra/...`
Expected: 无输出。

- [ ] **Step 3: 提交**

```bash
git add internal/infra
git commit -m "feat(infra): 添加 maven/mysql/process/searcher stub"
```

---

### Task 8: handlers（Echo + health + workspace）

**Files:**
- Create: `internal/handlers/server.go`, `internal/handlers/health.go`, `internal/handlers/workspace.go`
- Test: `internal/handlers/handlers_test.go`

**Interfaces:**
- Consumes: `common`、`service/workspace`。
- Produces: `handlers.New(eng, wsSvc)`（挂载 `/api/*` 路由）、`handlers.Mount(eng, ...)`。

- [ ] **Step 1: 写失败测试**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/common"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

func setupEng(t *testing.T) *echo.Echo {
	t.Helper()
	configDir := t.TempDir()
	wsDir := filepath.Join(configDir, "workspaces")
	os.MkdirAll(wsDir, 0o755)
	os.WriteFile(filepath.Join(wsDir, "demo.toml"), []byte(`[java]
version = 8
home = "/jdk"
[profiles]
items = ["dev", "staging"]
`), 0o644)
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
	json.Unmarshal(rec.Body.Bytes(), &env)
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
	var env common.Envelope
	json.Unmarshal(rec.Body.Bytes(), &env)
	if !env.Success {
		t.Fatalf("envelope = %+v", env)
	}
}

func TestWorkspaceGetAndSwitch(t *testing.T) {
	e := setupEng(t)
	// get
	req := httptest.NewRequest("POST", "/api/workspace/get", strings.NewReader(`{"workspace":"demo"}`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d", rec.Code)
	}
	// switch-profile
	req = httptest.NewRequest("POST", "/api/workspace/switch-profile", strings.NewReader(`{"workspace":"demo","profile":"dev"}`))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("switch status = %d", rec.Code)
	}
	var env common.Envelope
	json.Unmarshal(rec.Body.Bytes(), &env)
	if !env.Success {
		t.Fatalf("switch envelope = %+v", env)
	}
}

func TestWorkspaceGetValidation(t *testing.T) {
	e := setupEng(t)
	// missing workspace field -> VALIDATION_ERROR envelope (still HTTP 200)
	req := httptest.NewRequest("POST", "/api/workspace/get", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var env common.Envelope
	json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Success || env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %+v", env)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/handlers/...`
Expected: FAIL（`New` 未定义）。

- [ ] **Step 3: 写 server.go**

```go
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
```

- [ ] **Step 4: 写 health.go**

```go
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/common"
)

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, common.OK(map[string]string{"status": "ok"}))
}
```

- [ ] **Step 5: 写 workspace.go**

```go
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ravenmk2/jungle/internal/common"
	"github.com/ravenmk2/jungle/internal/service/workspace"
)

func mountWorkspace(eng *echo.Echo, wsSvc workspace.Service) {
	g := eng.Group("/api/workspace")
	g.POST("/list", func(c echo.Context) error {
		out, err := wsSvc.List(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusOK, common.MapError(err))
		}
		return c.JSON(http.StatusOK, common.OK(map[string][]string{"items": out}))
	})
	g.POST("/get", func(c echo.Context) error {
		var req struct {
			Workspace string `json:"workspace" validate:"required"`
		}
		if err := common.Bind(c, &req); err != nil {
			return c.JSON(http.StatusOK, common.MapError(err))
		}
		out, err := wsSvc.Get(c.Request().Context(), req.Workspace)
		if err != nil {
			return c.JSON(http.StatusOK, common.MapError(err))
		}
		return c.JSON(http.StatusOK, common.OK(out))
	})
	g.POST("/switch-profile", func(c echo.Context) error {
		var req struct {
			Workspace string `json:"workspace" validate:"required"`
			Profile   string `json:"profile" validate:"required"`
		}
		if err := common.Bind(c, &req); err != nil {
			return c.JSON(http.StatusOK, common.MapError(err))
		}
		if err := wsSvc.SwitchProfile(c.Request().Context(), req.Workspace, req.Profile); err != nil {
			return c.JSON(http.StatusOK, common.MapError(err))
		}
		return c.JSON(http.StatusOK, common.OK(map[string]string{"workspace": req.Workspace, "currentProfile": req.Profile}))
	})
}
```

- [ ] **Step 6: 运行测试确认通过**

Run: `go test ./internal/handlers/...`
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/handlers
git commit -m "feat(handlers): 添加 Echo 路由挂载、health 与 workspace 端点"
```

---

### Task 9: mcp 占位路由 + Tool/Registry

**Files:**
- Create: `internal/mcp/mcp.go`

**Interfaces:**
- Produces: `mcp.Tool`、`mcp.Registry`、`mcp.Mount(eng)`（`/mcp` 返回 501 占位）。

- [ ] **Step 1: 写实现**

```go
// Package mcp mounts the /mcp Streamable HTTP placeholder and defines the tool registry.
// Real MCP Go SDK wiring is deferred to the tool-wiring phase.
package mcp

import (
	"net/http"

	"github.com/labstack/echo/v4"
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
	eng.Any("/mcp", func(c echo.Context) error {
		return c.JSON(http.StatusOK, common.Fail(apperrors.New(apperrors.InternalError, "mcp not implemented")))
	})
}
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/mcp/...`
Expected: 无输出。

- [ ] **Step 3: 提交**

```bash
git add internal/mcp
git commit -m "feat(mcp): 添加 /mcp 占位路由与 Tool/Registry"
```

---

### Task 10: app 组装根 + main 入口

**Files:**
- Create: `internal/app/app.go`
- Create: `cmd/jungle/main.go`

**Interfaces:**
- Consumes: `config`、`handlers`、`mcp`、`service/workspace`、logrus。
- Produces: `app.Run(configDir, dataDir, addr string) error`（加载配置、组装、启动 HTTP server）。

- [ ] **Step 1: 写 app.go**

```go
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
```

> 注：脚手架阶段仅 `wsSvc` 有 handler 消费者，故只组装它（YAGNI）。build/run/database/search 的 service 与全部 infra 在对应端点接入时再装配。

- [ ] **Step 2: 写 main.go**

```go
// Package main is the jungle binary entrypoint.
package main

import (
	"flag"

	"github.com/ravenmk2/jungle/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("addr", "", "listen address (overrides config)")
	configDir := flag.String("config-dir", "./config", "config directory")
	dataDir := flag.String("data-dir", "", "data directory (overrides config)")
	flag.Parse()

	if err := app.Run(*configDir, *dataDir, *addr); err != nil {
		logrus.Fatal(err)
	}
}
```

- [ ] **Step 3: 验证编译**

Run: `go build ./...`
Expected: 无输出。

- [ ] **Step 4: 冒烟测试**

先准备配置（PowerShell here-string 产出真实换行，勿用单引号+`` `n ``）：
```powershell
if (-not (Test-Path config/config.toml)) { Copy-Item config/config-sample.toml config/config.toml }
New-Item -ItemType Directory -Path config/workspaces -Force | Out-Null
$demo = @'
[java]
version = 8
home = "/jdk"
[maven]
home = "/mvn"
repo = "/repo"
[profiles]
items = ["dev"]
'@
Set-Content -Path config/workspaces/demo.toml -Value $demo -Encoding utf8
```
构建并后台运行（用预编译二进制，便于精确停进程）：
```powershell
go build -o jungle.exe ./cmd/jungle
$proc = Start-Process -FilePath ".\jungle.exe" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 2
```
调 health：
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:7788/api/health" -Method GET
```
Expected: `success=True`，`data.status = "ok"`。
调 workspace/list：
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:7788/api/workspace/list" -Method POST -ContentType "application/json" -Body "{}"
```
Expected: `success=True`，`data.items` 含 `demo`。
清理：
```powershell
Stop-Process -Id $proc.Id
Remove-Item jungle.exe
```

- [ ] **Step 5: 提交**

```bash
git add internal/app cmd/jungle
git commit -m "feat(app): 添加组装根与 main 入口"
```

---

### Task 11: 收尾（README/AGENTS 已在 Task 1；最终验证）

**Files:**
- Verify: 全仓 `go build ./...`、`go vet ./...`、`go test ./...`

- [ ] **Step 1: 全量静态检查**

Run:
```bash
go build ./...
go vet ./...
go test ./...
```
Expected: 全部通过。

- [ ] **Step 2: 格式化**

Run: `gofmt -s -w .`
Expected: 无输出。

- [ ] **Step 3: 提交收尾**

```bash
git add -A
git commit -m "chore: 脚手架收尾格式化" --allow-empty
```
（若 `gofmt` 无改动则空提交跳过；有改动则提交。）

---

## 完成标准

- `go build ./...` / `go vet ./...` / `go test ./...` 全绿。
- `jungle` 可启动；`GET /api/health` 与 `POST /api/workspace/{list,get,switch-profile}` 可调用并返回统一信封。
- 其余 usecase/infra 为 stub（返回对应 `ErrNotImplemented` 错误码），目录结构与 spec 第 3 节一致。
- `.gitignore`、`config/config-sample.toml`、`Makefile`、`.golangci.yml` 就位。
