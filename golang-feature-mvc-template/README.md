# golang-feature-mvc-template

基于 [golang-common](https://github.com/Choi-Keith/golang-common) 的 Go Web 脚手架,按 **feature 纵向切分**:`internal/<feature>/` 承载接口层与应用层,`internal/infra/` 集中承载基础设施。内置 user(注册 / 登录 / 令牌刷新与吊销 / CRUD + 分页 + 缓存)与 order(创建 / 查询 / 分页 / 取消,全量 JWT 保护)两个示例模块,以及 Swagger 文档、健康检查与优雅退出。

## 目录结构

```text
├── cmd/server/                       # 入口:加载配置 + 手工装配依赖 + 启动(无 DI 框架)
├── configs/                          # 配置结构(config.go)与配置文件(config.yaml / config.example.yaml)
├── docs/                             # swag 生成的 Swagger 文档(make swagger)
├── doc.go                            # 根占位包:让 swag 的 go list 能解析模块路径,注解里的 import 别名引用(如 pkgdto.Resp[...])才有效
├── deployments/                      # docker-compose(MySQL + Redis)
├── pkg/                              # 对外可导出的公共面(internal/ 是内部实现,不可被外部 import)
│   ├── dto/                          #   共享 DTO:泛型包络 Resp[T]/ListResp[T]/ErrorResp、分页 PageReq 与 NormalizePage
│   └── xerr/                         #   全部业务错误码的唯一定义处(*ecode.Error,按 feature 分区)
└── internal/
    ├── user/                         # feature 包:接口层 + 应用层
    │   ├── router.go                 #   根包 user:只承载路由注册
    │   ├── handler/                  #   包 handler:HTTP 绑定 + swagger 注解(错误原样交 response.Error)
    │   ├── service/                  #   包 service:业务逻辑,定义 Repository/Cache 端口,返回 pkg/xerr 业务错误码
    │   └── dto/                      #   包 dto:请求/响应 DTO(命名约定 Req/Resp 后缀)
    ├── order/                        # feature 包:同构(handler/service/dto + router.go)
    └── infra/                        # 集中式基础设施,只被 main 装配,不 import 任何 feature 包
        ├── cache/                    #   包 cache:Redis 缓存适配(实现 service 定义的 Cache 端口)
        └── db/<feature>/
            ├── model/                #   包 model:GORM 模型(贫血模型直通 service/handler)
            └── repository/           #   包 repository:GORM 数据访问实现(实现 service 定义的 Repository 端口)
```

### 结构规则:什么进 feature、什么进 infra

- **`internal/<feature>/` 放"业务知道什么"**:路由注册(根包)、HTTP 绑定与错误映射(handler)、业务逻辑与端口接口(service)、请求/响应 DTO(dto)。
- **`internal/infra/` 放"怎么存取"**:GORM model、repository 实现、缓存适配。infra **不 import 任何 feature 包**,单向地被 `cmd/server/main.go` 装配进 service。
- **`pkg/` 放对外可导出的公共面**(`internal/` 是内部实现,Go 工具链禁止外部模块 import):跨 feature 共享的纯类型/工具,如 `pkg/dto` 的泛型包络与分页类型、`pkg/xerr` 的业务错误码。零业务语义的共享物才放这里;有业务语义的共享能力请建独立 feature。
- **接口定义在消费方**:`Repository` / `Cache` 接口定义在各自的 service 包;infra 只提供结构满足接口的具体实现(构造返回具体类型,如 `repository.New(db) *Repository`),由 main 传入 `service.NewService`。
- **业务错误码集中在 `pkg/xerr/error.go`**:全部 feature 的业务码(*ecode.Error)定义在这一个文件里,按 feature 分区、每区 100 个号段;service 业务失败直接返回,handler 不调映射函数,错误原样交 `response.Error`(内部 `ecode.FromError` 识别 *ecode.Error 并保留 code/HTTP status)。
- **错误词汇分层**:repository(infra)不引用业务码,按接口约定只返回 `gorm.ErrRecordNotFound` / `gorm.ErrDuplicatedKey`;service 的 `mapRepoErr` 负责翻译成 pkg/xerr 错误;参数绑定失败在 handler 直接用内置 `ecode.ErrInvalidParams`。
- **model 直通**:本模板有意合并 model/PO/domain,service 直接返回 `infra/db/<feature>/model` 的贫血模型,dto 转换留在 handler。领域复杂度高(不变量、聚合、事件)请改用 golang-ddd-template。
- **子包全部用短名**:`handler` / `service` / `dto` / `model` / `repository` / `cache`;跨包引用重名时在 import 处起别名(如 main.go 的 `usermodel` / `ordermodel`,handler 里 feature dto 与共享 dto 同名时统一 `pkgdto`)。
- **DTO 命名用 Req/Resp 后缀**:`RegisterReq` / `UserResp`;统一包络不逐 feature 定义,handler 注解直接引用共享泛型,如 `pkgdto.Resp[dto.UserResp]`、`pkgdto.ListResp[dto.UserResp]`、失败一律 `pkgdto.ErrorResp`。

## 模块依赖规则(重要)

- 模块间依赖**单向无环**:`order → user` 允许,反向禁止。
- 跨模块只允许调用对方 **service 包**暴露的方法(如 `order/service` 调用 `user/service` 的 `Exists` 校验用户存在);**禁止 import 对方的 repository / model / handler / dto**。
- 新增模块时沿用同一规则;出现双向依赖需求时,把共享能力下沉到第三个模块或在 `pkg/` 建公共包。

## 快速开始

```bash
# 1. 启动 MySQL + Redis
docker compose -f deployments/docker-compose.yaml up -d

# 2. 运行(启动时自动 GORM AutoMigrate 建表)
make run

# 也可以先复制一份配置再改
cp configs/config.example.yaml configs/config.yaml
```

- API 地址:`http://localhost:8082/api/v1`
- Swagger UI:`http://localhost:8082/swagger/index.html`
- 健康检查:`GET /health`、`GET /health/live`、`GET /health/ready`

环境变量可覆盖任意配置项(点号转下划线、全大写),如 `SERVER_PORT=9000`、`JWT_SECRET=xxx` 会覆盖 `server.port`、`jwt.secret`。

## 常用命令

```bash
make build    # 编译到 bin/server
make test     # 测试(-race + 覆盖率)
make lint     # golangci-lint
make run      # 本地运行
make swagger  # 重新生成 docs/
```

## API 约定

统一响应包络(运行时由 golang-common/response 提供,swagger 文档侧由 `pkg/dto` 的 `Resp[T]` 泛型镜像):

```json
{ "code": 0, "message": "success", "data": {}, "requestId": "550e8400-..." }
```

错误码:通用码 1000–17999 由 golang-common 内置;业务码 20000 起,集中定义在 `pkg/xerr/error.go`,按 feature 分区(user 20000–20099,order 20100–20199,新模块依次 +100)。

### 认证流程

1. `POST /api/v1/users/register` 注册,`POST /api/v1/users/login` 登录,返回 `{accessToken, refreshToken}`。
2. access token 有效期 2 小时,无状态校验;请求受保护接口携带 `Authorization: Bearer <accessToken>`。
3. refresh token 有效期 7 天,其 **SHA-256 哈希**存 Redis(`user:refresh:<userID>`);`POST /api/v1/users/refresh` 轮换令牌对,旧 refresh token 立即失效。
4. `POST /api/v1/users/logout` 删除 Redis key,吊销 refresh token。
5. `GET /api/v1/users/:id` 演示 cache-aside:先查 Redis(`user:detail:<id>`),miss 走 DB 并回写(10 分钟 TTL);update/delete 时使缓存失效。

## 如何新增一个 feature 模块

以新增 `product` 模块为例:

1. 建 feature 侧:`internal/product/{router.go, handler/, service/, dto/}`;建 infra 侧:`internal/infra/db/product/{model/, repository/}`。子包一律短名。
2. `service/` 定义 `Repository`(及需要的 `Cache`)接口;`infra/db/product/repository` 提供 GORM 实现,约定"不存在返回 `gorm.ErrRecordNotFound`",由 service 的 `mapRepoErr` 翻译成 pkg/xerr 错误。
3. 在 `pkg/xerr/error.go` 文末为该 feature 追加一个分区(如 product 20200 起),`ecode.New` 对重复 code 会 panic;service 业务失败直接返回这些错误,handler 不需要 handleError 映射。
4. `dto/` 只定义本模块的请求/响应结构,命名用 `Req`/`Resp` 后缀(如 `CreateProductReq`、`ProductResp`);包络与分页类型**不要**在 feature 里重复定义——handler 注解直接引用共享泛型 `pkgdto.Resp[dto.ProductResp]` / `pkgdto.ListResp[dto.ProductResp]`,错误响应一律 `pkgdto.ErrorResp`,分页请求内嵌 `pkgdto.PageReq`。
5. `router.go` 写 `RegisterRoutes(rg *gin.RouterGroup, h *handler.Handler)`,需要认证的路由额外收 `jwtMiddleware gin.HandlerFunc`。
6. 在 `cmd/server/main.go` 手工装配:`repository.New → service.NewService → handler.NewHandler`,AutoMigrate 加入新模型,调用 `product.RegisterRoutes(v1, productHandler)`。
7. `make swagger && make build` 验证。

## 生产化清单(模板刻意从简的部分)

- **AutoMigrate → golang-migrate**:启动迁移只适合开发;生产用版本化迁移文件,CI 中执行。
- **内存限流 → Redis 限流**:`middleware.RateLimit` 的计数在进程内存,仅单实例有效;多实例部署换成基于 Redis 的分布式限流(main.go 中默认注释掉)。
- **JWT secret** 必须经环境变量注入,不要提交真实值;access/refresh 使用同一 secret,生产可拆成两把。
- **user CRUD 未挂认证**:模板为演示缓存与 CRUD 保持公开;实际项目按需挂 `auth.JWTMiddleware` 并加权限校验(通用码 12000 段)。
- **refresh token 单点登录**:一个用户只保留一条登录态(`user:refresh:<userID>`),多端登录需要把 key 改为 `<userID>:<sessionID>` 并维护会话集合。
- **config.yaml 不要提交真实环境配置**:仓库只提交 `config.example.yaml`,各环境用环境变量或私密配置下发。
