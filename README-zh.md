# devkit

[English](README.md) | 中文

一个项目脚手架生成工具。内置多套后端项目模板，一条命令即可创建新项目；模板不打包进二进制，而是在生成时从模板仓库下载，始终与仓库保持同步。

## 模板一览

| 名称 | 语言 | 说明 |
| --- | --- | --- |
| `golang-ddd` | Go | Go DDD 分层架构模板（configs / deployments / internal 分层） |
| `golang-feature-mvc` | Go | Go Feature MVC 模板（按 feature 纵向切分 + GORM + Redis + JWT + Swagger） |
| `golang-simple` | Go | Go 轻量级单体服务模板 |
| `nestjs-simple` | TypeScript | NestJS 11 模板（TypeORM + Redis + JWT + RBAC） |
| `python-fastapi` | Python | FastAPI 模板（Alembic 迁移 + uv 依赖管理） |

## 安装

需要 Go 1.25+。

方式一：直接使用 `go install`：

```bash
go install github.com/Choi-Keith/devkit/cmd/devkit@latest
```

方式二：从源码构建：

```bash
git clone https://github.com/Choi-Keith/devkit.git
cd devkit/cmd/devkit
make install   # go install,装入 $GOBIN
# 或
make build     # 生成 ./bin/devkit
```

## 使用

```bash
# 查看可用模板
devkit list

# 交互式创建
devkit new

# 非交互式创建
devkit new golang-ddd my-service --module github.com/acme/my-service
```

`devkit new` 常用参数：

| 参数 | 说明 |
| --- | --- |
| `-d, --dir` | 项目创建的父目录(默认当前目录) |
| `--module` | 重写 Go module 路径及内部 import(仅 Go 模板) |
| `--force` | 允许写入非空目录 |
| `--repo` | 自定义模板仓库地址 |
| `--ref` | 指定模板仓库的分支或 tag |
| `--local` | 使用本地模板目录,不联网下载 |

## 生成后

```bash
# Go 模板
cd my-service && go mod tidy && go run ./cmd/...

# golang-feature-mvc 依赖 MySQL + Redis,先启动依赖
cd my-service && docker compose -f deployments/docker-compose.yaml up -d && make run

# NestJS 模板
cd my-service && pnpm install && pnpm start:dev

# FastAPI 模板
cd my-service && uv sync && uv run python -m src.main
```

## 仓库结构

```
.
├── cmd/devkit/                  # devkit CLI 源码
├── golang-ddd-template/         # Go DDD 模板
├── golang-feature-mvc-template/ # Go Feature MVC 模板
├── golang-simple-template/      # Go 单体模板
├── nestjs-simple-template/      # NestJS 模板
└── python-fastapi-template/     # FastAPI 模板
```

新增模板:在仓库根目录新建模板目录,并在 `cmd/devkit/internal/templates/templates.go` 的 `All` 中注册即可。

## License

[MIT](LICENSE)
