from fastapi import APIRouter

from .users import router

users_router = APIRouter()
users_router.include_router(router, prefix="/users", tags=["用户模块"])

__all__ = ["users_router"]
