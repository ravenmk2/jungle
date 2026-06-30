---
date: 2026-06-30
---

# Jungle 设计规范

Jungle 是一个为 AI Coding 设计的 Java 服务辅助开发环境，提供数据库操作、项目构建、服务运行、文档与 Maven 源码检索等能力，支持通过 HTTP API 与 MCP 两种方式接入。

本规范定义 Jungle 的总体架构、模块边界、配置与数据模型、接入层、各功能模块数据流、错误处理与测试策略，以及项目基础结构的脚手架范围。

## 1. 技术选型与关键决策

| 项                | 决策                                                                                            |
| ----------------- | ----------------------------------------------------------------------------------------------- |
| 项目名称          | Jungle（首字母大写）                                                                            |
| 语言/框架         | Golang + Echo                                                                                   |
| Module path       | `github.com/ravenmk2/jungle`                                                                    |
| 二进制形态        | 单二进制 `jungle`，一个 Echo server 同时挂 `/api/*` 与 `/mcp`                                   |
| MCP 传输          | Streamable HTTP                                                                                 |
| 检索技术          | 即时全文检索（ripgrep，无 rg 则 Go 回退）                                                       |
| 文档检索          | 支持配置多个目录，用 ripgrep 检索                                                               |
| Maven 源码检索    | 范围优先级 `dependency > project > global`；`project` 用 `mvn dependency:list`+`dependency:sources` 圈定依赖集合；sources.jar 解压到磁盘缓存 + rg；release 永不失效、SNAPSHOT 校验 metadata |
| 搜索响应          | 截断式 `{items, total, truncated}`，`maxCount` 默认 200、硬上限 1000                                  |
| 数据库            | MVP 仅 MySQL                                                                                    |
| DB reset 语义     | 仅执行 database 配置的 `init-sql`，Jungle 不主动清库（要求 SQL 自身幂等）                       |
| profile 语义      | MVP 仅作 Spring active profile 标签，Jungle 不解释内容                                          |
| 当前 profile 存储 | `./data/ws-{workspace}/state.json`，与静态配置分离                                              |
| 配置格式          | 人工编辑 = TOML，程序写入 = JSON                                                                |
| HTTP API 风格     | RPC-style（POST-only、统一信封、kebab-case 路径），遵循 `docs/references/rpc-style-http-api.md` |
| 日志库            | logrus（结构化日志）                                                                              |

## 2. 总体架构

单二进制 `jungle`，一个 Echo HTTP server 挂两组路由：

- `/api/*` — RPC-style HTTP API（薄适配器，`internal/handlers`）
- `/mcp` — MCP server，Streamable HTTP 传输（薄适配器，`internal/mcp`）

两者复用同一套 usecase 层（`internal/service`）。分层采用端口/适配器风格，`internal/app` 作为组装根负责所有组件初始化、静态依赖注入与 HTTP Server 初始化：

```txt
cmd/jungle ──> internal/app (组装根: 组件初始化、静态 DI、HTTP Server 初始化)
                ├──> internal/service (usecase + Port 接口)
                ├──> internal/handlers (/api 适配器)
                ├──> internal/mcp (/mcp 适配器)
                ├──> internal/infra (maven / mysql / process / searcher)
                ├──> internal/config (只读)
                └──> internal/storage (路径约定 + id 分配 + state 读写)
```

依赖铁律：

- `internal/service`（usecase）不得 import `handlers` / `mcp` / `infra`。
- 跨切面 Port 接口（如 `Searcher`）置于 `internal/service/common`，各 usecase 与 `infra` 平等依赖，避免 usecase 互相 import。
- `handlers` / `mcp` 可 import `service` + `common` + `apperrors`。
- `infra` 可 import `service` 的 Port（含 `service/common`）+ `config` + `storage` + `apperrors`。
- `apperrors` 为无依赖叶子包，各层均可 import。
- `internal/app` 是唯一可 import 全部包的组装根。
- `cmd/jungle` 仅解析参数并调用 `app`。

每个功能模块都是边界清晰、可独立测试的单元；HTTP 与 MCP 天然复用同一 usecase 实现。

## 3. 目录结构

```txt
Jungle/
├── cmd/
│   └── jungle/              # 入口：解析参数，调用 internal/app 启动
│       └── main.go
├── internal/
│   ├── app/                 # 组装根：组件初始化、静态 DI、HTTP Server 初始化
│   ├── service/             # usecase 层（不 import transport/infra）
│   │   ├── common/          # 跨切面 Port 接口（如 Searcher），各 usecase 与 infra 共用
│   │   ├── build/           # 构建 usecase + Port
│   │   ├── run/             # 服务运行 usecase + Port（start/stop/status + 运行日志检索）
│   │   ├── database/        # DB 重置/查询 usecase + Port
│   │   ├── search/          # 检索 usecase + Port（服务于 docs 与 maven 两个 HTTP 模块）
│   │   └── workspace/       # workspace/profile usecase + Port
│   ├── common/              # envelope.go 等共享（响应信封、错误映射、请求解码）
│   ├── apperrors/           # 类型化错误与错误码（无依赖叶子包）
│   ├── handlers/            # /api/* HTTP handlers（按 module 分文件）
│   ├── mcp/                 # /mcp Streamable HTTP 适配器 + tool 注册表
│   ├── config/              # TOML 加载，java/maven/docs/projects/databases/services/profiles 结构
│   ├── storage/             # 数据目录路径约定 + build-id/run-id 分配器 + state 读写
│   └── infra/               # 适配器实现
│       ├── maven/           # Maven: 构建执行 + dependency:list/sources + sources.jar 获取
│       ├── mysql/           # MySQL 连接/重置/查询
│       ├── process/         # 后台进程启停（Spring Boot 运行）
│       └── searcher/        # ripgrep shell out + Go 回退
├── config/                  # 静态配置（gitignored）
│   └── workspaces/          # {workspace}.toml
├── data/                    # 运行时数据（日志/state），gitignored，按 ws-{workspace} 划分
├── docs/
│   ├── references/          # 已有：rpc-style-http-api.md
│   └── specs/               # 设计规范
├── .gitignore
├── Makefile
├── go.mod
└── README.md
```

## 4. 配置与数据模型

### 4.1 静态配置

路径：`./config/workspaces/{workspace}.toml`。章节顺序为 `java > maven > docs > projects > databases > services > profiles`：

```toml
[java]
version = 8
home = "/path/to/java-home"

[maven]
home = "/path/to/maven-home"
repo = "/path/to/local-repo"

[docs]
dirs = ["<dir-a>", "<dir-b>"]       # 文档检索根目录列表

[projects."<project-name>"]
repo = "<git-repo-dir>"

[databases."<db-name>"]
host = "127.0.0.1"
port = 3306
db = "demo"
user = "root"
password = "***"
init-sql = "<path/to/init.sql>"   # 可选，DB reset 时执行

[services."<service-name>"]
project = "<project-name>"        # 引用 projects 中的条目
module = "<可运行 Jar 的模块>"
work-dir = "<运行 Jar 时的工作目录>"
port = 111                        # 可选，不指定则随机分配
database = "<db-name>"            # 可选，引用 databases 中的条目

[profiles]
items = ["dev", "staging"]        # workspace 支持的 profile 列表
```

层级关系：`workspace > project > service`。`databases` 为顶级命名表，被 service 通过名称引用。一个 workspace 含多 project、多 service、多 database。

### 4.2 运行时状态

路径：`./data/ws-{workspace}/state.json`，与静态配置分离。遵循"人工编辑 = TOML、程序写入 = JSON"原则：

```json
{ "current-profile": "dev" }
```

切换 profile 即更新此文件并持久化。Jungle 不解释 profile 内容，仅在运行 service 时将其作为 Spring active profile 传入。

### 4.3 数据目录约定

严格遵循 brief 并采用复数目录名：

| 路径                                         | 含义                                                |
| -------------------------------------------- | --------------------------------------------------- |
| `./data/ws-{workspace}/projects/{project}`   | 项目数据目录（`{project-dir}`）                     |
| `./data/ws-{workspace}/services/{service}`   | 服务数据目录（`{service-dir}`）                     |
| `./data/ws-{workspace}/maven-source-cache/<g 路径>/<a>/<v>/` | Maven sources.jar 解压缓存（6.4.3，`<g 路径>` = groupId 点转斜杠） |
| `{project-dir}/builds/{build-id}/stdout.log` | 构建日志（stdout）                                  |
| `{project-dir}/builds/{build-id}/stderr.log` | 构建日志（stderr）                                  |
| `{project-dir}/builds/{build-id}/state.json` | 构建状态（status/起止时间/退出码/命令等）           |
| `{project-dir}/build-id`                     | 最后一次分配的 build-id                             |
| `{service-dir}/runs/{run-id}/stdout.log`     | 运行日志（stdout）                                  |
| `{service-dir}/runs/{run-id}/stderr.log`     | 运行日志（stderr）                                  |
| `{service-dir}/runs/{run-id}/state.json`     | 运行状态（status/pid/port/profile/起止时间/退出码） |
| `{service-dir}/runs/{run-id}/pid`            | 运行进程 pid 文件                                   |
| `{service-dir}/run-id`                       | 最后一次分配的 run-id                               |

`build-id` / `run-id` 为 4 位递增序号，从 `0001` 起。分配规则：读当前值 → +1 → 写回，使用文件锁保证并发安全。不足 4 位时左补零（如 `0001`）。

### 4.4 .gitignore

`./data` 整体忽略；`./config` 下仅忽略真实配置（含敏感项），仓库保留模板：

- 忽略 `config/config.toml`（真实服务配置）
- 忽略 `config/workspaces/`（workspace 配置，含 DB 密码等）
- 忽略 `data/`（全部运行时数据）
- 保留 `config/config-sample.toml`（模板，tracked）

### 4.5 服务级配置

jungle 二进制自身的配置，路径 `./config/config.toml`（gitignored）。仓库保留模板 `config/config-sample.toml`（tracked）。字段：

```toml
[server]
addr = "127.0.0.1:7788"   # 监听地址

[data]
dir = "./data"            # 数据根目录

[log]
level = "info"            # logrus 日志级别
```

config 与 data 根目录默认相对 CWD（`./config`、`./data`），可由 CLI flag（`--config-dir`、`--data-dir`）覆盖。jungle 启动时先加载 `config.toml`，再按 `config/workspaces/*.toml` 加载各 workspace 配置。

## 5. 接入层

### 5.1 HTTP API

遵循 `docs/references/rpc-style-http-api.md`：

- HTTP method 总是 `POST`（`/api/health` 为约定例外，使用 `GET`）。
- 路径：`/api/{module}/{action}` 或 `/api/{module}/{group}/{action}`，kebab-case，参数统一在请求体。
- `Content-Type: application/json`。
- 响应统一信封：`{ success, data, error: { code, message, details? } }`，业务结果统一 `200`，`4xx`/`5xx` 仅限传输/系统层。
- 错误码 `UPPER_SNAKE_CASE`，带模块前缀。
- 分页 `data`：`{ items, page, size, total, pageCount }`。

**workspace 参数（横切约束）**：除 `/api/health` 外，所有端点请求体必须携带 `{ "workspace": "<name>" }`。由 `common` 的请求解码器统一抽取并解析对应 workspace 配置，注入后续 handler/usecase。`/api/workspace/list` 例外（其用途即枚举 workspace，不要求该字段）。

### 5.2 MCP

- 路径：`/mcp`，Streamable HTTP 传输。
- 每个 Jungle 操作映射为一个 MCP tool，tool 名由 `module(_group)_action` 派生（如 `build_run`、`service_start`、`database_reset`、`docs_search`、`maven_source_search`、`maven_source_read`）。
- tool 参数 schema = 对应 HTTP 请求体 schema（含 `workspace` 字段）。
- MCP 适配器只做 tool 调用 ↔ usecase 方法调用的编组，复用同一 `service` 实现。

### 5.3 端点清单

| 方法 | 路径                            | 关键请求参数（除 workspace） | 说明                                           |
| ---- | ------------------------------- | ---------------------------- | ---------------------------------------------- |
| GET  | `/api/health`                   | —                            | 健康检查，统一信封，后续可扩展组装状态         |
| POST | `/api/workspace/list`           | —（不需要 workspace）        | 列出所有 workspace                             |
| POST | `/api/workspace/get`            | —                            | 查看 workspace 配置 + 当前 profile             |
| POST | `/api/workspace/switch-profile` | `profile`                    | 切换当前 profile                               |
| POST | `/api/build/run`                | `project`, `opts?`           | 运行构建直到完成（同步），返回 build-id + 结果；`opts={goal,clean,test}` |
| POST | `/api/build/log/search`         | `project`, `build`, `query`, opts | 在指定构建日志中检索                           |
| POST | `/api/service/start`            | `service`                    | 后台启动 service，返回 run-id + port           |
| POST | `/api/service/stop`             | `service`                    | 停止 service 的活跃 run                        |
| POST | `/api/service/status`           | `service`                    | 查询 service 活跃 run 的状态                   |
| POST | `/api/run/log/search`           | `service`, `run`, `query`, opts | 在指定运行日志中检索                           |
| POST | `/api/database/reset`           | `database`                   | 执行对应 database 的 `init-sql`                |
| POST | `/api/database/query`           | `database`, `sql`, `opts?`   | 执行 SQL，返回 columns + rows；`opts.maxRows` 默认 200 截断 |
| POST | `/api/docs/search`              | `query`, opts                | 在配置的文档目录中检索                         |
| POST | `/api/maven/source/search`      | `dependency?`, `project?`, `query`, opts | 在 sources.jar 中检索；范围优先级 dependency > project > global |
| POST | `/api/maven/source/read`        | `dependency?`, `project?`, (`file`\|`class`), `range?` | 按精确文件名/类名读取源文件，全文或行范围；多义返回候选 |

> 说明：`build` = build-id，`run` = run-id，`service`/`project`/`database` 为配置中的名称；`dependency` 为 Maven 坐标 `groupId:artifactId:version`（可选）；`project` 可选，用于圈定项目依赖集合；`file` 为源码树内相对路径（如 `com/example/Foo.java`），`class` 为类名（简单名 `Foo` 或 FQCN `com.example.Foo`），`range` = `{start, end}` 1-based 行范围。

## 6. 功能模块数据流

### 6.1 Build（构建）

`BuildService.Run(workspace, project, opts)`（同步，运行直到完成）。`BuildOpts`：

- `goal`（string，默认 `package`）：Maven phase/goal，自由字符串（如 `compile`/`package`/`install`/`verify`）。
- `clean`（bool，默认 true）：前置 `clean`。
- `test`（bool，默认 false）：false 时加 `-DskipTests`。

执行 `mvn [clean] <goal> [-DskipTests]`。

1. 从 config 解析 `maven.home` / `maven.repo` 与 project 的 `repo` 目录。
2. 分配 `build-id`，创建 `{project-dir}/builds/{id}/`。
3. 写 `state.json`（status=running、开始时间、命令）。
4. `infra/maven` 在 repo 目录执行上述 `mvn` 命令，stdout/stderr 流式写入 `stdout.log` / `stderr.log`。
5. 构建结束，更新 `state.json`（status=success/failed、结束时间、退出码）。
6. 返回 `build-id` + 状态。

`BuildService.LogSearch(workspace, project, build, query, opts)`：以 `{build-dir}` 的 `stdout.log`+`stderr.log` 为 roots 调用共享 `Searcher`（6.4.1），返回 `SearchResult`；`lineContent` 默认剥离 ANSI（`opts.raw` 保留）。

### 6.2 Service / Run（服务运行）

`RunService.Start(workspace, service)`（异步，后台）：

1. 解析 `java.home`、service 的 `work-dir`、`port`。
2. **解析可运行 jar**：service 的 `module` 为项目 repo 内的模块目录；`infra/maven` 在该模块执行 `mvn help:evaluate -Dexpression=project.build.directory` 与 `project.build.finalName`，得到 `<target>/<finalName>.jar`（Spring Boot repackaged 后的原位 jar，`original-*` 留作备份）。jar 不存在视为未构建，返回 `RUN_FAILED`（提示先 `build/run`）。
3. **端口分配**：`port` 未配置时，jungle 绑定 TCP `:0` 取系统分配端口后关闭 listener，作为该 run 的端口（存在轻微竞态，可接受）。
4. 从 `state.json` 读取当前 profile。
5. 分配 `run-id`，创建 `{service-dir}/runs/{id}/`。
6. 写 `state.json`（status=running、pid、port、profile、开始时间）。
7. `infra/process` 后台启动：`java -jar <module-jar> --spring.profiles.active=<profile> --server.port=<port>`，工作目录为 `work-dir`；pid 落 `pid` 文件；stdout/stderr 流式写日志。
8. 立即返回 `run-id` + 端口。
9. 进程退出时（正常或崩溃）更新 `state.json`（status=exited/crashed、结束时间、退出码）。

> "活跃 run" = 该 service 下 `run-id` 最大的 run（不论状态）。`start` 不拒绝并发启动（每次新建 run）；若 service 配置了固定 `port` 且已被占用，进程启动失败 → `RUN_FAILED`。jungle 重启后不回收孤儿进程（pid 文件残留仅供查阅），MVP 不做 reattach。

`RunService.Stop(workspace, service)`：定位活跃 run（`run-id` 最大者）；若其 `state.json` status=running 则按 pid 终止进程并更新 `state.json`（status=stopped、结束时间），否则返回 `SERVICE_NOT_RUNNING`。

`RunService.Status(workspace, service)`：读取活跃 run（`run-id` 最大者）的 `state.json`，返回 status/pid/port/profile/起止时间/退出码。

`RunService.LogSearch(workspace, service, run, query, opts)`：以 `{run-dir}` 的 `stdout.log`+`stderr.log` 为 roots 调用共享 `Searcher`（6.4.1），返回 `SearchResult`；`lineContent` 默认剥离 ANSI（`opts.raw` 保留）。

### 6.3 Database（数据库）

`DatabaseService.Reset(workspace, database)`：

1. 读 `databases.<database>` 连接信息与 `init-sql` 路径。
2. `infra/mysql` 连接目标 DB。
3. **执行 `init-sql`**。Jungle 不主动清库，幂等性由 SQL 自身保证（建议 SQL 内含 `DROP TABLE IF EXISTS` / `TRUNCATE` 等）。`infra/mysql` 连接开启 `multiStatements=true` 以执行多语句文件（或按 `;` 拆分逐条执行）。
4. 返回执行结果。

`DatabaseService.Query(workspace, database, sql, opts)`：执行 SQL，返回 `columns` + `rows`；`opts.maxRows` 默认 200 截断以保护 AI 上下文（`truncated` 标识是否截断）。

### 6.4 Search（检索）

#### 6.4.1 共享 Searcher 契约

统一 `Searcher` Port，四个检索端点共用：

```go
Search(ctx, roots []string, query string, opts SearchOpts) (*SearchResult, error)
```

- **底层原语**：优先 shell out `rg --json`（解析 match/context 记录）；运行时检测不到 `rg` 则回退到 Go `bufio`+`regexp` 实现（复刻核心 opts：literal/ignoreCase/context/maxCount）。
- **`SearchOpts`**：`literal`（true→`--fixed-strings`）、`ignoreCase`、`glob`/`type`（文件过滤，主要用于 docs）、`context`（{A,B,C} 上下文行）、`maxCount`、`raw`（默认 false，见 ANSI 处理）。
- **`Match`**：`{ file, line, lineContent, type }`，`type` ∈ `match`/`context`。`file` 为相对其检索根的路径（日志=`stdout.log`/`stderr.log`，docs=相对目录，maven=相对缓存目录即 jar 内路径）。
- **`SearchResult`**（截断式，非分页）：`{ items: []Match, total: int, truncated: bool }`。`maxCount` 默认 200、硬上限 1000，仅 `match` 行计入 `maxCount`（`context` 行不计）。触达上限时 rg 被中断、真实总数未知，故 `total` = 已返回条数、`truncated=true` 表示"可能还有更多"。
- **ANSI 处理**：对所有检索默认正则剥离 `lineContent` 中的 ANSI 颜色码（对无 ANSI 文本为无害空操作），`opts.raw=true` 保留原文。搜索仍对原文进行（查询词一般不在 ANSI 码内，不受影响）。
- 调用使用 `exec.CommandContext`，随 ctx 取消/超时终止 `rg`。

#### 6.4.2 文档检索 `SearchService.Docs(workspace, query, opts)`

1. 从 `docs.dirs` 解析多个文档根目录。
2. `Searcher.Search(ctx, dirs, query, opts)`（rg 支持多根一次调用，递归）。
3. 返回 `SearchResult`。`glob`/`type` 用于按扩展名/类型过滤（如仅 `.md`）。

#### 6.4.3 Maven 源码检索 `SearchService.MavenSource(workspace, dependency, project, query, opts)`

采用"按范围圈定 → 确保 sources.jar 解压入缓存 → rg"路径。范围优先级 `dependency > project > global`（给定 `dependency` 时 `project` 被忽略）：

**范围解析 → roots**（共享子流程 `ensureSourcesExtracted`，返回待检索的缓存子目录列表）：

> 并发控制：`ensureSourcesExtracted` 对每个坐标（`g:a:v`）取 per-coord 文件锁，并发请求遇到他人正在解压则**等待**而非重复解压；已缓存则直接放行。

- **`dependency` 给定**（`g:a:v`）：单个 artifact。
  - 缓存目录 `./data/ws-{workspace}/maven-source-cache/<g 路径>/<a>/<v>/` 已有解压结果 → 命中。
  - 否则：优先在 `maven.repo` 本地仓库查 `<a>-<v>-sources.jar`；缺失则 `mvn dependency:copy -Dartifact=<g>:<a>:<v>:jar:sources` 下载（读 `settings.xml` 镜像/认证/私服）。
  - 解压到上述缓存目录。
  - roots = [该缓存目录]。
- **`project` 给定**（无 `dependency`）：该项目解析出的依赖集合。
  - 在 project 的 `repo` 目录执行 `mvn dependency:list -DoutputFile=...` 取解析后的依赖清单 `g:a:v`（含传递依赖、已解析版本、exclusions）。
  - 缺失 sources.jar 的，一次性 `mvn dependency:sources` 批量拉取（单次 JVM）。
  - 逐个解压进各自的缓存子目录（已缓存跳过）。
  - roots = 这些缓存子目录的并集。
- **都不给**：全本地仓库。
  - 扫描 `maven.repo` 下所有 `*-sources.jar`，缺失解压者惰性解压进缓存。
  - roots = [整个 `maven-source-cache/` 根]。

**缓存失效策略**（适用以上所有）：release 版本永不失效；`-SNAPSHOT` 版本校验远程 `maven-metadata.xml`，更新则重抓重解压。

**检索**：

1. `ensureSourcesExtracted(workspace, scope)` → roots。
2. `Searcher.Search(ctx, roots, query, opts)`（rg，支持多根）。
3. 返回 `SearchResult`。

#### 6.4.4 Maven 源码读取 `SearchService.MavenRead(workspace, dependency, project, file, class, range)`

按精确文件名/类名读取源文件，全文或行范围。范围解析与 6.4.3 共享 `ensureSourcesExtracted`（同三档优先级）。

1. `ensureSourcesExtracted(workspace, scope)` → roots。
2. **目标解析**：
   - `file` 给定：源码树内相对路径（如 `com/example/Foo.java`），在 roots 中精确匹配。
   - 否则 `class` 给定：FQCN `com.example.Foo` → `com/example/Foo.java`；简单名 `Foo` → 匹配 `<Foo>.java`。
3. **匹配判定**：
   - 唯一命中 → 读取内容；`range={start,end}`（1-based）给定时只返回该行范围。
   - 多个命中 → 错误 `AMBIGUOUS_TARGET`，`details` 列出候选 `[{ dependency, file }]`，调用方加 `dependency` 消歧。
   - 无命中 → `NOT_FOUND`。
4. 返回 `{ file, content, startLine, endLine, totalLines }`（`range` 缺省时 startLine=1、endLine=totalLines）。

## 7. 错误处理

- 错误模型遵循响应信封，`internal/apperrors`（无依赖叶子包）提供类型化错误与稳定错误码，被 `service`/`infra`/`common`/`handlers` 共用。
- 错误码（`UPPER_SNAKE_CASE`，带模块前缀）：
  - `VALIDATION_ERROR`（配合 `details` 给字段级错误）
  - `BUILD_FAILED`、`RUN_FAILED`、`SERVICE_NOT_RUNNING`、`DB_QUERY_FAILED`、`DB_RESET_FAILED`、`SEARCH_FAILED`、`AMBIGUOUS_TARGET`
  - `NOT_FOUND`（workspace/project/service/database/build/run/dependency 缺失）
  - `INTERNAL_ERROR`（系统层，不向终端暴露细节）
- 校验：请求 struct + `go-playground/validator`，失败 → `VALIDATION_ERROR` + `details`。
- usecase 返回类型化 error，`common` 的错误映射器统一转信封。
- `code` 稳定不变，客户端基于 `code` 而非 `message` 判断；`message` 面向用户，不含敏感信息。

## 8. 测试策略

- usecase 层（`internal/service/*`）：表驱动单测 + fake Port 实现，目标全覆盖。
- infra 层：使用 `t.TempDir()`；依赖真实二进制（`mvn`/`mysql`/`rg`）的用例用 build tag/guard 保护，环境缺失时跳过。
- handlers 层：对信封、错误映射、路由挂载、workspace 参数解析做轻量断言；端到端用 `httptest`。
- 测试数据置于 `testdata/`。

## 9. 脚手架范围（基础结构交付物）

经本规范批准后，通过 wf-writing-plans → 执行落地。脚手架包含：

- `go.mod`（Echo、TOML 解析、`go-playground/validator`、MCP Go SDK、logrus）。
- 第 3 节目录树，各包以 `doc.go` 占位说明职责。
- `internal/config`：TOML structs + loader（服务级 `config.toml` 的 server/data/log + workspace 的 java/maven/docs/projects/databases/services/profiles）。
- `config/config-sample.toml`：服务级配置模板（tracked）。
- `internal/storage`：数据目录路径助手 + `build-id`/`run-id` 分配器（文件锁）+ `state.json` 读写。
- `internal/common`：响应信封 helper、错误映射、请求解码器（含 workspace 参数解析）。
- `internal/handlers`：Echo 路由挂载、`/api/health`、workspace 基线端点（最小实现）。
- `internal/mcp`：Streamable HTTP 挂载骨架 + tool 注册表（后续阶段接 usecase）。
- `internal/service`：各模块 usecase 接口 + 骨架 struct；`workspace` 最小实现，其余返回 `ErrNotImplemented`。
- `internal/infra`：`maven`/`mysql`/`process`/`searcher` 的接口 + stub 实现。
- `internal/app`：组件初始化、静态 DI、HTTP Server 初始化（组装根）。
- `cmd/jungle/main.go`：解析参数，调用 `app` 启动。
- `.gitignore`：忽略 `config/config.toml`、`config/workspaces/`、`data/`；保留 `config/config-sample.toml`。
- Makefile：`build`/`run`/`test`/`vet`/`fmt`。
- lint 配置、README 骨架、AGENTS.md 补充 lint/test 命令。

**不在脚手架范围**（后续阶段）：build/run/database/search 的完整实现、全部 MCP tool 接线、sources.jar 下载与解压的完整实现。
