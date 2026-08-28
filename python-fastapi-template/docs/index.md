# FastAPI Backend Template

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } **开箱即用**

    ---

    企业级FastAPI后端模板，内置三层架构设计，提供完整的RBAC权限管理系统

    [:octicons-arrow-right-24: 快速开始](guide/)

-   :material-account-group:{ .lg .middle } **用户管理**

    ---

    完整的用户管理系统，支持用户注册、登录、权限控制等功能

    [:octicons-arrow-right-24: 用户管理](api/users.md)

-   :material-shield-check:{ .lg .middle } **权限控制**

    ---

    基于角色的访问控制(RBAC)，支持细粒度的权限管理

    [:octicons-arrow-right-24: 权限系统](api/roles.md)

-   :material-database:{ .lg .middle } **数据库**

    ---

    基于SQLAlchemy 2.x异步ORM，支持PostgreSQL和SQLite，包含完整的迁移系统

    [:octicons-arrow-right-24: 数据库设计](architecture/database.md)

</div>

## 项目特性

### 🏗️ 架构设计

- **三层架构**: API → Service → Repository → Model
- **依赖注入**: 基于FastAPI的依赖系统
- **异步支持**: 全异步设计，支持高并发
- **类型安全**: 完整的Python类型标注

### 🔐 安全特性

- **JWT认证**: 访问令牌(4小时) + 刷新令牌(7天)
- **RBAC权限**: 基于角色的访问控制
- **密码安全**: Argon2哈希算法
- **限流保护**: 登录频率限制

### 📁 核心功能

- **用户管理**: 用户CRUD、密码重置、状态管理
- **角色管理**: 角色分配、权限绑定
- **菜单管理**: 动态菜单配置
- **文件管理**: 安全的文件上传下载
- **审计日志**: 完整的操作记录
- **部门管理**: 组织架构管理

### 🛠️ 开发工具

- **UV包管理**: 快速的Python包管理器
- **代码规范**: Black + Ruff + MyPy
- **测试覆盖**: Pytest + 异步测试
- **数据库迁移**: Alembic迁移工具
- **API文档**: 自动生成OpenAPI文档

## 技术栈

=== "后端框架"

    - **FastAPI**: 现代、高性能的Web框架
    - **SQLAlchemy 2.x**: 异步ORM框架
    - **Pydantic**: 数据验证和设置管理
    - **JWT**: JSON Web Token认证

=== "数据库"

    - **PostgreSQL**: 生产环境推荐
    - **SQLite**: 开发环境默认
    - **Redis**: 缓存和会话存储
    - **Alembic**: 数据库迁移工具

=== "开发工具"

    - **UV**: Python包管理器
    - **Black**: 代码格式化
    - **Ruff**: 代码检查
    - **MyPy**: 类型检查
    - **Pytest**: 测试框架

=== "部署运维"

    - **Docker**: 容器化部署
    - **GitHub Actions**: CI/CD自动化
    - **Uvicorn**: ASGI服务器
    - **Nginx**: 反向代理

## 快速开始

```bash
# 克隆项目
git clone https://github.com/your-org/python-fastapi-template.git
cd FastAPI-Template

# 安装依赖
curl -LsSf https://astral.sh/uv/install.sh | sh
uv sync

# 初始化数据库
uv run alembic upgrade head

# 启动开发服务器
uv run uvicorn src:app --reload --host 0.0.0.0 --port 8000
```

访问 [http://localhost:8000/docs](http://localhost:8000/docs) 查看交互式API文档。

## 文档导航

<div class="grid cards" markdown>

-   [:material-book-open-page-variant:{ .lg .middle } **快速开始**](guide/)

    项目安装、配置和运行指南

-   [:material-library-outline:{ .lg .middle } **架构设计**](architecture/)

    系统架构、设计模式和最佳实践

-   [:material-api:{ .lg .middle } **API文档**](api/)

    完整的API接口文档和使用示例

-   [:material-code-braces:{ .lg .middle } **开发指南**](development/)

    开发环境配置和编码规范

</div>

## 贡献指南

欢迎参与项目建设！请参考以下步骤：

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](https://github.com/your-org/python-fastapi-template/blob/main/LICENSE) 文件。

## 联系我们

- **🌐 官网**: [http://your-org.github.io/python-fastapi-template/](http://your-org.github.io/python-fastapi-template/)
- **GitHub**: [your-org/python-fastapi-template](https://github.com/your-org/python-fastapi-template)
- **Issues**: [问题反馈](https://github.com/your-org/python-fastapi-template/issues)

---

<p align="center">
  <i>由 Your Name 用 ❤️ 制作</i>
</p>
