---
date: 2026-07-01
---

# Workspace API Design

- 日期: 2026-07-01
- 状态: 草案

## 约定

- 方法: 三个端点均为 `POST`。
- Content-Type: `application/json`。
- 路径前缀: `/api/workspace`，kebab-case。
- 响应信封: `{ success, data, error: { code, message, details? } }`。业务成功/失败统一 HTTP `200`；`4xx`/`5xx` 仅限传输/系统层（404→`NOT_FOUND`、405→`METHOD_NOT_ALLOWED`、500→`INTERNAL_ERROR`）。
- `data` 属性与请求体属性均使用 **camelCase**。
- 请求体需携带 `workspace` 字段；**`/api/workspace/list` 例外**（其用途即枚举 workspace，不要求该字段）。

## 清单

### 1. 列出 workspace — `/api/workspace/list`

- 方法: `POST`
- 路径: `/api/workspace/list`
- 说明: 枚举 `config/workspaces/` 目录下所有 `*.toml` 的文件名（去除 `.toml` 后缀），按目录返回顺序。不要求请求体携带 `workspace` 字段，请求体被忽略。

请求消息:

```jsonc
{ }
```

成功响应 `data`:

`data` 直接为 workspace 名称字符串数组（去 `.toml` 后缀）。无 workspace 时为空数组。

```jsonc
{
  "success": true,
  "data": ["demo", "staging"],
  "error": null
}
```

错误情形:

| 错误码           | HTTP | 触发条件                                  |
| ---------------- | ---- | ----------------------------------------- |
| `NOT_FOUND`      | 200  | `config/workspaces/` 目录不存在或不可读。 |
| `INTERNAL_ERROR` | 500  | 读目录时发生未预期系统错误。              |

### 2. 查看 workspace — `/api/workspace/get`

- 方法: `POST`
- 路径: `/api/workspace/get`
- 说明: 返回指定 workspace 的静态配置与当前 profile。`currentProfile` 取自 `data/ws-{workspace}/state.json`；该文件不存在（从未切换过 profile）时为空字符串。

请求消息:

| 字段        | 类型   | 必填 | 说明             |
| ----------- | ------ | ---- | ---------------- |
| `workspace` | string | 是   | workspace 名称。 |

```jsonc
{ "workspace": "demo" }
```

成功响应 `data`（workspace 视图，字段 camelCase）:

| 字段             | 类型   | 说明                                                                                                              |
| ---------------- | ------ | ----------------------------------------------------------------------------------------------------------------- |
| `name`           | string | workspace 名称（= 请求中的 `workspace`）。                                                                        |
| `currentProfile` | string | 当前 profile；state.json 缺失时为 `""`。                                                                          |
| `java`           | object | `{ version: int, home: string }`。                                                                                |
| `maven`          | object | `{ home: string, repo: string }`。                                                                                |
| `docs`           | object | `{ dirs: string[] }`。                                                                                            |
| `projects`       | object | 键为 project 名称，值为 `{ repo: string }`。                                                                      |
| `databases`      | object | 键为 database 名称，值为 `{ host, port, db, user, initSql }`。**不含 `password`**（脱敏）。`initSql` 缺省时省略。 |
| `services`       | object | 键为 service 名称，值为 `{ project, module, workDir, port, database }`。`port`/`database` 缺省时省略。            |
| `profiles`       | object | `{ items: string[] }`。                                                                                           |

```jsonc
{
  "success": true,
  "data": {
    "name": "demo",
    "currentProfile": "dev",
    "java":   { "version": 8, "home": "/path/to/java-home" },
    "maven":  { "home": "/path/to/maven-home", "repo": "/path/to/local-repo" },
    "docs":   { "dirs": ["<dir-a>", "<dir-b>"] },
    "projects": {
      "<project-name>": { "repo": "<git-repo-dir>" }
    },
    "databases": {
      "<db-name>": { "host": "127.0.0.1", "port": 3306, "db": "demo", "user": "root", "initSql": "<path/to/init.sql>" }
    },
    "services": {
      "<service-name>": { "project": "<project-name>", "module": "<module>", "workDir": "<work-dir>", "port": 111, "database": "<db-name>" }
    },
    "profiles": { "items": ["dev", "staging"] }
  },
  "error": null
}
```

错误情形:

| 错误码             | HTTP | 触发条件                                                             |
| ------------------ | ---- | -------------------------------------------------------------------- |
| `VALIDATION_ERROR` | 200  | 请求体缺 `workspace` 或 `workspace` 为空（`details` 给字段级错误）。 |
| `NOT_FOUND`        | 200  | `config/workspaces/<workspace>.toml` 不存在或解析失败。              |

`VALIDATION_ERROR` 示例:

```jsonc
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": [
      { "code": "required", "message": "field workspace failed required", "target": "workspace" }
    ]
  }
}
```

### 3. 切换 profile — `/api/workspace/switch-profile`

- 方法: `POST`
- 路径: `/api/workspace/switch-profile`
- 说明: 将 `data/ws-{workspace}/state.json` 的 `current-profile` 更新为请求值并持久化，随后返回新的当前 profile。Jungle 不解释 profile 内容: **不校验 `profile` 是否属于 `profiles.items`**，接受任意非空字符串。

请求消息:

| 字段        | 类型   | 必填 | 说明                                  |
| ----------- | ------ | ---- | ------------------------------------- |
| `workspace` | string | 是   | workspace 名称。                      |
| `profile`   | string | 是   | 目标 profile 名称（任意非空字符串）。 |

```jsonc
{ "workspace": "demo", "profile": "staging" }
```

成功响应 `data`:

| 字段             | 类型   | 说明                                           |
| ---------------- | ------ | ---------------------------------------------- |
| `workspace`      | string | workspace 名称（回显请求值）。                 |
| `currentProfile` | string | 切换后的当前 profile（回显请求的 `profile`）。 |

```jsonc
{
  "success": true,
  "data": {
    "workspace": "demo",
    "currentProfile": "staging"
  },
  "error": null
}
```

错误情形:

| 错误码             | HTTP | 触发条件                                                                      |
| ------------------ | ---- | ----------------------------------------------------------------------------- |
| `VALIDATION_ERROR` | 200  | 请求体缺 `workspace` 或 `profile`，或二者任一为空（`details` 给字段级错误）。 |
| `NOT_FOUND`        | 200  | `config/workspaces/<workspace>.toml` 不存在或解析失败。                       |
| `INTERNAL_ERROR`   | 500  | 写 `state.json` 失败（目录创建/写盘错误）。                                   |

## 验收标准

- [ ] 覆盖 `list`/`get`/`switch-profile` 三端点的 Path、请求 schema、成功 `data` schema 与示例。
- [ ] `list` 的 `data` 直接为字符串数组，无 `items` 包装层。
- [ ] `get` 响应 `data` 字段命名为 camelCase，`databases` 条目不含 `password`。
- [ ] 每个端点列出触发 `VALIDATION_ERROR`/`NOT_FOUND` 等业务错误码的条件。
