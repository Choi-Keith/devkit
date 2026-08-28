# golang-simple-template

基于 [golang-common](https://github.com/Choi-Keith/golang-common) 的 Go HTTP 服务模板,采用 `router → handler → service → repository → model` 分层与依赖注入容器。

## 目录结构

```
├── cmd/server/         # 服务入口
├── config/             # 配置文件(config.example.yaml 为模板)
├── internal/
│   ├── config/         # 配置加载
│   ├── container/      # 依赖注入容器
│   ├── ecode/          # 业务错误码(从 20000 开始)
│   ├── model/          # GORM 模型
│   ├── repository/     # 数据访问层
│   ├── service/        # 业务逻辑层
│   ├── handler/        # HTTP 处理器
│   ├── router/         # 路由注册
│   └── types/          # 请求/响应 DTO
├── Dockerfile
└── Makefile
```

## 快速开始

```bash
# 1. 准备配置
cp config/config.example.yaml config/config.yaml

# 2. 启动依赖的 PostgreSQL 与 Redis,并修改 config.yaml 中的连接信息

# 3. 运行
make run
```

启动后:

- 健康检查:`GET /health`、`/health/live`、`/health/ready`
- 用户 API:`/api/v1/users`(增删改查、分页)

```bash
curl -X POST localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.com","name":"alice"}'
```

## 作为新项目模板使用

1. 复制本目录,将 `go.mod` 的 module 名和代码中的 `golang-simple-template` 导入路径批量替换为你的模块名。
2. 按需扩展:`internal/model`、`internal/repository`、`internal/service`、`internal/handler` 各加一个文件,然后在 `container.go` 和 `router.go` 中装配。
3. 业务错误码在 `internal/ecode/` 中注册,从 20000 开始;统一响应格式为 `{code, message, data}`。

更多组件能力(JWT 认证、限流、加解密、缓存操作等)见 golang-common 各包的 README。
