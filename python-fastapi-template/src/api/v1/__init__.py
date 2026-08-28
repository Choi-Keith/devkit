from fastapi import APIRouter

from core.dependency import DependAuth, DependPermisson

from .apis import apis_router
from .auditlog import auditlog_router
from .auth import auth_router
from .base import base_router
from .depts import depts_router
from .files import files_router
from .menus import menus_router
from .roles import roles_router
from .tasks import tasks_router
from .users import users_router

v1_router = APIRouter()

v1_router.include_router(base_router, prefix="/base")
v1_router.include_router(auth_router, prefix="/auth")
# 资源路由的复数前缀在各子包 __init__ 中声明（FastAPI 不允许 include 时 prefix 与 path 同时为空）
v1_router.include_router(users_router, dependencies=[DependPermisson])
v1_router.include_router(roles_router, dependencies=[DependPermisson])
v1_router.include_router(menus_router, dependencies=[DependPermisson])
v1_router.include_router(apis_router, dependencies=[DependPermisson])
v1_router.include_router(depts_router, dependencies=[DependPermisson])
v1_router.include_router(auditlog_router, dependencies=[DependPermisson])
v1_router.include_router(files_router, dependencies=[DependPermisson])
# 任务查询路由：使用基础鉴权 + 路由内 superuser 校验，绕过 RBAC API 列表（避免每次新增任务都要刷 API）
v1_router.include_router(tasks_router, prefix="/tasks", dependencies=[DependAuth])

__all__ = ["v1_router"]
