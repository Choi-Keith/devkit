.PHONY: build run test lint fmt tidy clean docker-build help

BINARY_NAME=server
BUILD_DIR=bin
GO=go

build: ## 构建二进制
	$(GO) build -v -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: ## 本地运行
	$(GO) run ./cmd/server -config config/config.yaml

test: ## 运行测试
	$(GO) test -v -race -coverprofile=coverage.out ./...

lint: ## 静态检查
	golangci-lint run ./...

fmt: ## 格式化代码
	$(GO) fmt ./...

tidy: ## 整理依赖
	$(GO) mod tidy

clean: ## 清理构建产物
	rm -rf $(BUILD_DIR) coverage.out

docker-build: ## 构建 Docker 镜像
	docker build -t $(BINARY_NAME):latest .

help: ## 查看所有目标
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
