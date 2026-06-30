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
