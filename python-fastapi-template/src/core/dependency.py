import re
import secrets
from functools import lru_cache
from typing import Optional

import jwt
from fastapi import Depends, HTTPException, Request, status
from fastapi.security import (
    HTTPAuthorizationCredentials,
    HTTPBasic,
    HTTPBasicCredentials,
    HTTPBearer,
)
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from core.ctx import CTX_USER_ID
from core.database import get_session
from models import Role, User
from settings.config import settings
from utils.cache import cache_manager


@lru_cache(maxsize=512)
def _compile_perm_pattern(perm_path: str) -> re.Pattern:
    """缓存权限路径的正则编译。/api/v1/x/{id} → ^/api/v1/x/[^/]+$"""
    pattern = re.sub(r"\{[^}]+\}", r"[^/]+", perm_path)
    return re.compile(f"^{pattern}$")


async def _get_user_permission_apis(user_id: int) -> list[tuple[str, str]]:
    """获取用户全部权限 API 列表 (method, path)。带 5 分钟缓存。

    miss 时一次性 prefetch 所有 role.apis，避免 N+1。
    """
    cache_key = f"user_perms:{user_id}"
    cached = await cache_manager.get(cache_key)
    if cached is not None:
        return [tuple(item) for item in cached]

    user = (
        (
            await get_session().execute(
                select(User)
                .options(selectinload(User.roles).selectinload(Role.apis))
                .where(User.id == user_id)
            )
        )
        .scalars()
        .first()
    )
    if user is None:
        return []
    perms: set[tuple[str, str]] = set()
    for role in user.roles:
        for api in role.apis:
            perms.add((api.method, api.path))
    perms_list = list(perms)
    await cache_manager.set(cache_key, [list(p) for p in perms_list], ttl=300)
    return perms_list


security = HTTPBasic()
bearer_scheme = HTTPBearer(auto_error=False)


def get_current_username(
    credentials: HTTPBasicCredentials = Depends(security),
):
    correct_username = secrets.compare_digest(
        credentials.username, settings.SWAGGER_UI_USERNAME
    )
    correct_password = secrets.compare_digest(
        credentials.password, settings.SWAGGER_UI_PASSWORD
    )
    if not (correct_username and correct_password):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authentication Required",
            headers={"WWW-Authenticate": "Basic"},
        )
    return credentials.username


class AuthControl:
    @classmethod
    async def is_authed(
        cls, token: HTTPAuthorizationCredentials | None = Depends(bearer_scheme)
    ) -> Optional["User"]:
        try:
            # 直接使用 HTTPBearer 提供的 token (已经去掉了 Bearer 前缀)
            if token is None or not token.credentials:
                raise HTTPException(
                    status_code=401, detail="Missing authentication token"
                )

            decode_data = jwt.decode(
                token.credentials,
                settings.SECRET_KEY,
                algorithms=settings.JWT_ALGORITHM,
            )
            user_id = decode_data.get("user_id")
            user = (
                (await get_session().execute(select(User).where(User.id == user_id)))
                .scalars()
                .first()
            )
            if not user:
                raise HTTPException(status_code=401, detail="Authentication failed")
            CTX_USER_ID.set(int(user_id))
            return user
        except jwt.DecodeError as e:
            raise HTTPException(status_code=401, detail="无效的Token") from e
        except jwt.ExpiredSignatureError as e:
            raise HTTPException(status_code=401, detail="登录已过期") from e
        except Exception as e:
            # 记录详细错误信息到日志，但不返回给用户
            raise HTTPException(status_code=401, detail="认证失败") from e


class PermissionControl:
    @classmethod
    async def has_permission(
        cls,
        request: Request,
        current_user: User = Depends(AuthControl.is_authed),
    ) -> None:
        """检查用户是否有访问指定API的权限。

        权限列表带 5 分钟缓存，正则编译带 LRU 缓存，消除 N+1 与重复编译。
        """
        if current_user.is_superuser:
            return

        permission_apis = await _get_user_permission_apis(current_user.id)
        if not permission_apis:
            raise HTTPException(
                status_code=403, detail="The user is not bound to a role"
            )

        method = request.method
        path = request.url.path
        for perm_method, perm_path in permission_apis:
            if method == perm_method and _compile_perm_pattern(perm_path).match(path):
                return

        raise HTTPException(
            status_code=403,
            detail=f"Permission denied method:{method} path:{path}",
        )

DependAuth = Depends(AuthControl.is_authed)
DependPermisson = Depends(PermissionControl.has_permission)
