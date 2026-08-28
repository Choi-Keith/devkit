"""Worker 进程的启动 / 关闭钩子。

worker 是独立进程，不会执行 FastAPI 的 lifespan。SQLAlchemy engine 的连接是
惰性建立的，无需显式初始化；任务里通过 core.database.session_scope() 获取
session 即可。这里仅负责连接 cache_manager。
"""

from __future__ import annotations

from log import logger
from utils.cache import cache_manager


async def worker_startup(ctx: dict) -> None:
    await cache_manager.connect()
    logger.info("arq worker 启动: Redis 就绪（数据库连接按需建立）")


async def worker_shutdown(ctx: dict) -> None:
    try:
        await cache_manager.disconnect()
    finally:
        from core.database import close_db

        await close_db()
        logger.info("arq worker 关闭")
