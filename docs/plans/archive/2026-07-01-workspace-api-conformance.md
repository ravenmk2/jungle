# Workspace API 收敛实现计划

> **给执行者:** 用 ack-executing-plan 逐任务实现本计划。步骤用 checkbox(`- [ ]`)跟踪。

**目标:** 让 `/api/workspace/list`、`/get`、`/switch-profile` 三端点的实现收敛到 `docs/specs/workspace-api.md` 契约。

**架构:** 三端点已最小实现于 `internal/handlers/workspace.go` 与 `internal/service/workspace/service.go`。本计划仅做收敛式修改，不新增文件：(1) `list` 的 `data` 去掉 `items` 包装、空结果返回 `[]`；(2) `get` 给 `config.WorkspaceConfig` 及子结构补 `json` tag（camelCase）并对 `password` 脱敏；(3) `VALIDATION_ERROR` 的 `target`/`message` 用请求体 json 字段名而非 Go 字段名。

**技术栈:** Go 1.26.4、Echo v5 (`github.com/labstack/echo/v5` v5.2.1)、go-playground/validator v10 (`v10.30.3`)、pelletier/go-toml/v2、module `github.com/ravenmk2/jungle`。

## 全局约束

- 方法: 三端点均为 `POST`。
- Content-Type: `application/json`。
- 路径前缀: `/api/workspace`，kebab-case。
- 响应信封: `{ success, data, error: { code, message, details? } }`。业务成功/失败统一 HTTP `200`；`4xx`/`5xx` 仅限传输/系统层（404→`NOT_FOUND`、405→`METHOD_NOT_ALLOWED`、500→`INTERNAL_ERROR`）。
- `data` 属性与请求体属性均使用 **camelCase**。
- 请求体需携带 `workspace` 字段；**`/api/workspace/list` 例外**。
- `get` 响应 `databases` 条目**不含 `password`**（脱敏）。

测试/lint 命令：`go test ./...`（或 `make test`）、`go vet ./...`、`gofmt -s -w .`、`golangci-lint run`（或 `make lint`）。

---

## 文件结构

无新文件，仅修改既有文件：

- `internal/service/workspace/service.go` — `List` 返回非 nil 空切片。
- `internal/handlers/workspace.go` — `list` handler 的 `data` 直接传 `[]string`。
- `internal/config/config.go` — `WorkspaceConfig` 及子结构补 `json` tag；`Database.Password` 用 `json:"-"`；`Database.InitSQL`/`Service.Port`/`Service.Database` 用 `omitempty`。
- `internal/common/bind.go` — `validate` 改为工厂构造并注册 `RegisterTagNameFunc`，使校验错误用 json 字段名。
- 测试：`internal/handlers/handlers_test.go`、`internal/service/workspace/service_test.go`、`internal/common/common_test.go`。

---

### 任务 1: `/api/workspace/list` 数据形态收敛

**文件:**
- 修改: `internal/service/workspace/service.go`（`List` 方法）
- 修改: `internal/handlers/workspace.go`（`list` handler）
- 测试: `internal/service/workspace/service_test.go`、`internal/handlers/handlers_test.go`

**接口:**
- 消费: 无前置任务。
- 产出: `workspace.Service.List` 返回 `[]string`（目录存在但无 `*.toml` 时返回长度 0 的非 nil 切片）；`list` handler 响应 `data` 直接为字符串数组。

- [ ] **步骤 1: 写失败测试**

在 `internal/service/workspace/service_test.go` 末尾追加：

```go
func TestListEmpty(t *testing.T) {
	configDir := t.TempDir()
	os.MkdirAll(filepath.Join(configDir, "workspaces"), 0o755)
	s := New(configDir, t.TempDir())
	got, err := s.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("List = nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("List = %+v, want empty", got)
	}
}
```

将 `internal/handlers/handlers_test.go` 中的 `TestWorkspaceList` 整体替换为：

```go
func TestWorkspaceList(t *testing.T) {
	e := setupEng(t)
	req := httptest.NewRequest("POST", "/api/workspace/list", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got struct {
		Success bool           `json:"success"`
		Data    []string       `json:"data"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, rec.Body.String())
	}
	if !got.Success || len(got.Data) != 1 || got.Data[0] != "demo" {
		t.Fatalf("list = %+v body=%s", got, rec.Body.String())
	}
}
```

- [ ] **步骤 2: 跑测试确认失败**

运行: `go test ./internal/handlers/ ./internal/service/workspace/ -run 'TestWorkspaceList|TestListEmpty' -v`
预期: FAIL — `TestWorkspaceList` 因 `data` 当前为 `{"items":[...]}` 对象、反序列化进 `[]string` 失败；`TestListEmpty` 因 `List` 空目录返回 `nil`。

- [ ] **步骤 3: 写最小实现**

`internal/service/workspace/service.go` 的 `List` 中，将：

```go
	var out []string
```

改为：

```go
	out := make([]string, 0, len(entries))
```

`internal/handlers/workspace.go` 的 `list` handler 中，将：

```go
		return c.JSON(http.StatusOK, common.OK(map[string][]string{"items": out}))
```

改为：

```go
		return c.JSON(http.StatusOK, common.OK(out))
```

- [ ] **步骤 4: 跑测试确认通过**

运行: `go test ./internal/handlers/ ./internal/service/workspace/ -run 'TestWorkspaceList|TestListEmpty' -v`
预期: PASS。

- [ ] **步骤 5: 提交**

```bash
git add internal/service/workspace/service.go internal/handlers/workspace.go internal/service/workspace/service_test.go internal/handlers/handlers_test.go
git commit -m "refactor(workspace): list 端点 data 直接返回字符串数组"
```

---

### 任务 2: `/api/workspace/get` 响应 camelCase + 脱敏

**文件:**
- 修改: `internal/config/config.go`（`WorkspaceConfig` 及子结构补 `json` tag）
- 测试: `internal/handlers/handlers_test.go`（丰富 `setupEng` 的 toml + 新增 `TestWorkspaceGetShape`）

**接口:**
- 消费: 无前置任务。
- 产出: `config.WorkspaceConfig` 经 JSON 序列化时字段为 camelCase；`Database.Password` 不出现在任何 JSON 输出中。

- [ ] **步骤 1: 写失败测试**

在 `internal/handlers/handlers_test.go` 的 `setupEng` 中，将写入 `demo.toml` 的 toml 内容整体替换为（新增 `databases`、`services` 两节，含 `password` 与 `work-dir`/`init-sql`）：

```go
	os.WriteFile(filepath.Join(wsDir, "demo.toml"), []byte(`[java]
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
`), 0o644)
```

在 `internal/handlers/handlers_test.go` 末尾追加：

```go
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
	// password 必须脱敏（任意大小写的 key 都不得出现）
	if strings.Contains(body, "password") || strings.Contains(body, "Password") {
		t.Fatalf("password leaked: %s", body)
	}
	// camelCase key 必须存在
	for _, want := range []string{`"currentProfile"`, `"initSql"`, `"workDir"`, `"databases"`, `"services"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %s", want, body)
		}
	}
	// PascalCase 泄漏必须缺席
	for _, bad := range []string{`"InitSQL"`, `"WorkDir"`, `"CurrentProfile"`, `"Password"`, `"Databases"`} {
		if strings.Contains(body, bad) {
			t.Fatalf("PascalCase leak %q in %s", bad, body)
		}
	}
}
```

- [ ] **步骤 2: 跑测试确认失败**

运行: `go test ./internal/handlers/ -run TestWorkspaceGetShape -v`
预期: FAIL — 当前 `config` 子结构无 `json` tag，序列化为 PascalCase（`"InitSQL"`/`"WorkDir"`/`"Databases"`）且 `"Password":"pw"` 暴露；camelCase 断言缺失、PascalCase/脱敏断言触发。

- [ ] **步骤 3: 写最小实现**

将 `internal/config/config.go` 中 `WorkspaceConfig` 及其子结构体的字段 tag 整体替换为（在原 `toml` tag 旁补 `json` tag）：

```go
// WorkspaceConfig is a per-workspace config (config/workspaces/<name>.toml).
type WorkspaceConfig struct {
	Java      JavaSection         `toml:"java" json:"java"`
	Maven     MavenSection        `toml:"maven" json:"maven"`
	Docs      DocsSection         `toml:"docs" json:"docs"`
	Projects  map[string]Project  `toml:"projects" json:"projects"`
	Databases map[string]Database `toml:"databases" json:"databases"`
	Services  map[string]Service  `toml:"services" json:"services"`
	Profiles  ProfilesSection     `toml:"profiles" json:"profiles"`
}

type JavaSection struct {
	Version int    `toml:"version" json:"version"`
	Home    string `toml:"home" json:"home"`
}

type MavenSection struct {
	Home string `toml:"home" json:"home"`
	Repo string `toml:"repo" json:"repo"`
}

type DocsSection struct {
	Dirs []string `toml:"dirs" json:"dirs"`
}

type Project struct {
	Repo string `toml:"repo" json:"repo"`
}

type Database struct {
	Host     string `toml:"host" json:"host"`
	Port     int    `toml:"port" json:"port"`
	DB       string `toml:"db" json:"db"`
	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"-"`
	InitSQL  string `toml:"init-sql" json:"initSql,omitempty"`
}

type Service struct {
	Project  string `toml:"project" json:"project"`
	Module   string `toml:"module" json:"module"`
	WorkDir  string `toml:"work-dir" json:"workDir"`
	Port     int    `toml:"port" json:"port,omitempty"`
	Database string `toml:"database" json:"database,omitempty"`
}

type ProfilesSection struct {
	Items []string `toml:"items" json:"items"`
}
```

- [ ] **步骤 4: 跑测试确认通过**

运行: `go test ./internal/config/ ./internal/handlers/ -v`
预期: PASS（含既有 `TestLoadWorkspace`、`TestWorkspaceGetAndSwitch`，json tag 不影响 toml 加载）。

- [ ] **步骤 5: 提交**

```bash
git add internal/config/config.go internal/handlers/handlers_test.go
git commit -m "fix(workspace): get 响应改 camelCase 并脱敏 database password"
```

---

### 任务 3: `VALIDATION_ERROR` 用 json 字段名

**文件:**
- 修改: `internal/common/bind.go`（`validate` 改工厂构造 + `RegisterTagNameFunc`）
- 测试: `internal/common/common_test.go`

**接口:**
- 消费: 无前置任务。
- 产出: `common.Bind` 校验失败时，`apperrors.Detail.Target` 为请求体 json 字段名（如 `"workspace"`），`Message` 亦含该名。

- [ ] **步骤 1: 写失败测试**

在 `internal/common/common_test.go` 末尾追加：

```go
func TestBindValidationErrorTarget(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var got struct {
		Workspace string `json:"workspace" validate:"required"`
	}
	err := Bind(c, &got)
	if err == nil {
		t.Fatal("expected validation error")
	}
	ae, ok := err.(*apperrors.Error)
	if !ok {
		t.Fatalf("err type %T, want *apperrors.Error", err)
	}
	if ae.Code != apperrors.ValidationError {
		t.Fatalf("code = %s", ae.Code)
	}
	if len(ae.Details) != 1 {
		t.Fatalf("details = %+v", ae.Details)
	}
	d := ae.Details[0]
	if d.Code != "required" {
		t.Fatalf("detail code = %s", d.Code)
	}
	if d.Target != "workspace" {
		t.Fatalf("target = %q, want workspace", d.Target)
	}
}
```

- [ ] **步骤 2: 跑测试确认失败**

运行: `go test ./internal/common/ -run TestBindValidationErrorTarget -v`
预期: FAIL — 当前 `toValidationError` 用 `fe.Field()` 返回 Go 字段名 `"Workspace"`，`d.Target != "workspace"` 触发。

- [ ] **步骤 3: 写最小实现**

将 `internal/common/bind.go` 整体替换为：

```go
package common

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/ravenmk2/jungle/internal/apperrors"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return v
}

// Bind decodes the JSON request body into req and validates it.
// Returns a typed ValidationError (with field-level details) on failure.
func Bind(c *echo.Context, req interface{}) error {
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

- [ ] **步骤 4: 跑测试确认通过**

运行: `go test ./internal/common/ -v`
预期: PASS（含既有 `TestBindRequired`、`TestBindOK`、`TestOKAndFail`、`TestMapError`；TagNameFunc 仅改字段名解析，不破坏既有断言）。

- [ ] **步骤 5: 提交**

```bash
git add internal/common/bind.go internal/common/common_test.go
git commit -m "fix(common): 校验错误 target 用 json 字段名而非 Go 字段名"
```

---

## 收尾验证

三任务完成后整体回归：

- [ ] `go test ./...` 全绿
- [ ] `go vet ./...` 无告警
- [ ] `gofmt -s -w .` 无变更（或仅格式化）
- [ ] `golangci-lint run` 通过

## spec 覆盖对照

- `/list` Path/请求/响应/错误（NOT_FOUND 目录缺失）→ 任务 1（NOT_FOUND 目录缺失为既有行为，不变更，仅 `data` 形态收敛）。
- `/get` Path/请求/响应（camelCase + 脱敏）/错误 → 任务 2。
- `/switch-profile` Path/请求/响应/错误 → 已实现且与 spec 一致（`data` 回显 `workspace`/`currentProfile`，不校验 profile 成员），本计划无需改动；既有 `TestWorkspaceGetAndSwitch` 覆盖其行为。
- `VALIDATION_ERROR` 示例 `target: "workspace"` → 任务 3。
