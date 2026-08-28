# 常见问题

## 🚀 快速开始

### Q: 安装依赖时遇到错误怎么办？

A: 请确保您使用的是Python 3.11+版本，并且网络连接正常。如果遇到包冲突，可以尝试：

```bash
# 清理缓存
uv cache clean

# 重新安装
uv sync --reinstall
```

### Q: 如何修改默认端口？

A: 在启动命令中指定端口：

```bash
uv run uvicorn src:app --reload --host 0.0.0.0 --port 8080
```

或者在 `.env` 文件中设置：

```env
PORT=8080
```

### Q: 默认管理员账号是什么？

A: 默认管理员账号：
- 用户名: `admin`
- 密码: `abcd1234`

**重要**: 请在生产环境中立即修改默认密码！

## 🔧 配置相关

### Q: 如何切换数据库？

A: 修改 `.env` 文件中的数据库配置：

=== "PostgreSQL"

    ```env
    DB_ENGINE=postgres
    DB_HOST=localhost
    DB_PORT=5432
    DB_NAME=fastapi_template
    DB_USER=your_username
    DB_PASSWORD=your_password
    ```

=== "SQLite"

    ```env
    DB_ENGINE=sqlite
    DB_NAME=fastapi_template.db
    ```

### Q: 如何配置CORS？

A: 在 `.env` 文件中设置允许的源：

```env
CORS_ORIGINS=http://localhost:3000,http://localhost:8080,https://yourdomain.com
```

### Q: 如何更改JWT过期时间？

A: 在 `.env` 文件中配置：

```env
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=240  # 访问令牌4小时
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7      # 刷新令牌7天
```

## 🗄️ 数据库相关

### Q: 如何重置数据库？

A: 删除数据库文件并重新初始化：

```bash
# SQLite
rm fastapi_template.db

# 删除迁移记录
rm -rf migrations/

# 重新初始化
uv run alembic upgrade head
```

### Q: 如何添加新的数据表？

A: 按照以下步骤：

1. 在 `src/models/` 中定义模型
2. 生成迁移文件
3. 应用迁移

```bash
# 生成迁移
uv run alembic revision --autogenerate -m "add_new_table"

# 应用迁移
uv run alembic upgrade head
```

### Q: 迁移失败怎么办？

A: 检查迁移历史并手动修复：

```bash
# 查看迁移历史
uv run alembic history

# 如果需要回滚
uv run alembic downgrade -1

# 手动修复后重新迁移
uv run alembic revision --autogenerate -m "fix_migration"
uv run alembic upgrade head
```

## 🔐 认证授权

### Q: 如何添加新的用户角色？

A: 通过API或直接在数据库中添加：

```python
# 通过代码添加
from src.models.admin import Role

role = await Role.create(
    name="editor",
    description="编辑者角色"
)
```

### Q: 如何自定义权限检查？

A: 创建自定义依赖项：

```python
from fastapi import Depends, HTTPException
from src.core.dependency import get_current_user

def require_admin(current_user = Depends(get_current_user)):
    if not current_user.is_superuser:
        raise HTTPException(status_code=403, detail="需要管理员权限")
    return current_user
```

### Q: JWT令牌过期后如何处理？

A: 使用刷新令牌获取新的访问令牌：

```python
# 使用刷新令牌
response = requests.post("/api/v1/base/refresh_token", json={
    "refresh_token": "your_refresh_token"
})
```

## 📁 文件管理

### Q: 如何限制文件上传大小？

A: 在 `src/services/file_service.py` 中修改：

```python
MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB
```

### Q: 如何添加新的文件类型支持？

A: 修改允许的文件类型列表：

```python
ALLOWED_EXTENSIONS = {
    'image': ['.jpg', '.jpeg', '.png', '.gif', '.bmp'],
    'document': ['.pdf', '.doc', '.docx', '.txt'],
    'video': ['.mp4', '.avi', '.mkv']  # 新增视频支持
}
```

### Q: 上传的文件存储在哪里？

A: 默认存储在 `uploads/` 目录下，可以通过环境变量修改：

```env
UPLOAD_DIR=/path/to/uploads
```

## 🧪 测试相关

### Q: 如何运行测试？

A: 使用pytest运行测试：

```bash
# 运行所有测试
uv run pytest

# 运行特定测试文件
uv run pytest tests/test_users.py

# 运行带覆盖率的测试
uv run pytest --cov=src --cov-report=html
```

### Q: 如何添加新的测试？

A: 在 `tests/` 目录下创建测试文件：

```python
import pytest
from httpx import AsyncClient
from src.main import app

@pytest.mark.asyncio
async def test_create_user():
    async with AsyncClient(app=app, base_url="http://test") as ac:
        response = await ac.post("/api/v1/users/create", json={
            "username": "testuser",
            "password": "password123"
        })
    assert response.status_code == 200
```

## 🚀 部署相关

### Q: 如何部署到生产环境？

A: 使用Docker部署：

```bash
# 构建镜像
docker build -t fastapi-template .

# 运行容器
docker run -d -p 8000:8000 --name fastapi-app fastapi-template
```

### Q: 如何配置反向代理？

A: Nginx配置示例：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Q: 如何设置环境变量？

A: 生产环境推荐使用环境变量而不是 `.env` 文件：

```bash
export SECRET_KEY="your-secret-key"
export DB_HOST="your-db-host"
export DB_PASSWORD="your-db-password"
```

## 🐛 故障排除

### Q: 应用启动时报错怎么办？

A: 检查以下项目：

1. 确保所有依赖已安装
2. 检查数据库连接配置
3. 验证环境变量设置
4. 查看详细错误日志

```bash
# 查看详细日志
uv run uvicorn src:app --reload --log-level debug
```

### Q: 数据库连接失败？

A: 检查数据库配置和连接：

```python
# 测试数据库连接
from src.core.database import get_db_connection

async def test_db():
    try:
        conn = await get_db_connection()
        print("数据库连接成功")
    except Exception as e:
        print(f"数据库连接失败: {e}")
```

### Q: 如何启用调试模式？

A: 在 `.env` 文件中设置：

```env
DEBUG=True
APP_ENV=development
```

## 📚 开发相关

### Q: 如何添加新的API端点？

A: 按照三层架构添加：

1. 定义模型 (`src/models/`)
2. 创建仓储 (`src/repositories/`)
3. 实现服务 (`src/services/`)
4. 添加路由 (`src/api/v1/`)

### Q: 如何自定义中间件？

A: 在 `src/core/middleware.py` 中添加：

```python
from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware

class CustomMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        # 处理请求前
        response = await call_next(request)
        # 处理响应后
        return response
```

### Q: 如何添加定时任务？

A: 使用APScheduler：

```python
from apscheduler.schedulers.asyncio import AsyncIOScheduler

scheduler = AsyncIOScheduler()

@scheduler.scheduled_job("interval", minutes=30)
async def cleanup_expired_tokens():
    # 清理过期令牌
    pass

scheduler.start()
```

## 🔍 性能优化

### Q: 如何优化数据库查询？

A: 使用以下技巧：

1. 使用 `select_related()` 预加载关联数据
2. 使用 `prefetch_related()` 优化多对多查询
3. 添加适当的数据库索引
4. 使用查询分页

```python
# 预加载关联数据
users = await User.all().select_related("roles")

# 批量预加载
users = await User.all().prefetch_related("roles__permissions")
```

### Q: 如何添加缓存？

A: 使用Redis缓存：

```python
import redis
from functools import wraps

redis_client = redis.Redis(host='localhost', port=6379, db=0)

def cache_result(expire_time=300):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            cache_key = f"{func.__name__}:{hash(str(args)+str(kwargs))}"
            cached = redis_client.get(cache_key)
            if cached:
                return json.loads(cached)

            result = await func(*args, **kwargs)
            redis_client.setex(cache_key, expire_time, json.dumps(result))
            return result
        return wrapper
    return decorator
```

## 📞 获取帮助

如果以上FAQ没有解决您的问题，您可以：

1. 访问 [官网](https://your-org.github.io/python-fastapi-template/) 获取最新文档
2. 查看 [GitHub Issues](https://github.com/your-org/python-fastapi-template/issues)
3. 创建新的 [Issue](https://github.com/your-org/python-fastapi-template/issues/new)
4. 查看项目文档的其他部分

## 🤝 贡献指南

如果您发现了新的问题或有改进建议，欢迎：

1. 提交Issue报告问题
2. 提交PR改进文档
3. 参与讨论和代码审查

感谢您的贡献！
