# golang-ddd-template

基于 [golang-common](https://github.com/Choi-Keith/golang-common) 的 Go DDD 脚手架，内置用户（注册 / 查询 / 分页 / 删除）与订单（创建 / 查询 / 分页 / 取消）模块示例、Swagger 文档、健康检查与优雅退出。

## 目录结构

```text
├── cmd/server/             # 入口:加载配置 + 依赖注入 + 启动
├── configs/                # 配置文件
├── docs/                   # swag 生成的 Swagger 文档
├── internal/
│   ├── domain/             # 领域层:user、order 聚合(实体、不变量、仓储接口,零第三方依赖)
│   ├── application/        # 应用层:user、order 用例编排与 DTO(包名 userapp、orderapp)
│   ├── adapter/http/       # HTTP 输入适配器:user、order 子目录(handler、dto、ecode,包名 userhttp、orderhttp)
│   ├── adapter/grpc/       # gRPC 输入适配器(预留):同样按业务模块分目录,如 grpc/user/server.go,包名 usergrpc
│   └── infra/db/           # 基础设施:user、order 子目录,GORM 仓储实现(包名 userdb、orderdb)
├── deployments/            # docker-compose(MySQL)
├── Makefile
└── Dockerfile
```

依赖方向：`adapter/infra → application → domain`，仓储接口定义在 domain，实现在 infra，组装在 `cmd/server/main.go`。

## 命名约定

- 每一层都按业务模块分目录（`http/user`、`application/user`、`infra/db/user`)，目录名即模块名。
- 固定包名（等价于固定导入别名，import 时无需再写别名）:
  - `internal/application/<m>` → `<m>app`（如 `userapp`)
  - `internal/adapter/http/<m>` → `<m>http`（如 `userhttp`)
  - `internal/adapter/grpc/<m>` → `<m>grpc`（如 `usergrpc`)
  - `internal/infra/db/<m>` → `<m>db`（如 `userdb`)
  - `internal/domain/<m>` → `<m>`（领域包保持模块名，引用时不加别名）
- 文件名不带模块名或角色后缀：同包内直接叫 `handler.go`、`server.go`、`service.go`、`mysql.go`、`po.go`、`dto.go`、`ecode.go`，不要 `user_handler.go`、`mysql_user_repo.go` 这类冗余命名。

## 快速开始

```bash
# 1. 启动 MySQL(或修改 configs/config.yaml 指向已有实例)
docker compose -f deployments/docker-compose.yaml up -d

# 2. 建表(首次),主键为 bigint unsigned 自增
docker exec -i golang-ddd-template-mysql mysql -uroot -proot app -e \
  "CREATE TABLE IF NOT EXISTS users (id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY, email varchar(255) NOT NULL UNIQUE, name varchar(128) NOT NULL, password varchar(255) NOT NULL, created_at timestamp NOT NULL, updated_at timestamp NOT NULL);"
docker exec -i golang-ddd-template-mysql mysql -uroot -proot app -e \
  "CREATE TABLE IF NOT EXISTS orders (id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY, user_id bigint unsigned NOT NULL, order_no varchar(64) NOT NULL UNIQUE, amount bigint NOT NULL, status varchar(16) NOT NULL, created_at timestamp NOT NULL, updated_at timestamp NOT NULL, INDEX idx_orders_user_id (user_id), INDEX idx_orders_status (status));"

# 3. 运行
make run
```

- API 地址：`http://localhost:8081/api/v1`
- Swagger UI：`http://localhost:8081/swagger/index.html`
- 健康检查：`GET /health`、`GET /health/liveness`、`GET /health/readiness`

## API 约定

- 统一响应包络（由 golang-common/response 提供）：

```json
{ "code": 0, "message": "success", "data": {} }
```

- JSON tag 一律小驼峰（`userName` 而非 `user_name`)。
- 分页参数：`page` / `pageSize`(pageSize 上限 100)。
- 业务错误码从 20000 起，在每个模块的 `internal/adapter/http/<m>/ecode.go` 注册（user 从 20000、order 从 20100);1000–17999 为 golang-common 内置通用码。

## 常用命令

```bash
make build    # 构建到 bin/server(注入版本号)
make test     # go test -race -coverprofile
make lint     # golangci-lint run
make swagger  # 重新生成 docs/(修改注解后执行)
make run      # 本地运行
```

## 新增一个业务模块

以 `product` 为例（用户、订单模块即按此模式实现）：

1. `internal/domain/product/`：实体 + 仓储接口 + 领域错误（包名 `product`);
2. `internal/application/product/`：用例服务 + DTO（包名 `productapp`,文件如 `service.go`、`dto.go`);
3. `internal/infra/db/product/`：PO 与 GORM 仓储实现（包名 `productdb`,文件如 `mysql.go`、`po.go`);
4. `internal/adapter/http/product/`：handler、DTO 与业务错误码（包名 `producthttp`,文件如 `handler.go`、`dto.go`、`ecode.go`)，并在 `internal/adapter/http/router.go` 注册路由；
5. `cmd/server/main.go` 中组装注入；
6. `make swagger && make test`。
