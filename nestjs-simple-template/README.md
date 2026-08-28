# nestjs-simple-template

生产可用的 NestJS 11 基础模板,与 `golang-simple-template` 约定一致:统一响应体 `{code, message, data}`、错误码分段管理、API 前缀 `/api/v1`、`/health` 健康检查。

技术栈:NestJS 11 + TypeScript 5 + TypeORM 0.3(MySQL / PostgreSQL 切换)+ cache-manager v6(Keyv Redis,可降级内存)+ nestjs-pino 日志 + @nestjs/throttler 限流 + JWT(access + refresh,argon2id 密码哈希)+ RBAC + class-validator + Joi 配置校验 + helmet + Swagger + terminus。

## 目录结构

```
├── src/
│   ├── main.ts                # helmet、全局前缀、URI 版本化、ValidationPipe、Swagger、pino
│   ├── app.module.ts          # 全局 Guard / Interceptor / Filter 装配
│   ├── config/                # typed config + Joi env 校验
│   ├── common/
│   │   ├── ecode/             # 错误码注册(common 1000-19999,业务从 20000 起按模块分文件)
│   │   ├── exceptions/        # BusinessException(业务 code + httpStatus)
│   │   ├── filters/           # GlobalExceptionFilter(统一 {code, message, data:null})
│   │   ├── interceptors/      # ResponseInterceptor(成功包装 {code:0, message:'success', data})
│   │   ├── decorators/        # @Public() @Roles() @CurrentUser()
│   │   ├── guards/            # RolesGuard(RBAC)
│   │   ├── enums/             # Role 枚举(ADMIN / USER)
│   │   └── dto/               # 分页 DTO、统一响应 DTO
│   ├── infra/
│   │   ├── logger/            # nestjs-pino(开发 pretty / 生产 JSON)
│   │   ├── database/          # TypeORM,DB_TYPE=mysql|postgres
│   │   └── cache/             # 全局缓存模块(Redis 不可用时降级内存,启动不崩)
│   └── modules/
│       ├── health/            # terminus: /health /health/live /health/ready(免前缀、免认证)
│       ├── auth/              # register / login / refresh / logout / profile
│       └── user/              # CRUD + 分页示例(entity/dto/service/controller/spec)
├── test/                      # e2e 冒烟(需真实 DB,RUN_DB_E2E=1 才执行)
├── .env.example               # 环境变量模板(.env 不提交)
├── Dockerfile                 # 多阶段构建,非 root 运行
└── docker-compose.yml         # mysql / postgres / redis
```

## 快速开始

```bash
# 1. 准备配置(注意 JWT_SECRET 必填)
cp .env.example .env

# 2. 启动依赖(MySQL / PostgreSQL / Redis)
docker compose up -d mysql redis      # 用 postgres 时:docker compose up -d postgres redis

# 3. 安装依赖并运行
npm install
npm run start:dev
```

启动后:

- Swagger 文档:`http://localhost:3000/docs`(仅非生产环境)
- 健康检查:`GET /health`、`/health/live`、`/health/ready`(不带 `/api` 前缀)
- 业务 API:`/api/v1/**`

切换到 PostgreSQL:`.env` 中设置 `DB_TYPE=postgres`、`DB_PORT=5432` 及对应账号即可。

## API 一览

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | /api/v1/auth/register | 注册(返回 token 对) | 公开 |
| POST | /api/v1/auth/login | 登录(限流 5 次/分钟) | 公开 |
| POST | /api/v1/auth/refresh | 刷新令牌(refresh token 轮换) | 公开 |
| POST | /api/v1/auth/logout | 登出(吊销 refresh token) | 登录 |
| GET | /api/v1/auth/profile | 当前用户信息 | 登录 |
| GET | /api/v1/users | 用户分页列表(`?page=1&pageSize=10`) | 登录 |
| GET | /api/v1/users/:id | 用户详情 | 登录 |
| POST | /api/v1/users | 创建用户 | ADMIN |
| PATCH | /api/v1/users/:id | 更新用户 | ADMIN |
| DELETE | /api/v1/users/:id | 删除用户 | ADMIN |

```bash
# 注册
curl -X POST localhost:3000/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"alice","password":"Passw0rd!"}'

# 登录
curl -X POST localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"Passw0rd!"}'

# 携带 access token 访问
curl localhost:3000/api/v1/auth/profile -H "Authorization: Bearer <accessToken>"
curl 'localhost:3000/api/v1/users?page=1&pageSize=10' -H "Authorization: Bearer <accessToken>"

# 刷新令牌
curl -X POST localhost:3000/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refreshToken":"<refreshToken>"}'
```

成功响应统一为:

```json
{ "code": 0, "message": "success", "data": { "list": [], "total": 0, "page": 1, "pageSize": 10 } }
```

## 错误码说明

响应体 `{code, message, data}`,HTTP 状态码与错误码语义一致。

| 段 | 范围 | 说明 |
| --- | --- | --- |
| 成功 | 0 | success |
| 通用错误 | 1000-19999 | `src/common/ecode/common.ecode.ts`:1000 参数校验失败、1001 未认证、1002 无权限、1003 资源不存在、1004 请求过于频繁、1500 服务器内部错误 |
| user 模块 | 20000-20999 | `src/common/ecode/user.ecode.ts`:20000 user not found、20001 email already exists、20002 invalid credentials |
| 新业务模块 | 21000 起 | 每模块一个 `*.ecode.ts`,按模块分段递增 |

业务代码抛错方式:

```ts
throw new BusinessException(UserEcode.USER_NOT_FOUND);           // 使用默认消息
throw new BusinessException(CommonEcode.UNAUTHORIZED, 'invalid refresh token'); // 覆盖消息
```

## 常用命令

```bash
npm run start:dev     # 开发模式(watch)
npm run build         # 编译到 dist/
npm run start:prod    # 生产运行
npm run lint          # ESLint
npm run format        # Prettier
npm test              # 单元测试(user / auth service)
npm run test:e2e      # e2e(需先起 DB 且 RUN_DB_E2E=1)
```

## 作为新项目模板使用

1. 复制本目录,修改 `package.json` 的 `name`,批量替换代码中的 `nestjs-simple-template`。
2. `cp .env.example .env` 并填写连接信息;生产环境务必替换 `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET`。
3. 新业务模块:在 `src/modules/` 下新建模块(entity/dto/service/controller),错误码在 `src/common/ecode/` 新增 `xxx.ecode.ts`(从 21000 起),并在 `app.module.ts` 装配。
4. 免认证接口加 `@Public()`;角色限制加 `@Roles(Role.ADMIN)`;跳过限流加 `@SkipThrottle()`,定制限流加 `@Throttle()`(示例见 `auth.controller.ts`)。
5. 首个管理员:注册普通用户后在数据库中将其 `roles` 改为 `["ADMIN"]`(如 `UPDATE users SET roles='["ADMIN"]' WHERE email='...'`)。
6. 生产环境关闭 `DB_SYNCHRONIZE`,使用 TypeORM migrations 管理表结构;`NODE_ENV=production` 时 Swagger 自动关闭、日志输出 JSON。

## 实现备注(版本取舍)

- `cache-manager-redis-yet` 已被官方废弃(cache-manager v6 迁移到 Keyv),且与 `@nestjs/cache-manager` v3 的 Keyv 包装不兼容(启动报 `Invalid storage adapter`)。本模板改用官方 `keyv` + `@keyv/redis`,Redis 不可用时探测失败即降级为内存缓存,启动不崩。
- 主要依赖版本:NestJS 11.2 / TypeORM 0.3.31 / cache-manager 6.4 / @nestjs/cache-manager 3.1 / @nestjs/throttler 6.5 / nestjs-pino 4.6 / @nestjs/terminus 11.1 / @nestjs/jwt 11.0 / argon2 0.45 / ESLint 9(flat config)/ Jest 30 + ts-jest 29。
- refresh token 携带 `jti` 保证每次签发唯一,避免同一秒内轮换后旧 token 仍与库存哈希匹配。
