from fastapi import APIRouter

from .files import router

files_router = APIRouter()
files_router.include_router(router, prefix="/files", tags=["上传文件"])

__all__ = ["files_router"]
