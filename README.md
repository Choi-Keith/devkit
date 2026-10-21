# devkit

English | [中文](README-zh.md)

A project scaffold generator. Ships with several backend project templates and creates new projects with a single command. Templates are not embedded in the binary — they are downloaded from the template repository at generation time, so they always stay in sync with the repo.

## Templates

| Name | Language | Description |
| --- | --- | --- |
| `golang-ddd` | Go | Go DDD layered architecture template (configs / deployments / internal layers) |
| `golang-feature-mvc` | Go | Go Feature MVC template (feature-sliced modules + GORM + Redis + JWT + Swagger) |
| `golang-simple` | Go | Lightweight Go monolith service template |
| `nestjs-simple` | TypeScript | NestJS 11 template (TypeORM + Redis + JWT + RBAC) |
| `python-fastapi` | Python | FastAPI template (Alembic migrations + uv dependency management) |

## Installation

Requires Go 1.25+.

Option 1: install directly with `go install`:

```bash
go install github.com/Choi-Keith/devkit/cmd/devkit@latest
```

Option 2: build from source:

```bash
git clone https://github.com/Choi-Keith/devkit.git
cd devkit/cmd/devkit
make install   # go install into $GOBIN
# or
make build     # outputs ./bin/devkit
```

## Usage

```bash
# List available templates
devkit list

# Create interactively
devkit new

# Create non-interactively
devkit new golang-ddd my-service --module github.com/acme/my-service
```

Common flags for `devkit new`:

| Flag | Description |
| --- | --- |
| `-d, --dir` | Parent directory where the project is created (default: current directory) |
| `--module` | Rewrite the Go module path and internal imports (Go templates only) |
| `--force` | Allow writing into a non-empty directory |
| `--repo` | Custom template repository URL |
| `--ref` | Git branch or tag to download templates from |
| `--local` | Use a local template directory instead of downloading |

## After Generating

```bash
# Go templates
cd my-service && go mod tidy && go run ./cmd/...

# golang-feature-mvc needs MySQL + Redis; start them first
cd my-service && docker compose -f deployments/docker-compose.yaml up -d && make run

# NestJS template
cd my-service && pnpm install && pnpm start:dev

# FastAPI template
cd my-service && uv sync && uv run python -m src.main
```

## Repository Layout

```
.
├── cmd/devkit/                  # devkit CLI source
├── golang-ddd-template/         # Go DDD template
├── golang-feature-mvc-template/ # Go Feature MVC template
├── golang-simple-template/      # Go monolith template
├── nestjs-simple-template/      # NestJS template
└── python-fastapi-template/     # FastAPI template
```

To add a template: create a template directory at the repository root and register it in `All` inside `cmd/devkit/internal/templates/templates.go`.

## License

[MIT](LICENSE)
