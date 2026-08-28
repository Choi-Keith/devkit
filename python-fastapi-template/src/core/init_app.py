import asyncio
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware import Middleware
from fastapi.middleware.cors import CORSMiddleware
from slowapi import _rate_limit_exceeded_handler
from slowapi.errors import RateLimitExceeded
from sqlalchemy import func, or_, select

from api import api_router
from api.v1.auth.auth import limiter
from core.database import DBSessionMiddleware, session_scope
from core.exceptions import (
    DoesNotExistHandle,
    HTTPException,
    HttpExcHandle,
    IntegrityError,
    IntegrityHandle,
    NotFoundError,
    RequestValidationError,
    RequestValidationHandle,
    ResponseValidationError,
    ResponseValidationHandle,
    UnhandledExceptionHandle,
)
from core.metrics import MetricsMiddleware
from core.middlewares import (
    BackGroundTaskMiddleware,
    HttpAuditLogMiddleware,
    RequestLoggingMiddleware,
    SecurityHeadersMiddleware,
)
from log import logger
from models.admin import Api, Menu, Role
from repositories.api import api_repository
from repositories.user import UserCreate, user_repository
from schemas.menus import MenuType
from settings.config import settings


def make_middlewares():
    middleware = [
        Middleware(DBSessionMiddleware),  # 数据库 session 绑定（须最外层）
        Middleware(
            CORSMiddleware,
            allow_origins=settings.CORS_ORIGINS_LIST,
            allow_credentials=settings.CORS_ALLOW_CREDENTIALS,
            allow_methods=settings.CORS_ALLOW_METHODS,
            allow_headers=settings.CORS_ALLOW_HEADERS,
        ),
        Middleware(SecurityHeadersMiddleware),  # 安全头中间件
        Middleware(RequestLoggingMiddleware),  # 请求日志中间件 + trace_id
        Middleware(MetricsMiddleware),  # Prometheus 指标
        Middleware(BackGroundTaskMiddleware),
        Middleware(
            HttpAuditLogMiddleware,
            methods=["GET", "POST", "PUT", "DELETE"],
            exclude_paths=[
                "/api/v1/auth/login",
                "/docs",
                "/openapi.json",
            ],
        ),
    ]
    return middleware


def register_exceptions(app: FastAPI):
    app.add_exception_handler(NotFoundError, DoesNotExistHandle)
    app.add_exception_handler(HTTPException, HttpExcHandle)
    app.add_exception_handler(IntegrityError, IntegrityHandle)
    app.add_exception_handler(RequestValidationError, RequestValidationHandle)
    app.add_exception_handler(ResponseValidationError, ResponseValidationHandle)
    # 注册通用异常处理器（必须放在最后，作为兜底）
    app.add_exception_handler(Exception, UnhandledExceptionHandle)
    # 注册限流异常处理
    app.state.limiter = limiter
    app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)


def register_routers(app: FastAPI, prefix: str = "/api"):
    app.include_router(api_router, prefix=prefix)


async def init_superuser():
    logger.info("🔧 开始初始化超级管理员用户...")
    async with session_scope() as session:
        user_count = await session.scalar(select(func.count()).select_from(
            user_repository.model
        ))
        if not user_count:
            await user_repository.create_user(
                UserCreate(
                    username="admin",
                    email="admin@admin.com",
                    password="abcd1234",
                    is_active=True,
                    is_superuser=True,
                )
            )
            logger.info("✅ 超级管理员用户创建成功 - 用户名: admin")
        else:
            logger.info("ℹ️ 超级管理员用户已存在，跳过创建")


async def init_menus():
    logger.info("🔧 开始初始化系统菜单...")
    async with session_scope() as session:
        menu_count = await session.scalar(select(func.count()).select_from(Menu))
        if not menu_count:
            parent_menu = Menu(
                menu_type=MenuType.CATALOG,
                name="系统管理",
                path="/system",
                order=1,
                parent_id=0,
                icon="carbon:gui-management",
                is_hidden=False,
                component="Layout",
                keepalive=False,
                redirect="/system/user",
            )
            session.add(parent_menu)
            await session.flush()  # 获取父菜单 ID
            children_menu = [
                Menu(
                    menu_type=MenuType.MENU,
                    name="用户管理",
                    path="user",
                    order=1,
                    parent_id=parent_menu.id,
                    icon="material-symbols:person-outline-rounded",
                    is_hidden=False,
                    component="/system/user",
                    keepalive=False,
                ),
                Menu(
                    menu_type=MenuType.MENU,
                    name="角色管理",
                    path="role",
                    order=2,
                    parent_id=parent_menu.id,
                    icon="carbon:user-role",
                    is_hidden=False,
                    component="/system/role",
                    keepalive=False,
                ),
                Menu(
                    menu_type=MenuType.MENU,
                    name="菜单管理",
                    path="menu",
                    order=3,
                    parent_id=parent_menu.id,
                    icon="material-symbols:list-alt-outline",
                    is_hidden=False,
                    component="/system/menu",
                    keepalive=False,
                ),
                Menu(
                    menu_type=MenuType.MENU,
                    name="API管理",
                    path="api",
                    order=4,
                    parent_id=parent_menu.id,
                    icon="ant-design:api-outlined",
                    is_hidden=False,
                    component="/system/api",
                    keepalive=False,
                ),
                Menu(
                    menu_type=MenuType.MENU,
                    name="部门管理",
                    path="dept",
                    order=5,
                    parent_id=parent_menu.id,
                    icon="mingcute:department-line",
                    is_hidden=False,
                    component="/system/dept",
                    keepalive=False,
                ),
                Menu(
                    menu_type=MenuType.MENU,
                    name="审计日志",
                    path="auditlog",
                    order=6,
                    parent_id=parent_menu.id,
                    icon="ph:clipboard-text-bold",
                    is_hidden=False,
                    component="/system/auditlog",
                    keepalive=False,
                ),
            ]
            session.add_all(children_menu)
            session.add(
                Menu(
                    menu_type=MenuType.MENU,
                    name="一级菜单",
                    path="/top-menu",
                    order=2,
                    parent_id=0,
                    icon="material-symbols:featured-play-list-outline",
                    is_hidden=False,
                    component="/top-menu",
                    keepalive=False,
                    redirect="",
                )
            )
            await session.commit()
            logger.info("✅ 系统菜单初始化成功 - 菜单数量: 8")
        else:
            logger.info("ℹ️ 系统菜单已存在，跳过初始化")


async def init_apis():
    """每次启动同步 API 表与路由；若有增删则刷新管理员角色的 API 关联。

    解决"已有数据库升级后新加的路由不会被写入 Api 表"的问题：
    旧实现仅在 Api 表为空时才 refresh，导致升级路径下新启用的路由
    （如 menus/depts/auditlog）永远不会进入权限分配池。
    """
    logger.info("🔧 开始同步 API 数据...")
    async with session_scope() as session:
        result = await api_repository.refresh_api()
        api_count = await session.scalar(select(func.count()).select_from(Api))
        logger.info(
            f"✅ API 同步完成 - 当前 {api_count} 条 "
            f"(创建 {result['created']}, 更新 {result['updated']}, 删除 {result['deleted']})"
        )

        # 升级路径补偿：API 集合发生变化时，刷新管理员角色的全量 API 关联。
        # init_roles 仅在角色不存在时跑（首装），已有库下不会触发 → 必须在这里补。
        if result["created"] or result["deleted"]:
            admin = (
                (await session.execute(select(Role).where(Role.name == "管理员")))
                .scalars()
                .first()
            )
            if admin is not None:
                all_apis = list((await session.execute(select(Api))).scalars().all())
                admin.apis = all_apis
                session.add(admin)
                await session.commit()
                logger.info(f"🔄 管理员角色 API 关联已刷新（{len(all_apis)} 条）")


async def init_db():
    """通过 Alembic 将数据库迁移到最新版本（在线程中执行同步 API）。"""
    await asyncio.to_thread(_run_migrations)


def _run_migrations():
    from alembic import command
    from alembic.config import Config

    cfg = Config(str(Path(settings.BASE_DIR) / "alembic.ini"))
    command.upgrade(cfg, "head")


async def init_roles():
    logger.info("🔧 开始初始化用户角色...")
    async with session_scope() as session:
        role_count = await session.scalar(select(func.count()).select_from(Role))
        if not role_count:
            admin_role = Role(
                name="管理员",
                desc="管理员角色",
            )
            user_role = Role(
                name="普通用户",
                desc="普通用户角色",
            )

            # 分配所有API给管理员角色
            all_apis = list((await session.execute(select(Api))).scalars().all())
            admin_role.apis = all_apis
            # 分配所有菜单给管理员和普通用户
            all_menus = list((await session.execute(select(Menu))).scalars().all())
            admin_role.menus = all_menus
            user_role.menus = all_menus

            # 为普通用户分配基本API
            basic_apis = list(
                (
                    await session.execute(
                        select(Api).where(
                            or_(Api.method.in_(["GET"]), Api.tags == "基础模块")
                        )
                    )
                )
                .scalars()
                .all()
            )
            user_role.apis = basic_apis

            session.add_all([admin_role, user_role])
            await session.commit()

            logger.info("✅ 用户角色初始化成功 - 角色: 管理员, 普通用户")
        else:
            logger.info(f"ℹ️ 用户角色已存在，跳过初始化 - 当前角色数量: {role_count}")


async def init_data():
    logger.info("🚀 系统初始化开始...")

    logger.info("🔧 开始数据库初始化和迁移...")
    await init_db()
    logger.info("✅ 数据库初始化完成")

    logger.info("🔄 并行初始化基础数据...")
    await asyncio.gather(
        init_superuser(),
        init_menus(),
        init_apis(),
    )
    logger.info("✅ 基础数据初始化完成")

    await init_roles()

    logger.info("🎉 系统初始化完成！")
