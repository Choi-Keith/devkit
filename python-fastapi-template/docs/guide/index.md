# 快速开始

欢迎使用FastAPI Backend Template！本指南将帮助您快速搭建并运行项目。

## 系统要求

- **Python**: 3.11+
- **操作系统**: Windows、macOS、Linux
- **内存**: 建议4GB以上
- **存储**: 至少1GB可用空间

## 安装步骤

### 1. 获取项目

```bash
git clone https://github.com/your-org/python-fastapi-template.git
cd FastAPI-Template
```

### 2. 安装UV包管理器

=== "Linux/macOS"

    ```bash
    curl -LsSf https://astral.sh/uv/install.sh | sh
    ```

=== "Windows"

    ```powershell
    powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
    ```

=== "使用pip"

    ```bash
    pip install uv
    ```

### 3. 安装依赖

```bash
# 安装项目依赖
uv sync

# 安装开发依赖
uv sync --dev
```

### 4. 环境配置

复制环境配置文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置必要的环境变量：

```env
# 应用配置
APP_ENV=development
SECRET_KEY=your-secret-key-here
DEBUG=True

# 数据库配置
DB_ENGINE=sqlite
DB_NAME=fastapi_template.db

# JWT配置
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=240
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7

# CORS配置
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

### 5. 初始化数据库

```bash
# 初始化数据库
uv run alembic upgrade head
```

### 6. 启动服务

```bash
# 开发环境启动
uv run uvicorn src:app --reload --host 0.0.0.0 --port 8000
```

## 验证安装

### 1. 检查健康状态

```bash
curl http://localhost:8000/api/v1/base/health
```

预期响应：

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "environment": "development",
  "service": "FastAPI Backend Template"
}
```

### 2. 访问API文档

打开浏览器访问以下地址：

- **Swagger UI**: http://localhost:8000/docs
- **ReDoc**: http://localhost:8000/redoc

### 3. 测试登录

使用默认管理员账号登录：

```bash
curl -X POST "http://localhost:8000/api/v1/base/access_token" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "abcd1234"
  }'
```

## 项目结构

```
FastAPI-Template/
├── src/                    # 源代码目录
│   ├── api/               # API路由层
│   │   └── v1/           # API版本v1
│   ├── services/         # 业务逻辑层
│   ├── repositories/     # 数据访问层
│   ├── models/           # 数据模型
│   ├── schemas/          # 数据验证模式
│   ├── core/             # 核心功能
│   ├── utils/            # 工具函数
│   └── main.py           # 应用入口
├── tests/                 # 测试文件
├── migrations/           # 数据库迁移文件
├── docs/                 # 文档源文件
├── .env.example          # 环境变量示例
├── pyproject.toml        # 项目配置
└── README.md             # 项目说明
```

## 常见问题

### Q: 如何更改默认端口？

A: 在启动命令中指定端口：

```bash
uv run uvicorn src:app --reload --host 0.0.0.0 --port 8080
```

### Q: 如何切换到PostgreSQL？

A: 修改 `.env` 文件中的数据库配置：

```env
DB_ENGINE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=fastapi_template
DB_USER=your_username
DB_PASSWORD=your_password
```

### Q: 如何重置数据库？

A: 删除数据库文件和迁移记录：

```bash
# SQLite
rm fastapi_template.db

# 重新初始化
uv run alembic upgrade head
```

### Q: 如何更改默认管理员密码？

A: 登录后通过用户管理接口修改，或者在首次启动时通过环境变量设置：

```env
DEFAULT_ADMIN_PASSWORD=your_new_password
```

## 下一步

- 📖 阅读 [架构设计](../architecture/) 了解系统架构
- 🔧 查看 [开发指南](../development/) 了解开发规范
- 📚 浏览 [API文档](../api/) 了解接口使用
- 🚀 学习 [部署指南](../development/deployment.md) 进行生产部署

## 获取帮助

如果您在使用过程中遇到问题，可以：

1. 访问 [官网](https://your-org.github.io/python-fastapi-template/) 获取最新文档
2. 查看 [常见问题](../faq.md)
3. 搜索 [GitHub Issues](https://github.com/your-org/python-fastapi-template/issues)
4. 创建新的 [Issue](https://github.com/your-org/python-fastapi-template/issues/new)
