import locale
import os

from fastapi import APIRouter, HTTPException, Request
from slowapi import Limiter
from slowapi import extension as slowapi_extension
from slowapi.util import get_remote_address
from starlette.config import Config as StarletteConfig

from core.ctx import CTX_USER_ID
from core.dependency import DependAuth
from models.admin import User
from repositories.user import user_repository
from schemas.base import Fail, Success
from schemas.login import (
    CredentialsSchema,
    JWTOut,
    RefreshTokenRequest,
    TokenRefreshOut,
)
from schemas.response import CurrentUserResponse, TokenResponse
from settings import settings
from utils.jwt import create_token_pair, verify_token


class AdaptiveEnvConfig(StarletteConfig):
    def _read_file(self, file_name, encoding: str | None = None):
        encodings = ["utf-8", "utf-8-sig"]
        if encoding and encoding.lower() not in (e.lower() for e in encodings):
            encodings.insert(0, encoding)
        preferred = locale.getpreferredencoding(do_setlocale=False)
        if preferred and preferred.lower() not in (e.lower() for e in encodings):
            encodings.append(preferred)
        encodings.append("latin-1")  # final fallback to avoid UnicodeDecodeError

        last_error: UnicodeDecodeError | None = None
        for encoding in encodings:
            try:
                file_values: dict[str, str] = {}
                with open(file_name, encoding=encoding) as input_file:
                    for line in input_file.readlines():
                        line = line.strip()
                        if "=" in line and not line.startswith("#"):
                            key, value = line.split("=", 1)
                            key = key.strip()
                            value = value.strip().strip("\"'")
                            file_values[key] = value
                return file_values
            except UnicodeDecodeError as exc:
                last_error = exc
        if last_error:
            raise last_error
        return {}


if getattr(slowapi_extension.Config, "__name__", "") != "AdaptiveEnvConfig":
    slowapi_extension.Config = AdaptiveEnvConfig

limiter = Limiter(key_func=get_remote_address)

router = APIRouter()


def apply_rate_limit(rate="5/minute"):
    """根据环境应用限流装饰器"""

    def decorator(func):
        if os.getenv("TESTING", "false").lower() == "true":
            return func  # 测试环境不应用限流
        return limiter.limit(rate)(func)

    return decorator


@router.post("/login", summary="获取token", response_model=TokenResponse)
@apply_rate_limit()
async def login_access_token(request: Request, credentials: CredentialsSchema):
    user: User = await user_repository.authenticate(credentials)
    await user_repository.update_last_login(user.id)

    # 创建访问令牌和刷新令牌
    access_token, refresh_token = create_token_pair(user_id=user.id)

    data = JWTOut(
        access_token=access_token,
        refresh_token=refresh_token,
        username=user.username,
        expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
    )
    return Success(data=data.model_dump())


@router.post("/refresh", summary="刷新token", response_model=TokenResponse)
@apply_rate_limit("10/minute")
async def refresh_access_token(request: Request, refresh_request: RefreshTokenRequest):
    """
    使用刷新令牌获取新的访问令牌和刷新令牌
    """
    try:
        # 验证刷新令牌
        payload = verify_token(refresh_request.refresh_token, token_type="refresh")

        # 验证用户是否仍然存在且有效
        user = await user_repository.get(id=payload.user_id)
        if not user or not user.is_active:
            return Fail(code=401, msg="用户不存在或已被禁用")

        # 创建新的令牌对
        access_token, refresh_token = create_token_pair(user_id=user.id)

        data = TokenRefreshOut(
            access_token=access_token,
            refresh_token=refresh_token,
            expires_in=settings.JWT_ACCESS_TOKEN_EXPIRE_MINUTES * 60,
        )

        return Success(data=data.model_dump())

    except Exception as exc:  # noqa: BLE001 - propagate as HTTP error for clarity
        raise HTTPException(status_code=401, detail="令牌无效或已过期") from exc


@router.get("/me", summary="查看用户信息", response_model=CurrentUserResponse)
async def get_userinfo(current_user: User = DependAuth):
    user_id = CTX_USER_ID.get()
    user_obj = await user_repository.get(id=user_id)
    user_dict = user_obj.to_dict()
    return Success(data=user_dict)
