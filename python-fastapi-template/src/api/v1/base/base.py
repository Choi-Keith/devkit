import ipaddress
import os
import platform
from datetime import UTC, datetime

from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import Response

from core.metrics import db_up, redis_up, render_metrics
from settings import settings
from utils.cache import cache_manager

router = APIRouter()


@router.get("/health", summary="健康检查")
async def health_check():
    """系统健康检查（同时刷新 redis_up / db_up 指标）"""

    redis_ok = cache_manager.is_available
    redis_up.set(1 if redis_ok else 0)

    db_ok = True
    try:
        from sqlalchemy import text

        from core.database import engine

        async with engine.connect() as conn:
            await conn.execute(text("SELECT 1"))
    except Exception:
        db_ok = False
    db_up.set(1 if db_ok else 0)

    return {
        "status": "healthy" if (redis_ok and db_ok) else "degraded",
        "timestamp": datetime.now(UTC).isoformat(),
        "version": settings.VERSION,
        "environment": settings.APP_ENV,
        "service": settings.PROJECT_NAME,
        "database": "connected" if db_ok else "disconnected",
        "redis": "connected" if redis_ok else "disconnected",
    }


def _is_metrics_allowed(client_ip: str | None) -> bool:
    if not client_ip:
        return False
    try:
        addr = ipaddress.ip_address(client_ip)
    except ValueError:
        return False
    return any(addr in net for net in settings.METRICS_ALLOWED_NETWORKS)


@router.get("/metrics", summary="Prometheus 指标", include_in_schema=False)
async def metrics_endpoint(request: Request):
    """暴露 Prometheus 指标。仅放行白名单 IP（默认 localhost + 内网网段）。"""
    client_ip = request.client.host if request.client else None
    if not _is_metrics_allowed(client_ip):
        raise HTTPException(status_code=403, detail="Forbidden")
    body, content_type = render_metrics()
    return Response(content=body, media_type=content_type)


@router.get("/version", summary="版本信息")
async def get_version():
    """获取API版本信息"""

    return {
        "version": settings.VERSION,
        "app_title": settings.APP_TITLE,
        "project_name": settings.PROJECT_NAME,
        "build": os.getenv("APP_BUILD", "dev"),
        "commit": os.getenv("GIT_COMMIT", "unknown"),
        "python_version": platform.python_version(),
    }
