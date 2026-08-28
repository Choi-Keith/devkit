"""数据库连接与 Session 管理（SQLAlchemy 2.x async）。

- 全局 engine + async_session_factory，启动即创建，连接惰性建立
- 请求级 session 由 DBSessionMiddleware 通过 contextvar 绑定，请求结束自动关闭
- 非请求场景（启动初始化、arq worker、脚本）使用 session_scope() 获取独立 session
"""

from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from contextvars import ContextVar

from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from starlette.types import ASGIApp, Receive, Scope, Send

from settings.config import settings


def _create_engine() -> AsyncEngine:
    return create_async_engine(
        settings.DATABASE_URL,
        echo=False,
        **settings.DATABASE_ENGINE_OPTIONS,
    )


engine: AsyncEngine = _create_engine()
async_session_factory: async_sessionmaker[AsyncSession] = async_sessionmaker(
    engine, expire_on_commit=False
)

_current_session: ContextVar[AsyncSession | None] = ContextVar(
    "db_session", default=None
)


def get_session() -> AsyncSession:
    """获取当前上下文的 session；无绑定时创建独立 session 并绑定到当前上下文。"""
    session = _current_session.get()
    if session is None:
        session = async_session_factory()
        _current_session.set(session)
    return session


@asynccontextmanager
async def session_scope() -> AsyncGenerator[AsyncSession, None]:
    """独立 session 上下文：自动提交/回滚/关闭，用于非请求场景。"""
    session = async_session_factory()
    token = _current_session.set(session)
    try:
        yield session
        await session.commit()
    except Exception:
        await session.rollback()
        raise
    finally:
        await session.close()
        _current_session.reset(token)


class DBSessionMiddleware:
    """每个 HTTP/WebSocket 请求绑定一个 session，请求结束后关闭。"""

    def __init__(self, app: ASGIApp):
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] not in ("http", "websocket"):
            await self.app(scope, receive, send)
            return
        session = async_session_factory()
        token = _current_session.set(session)
        try:
            await self.app(scope, receive, send)
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()
            _current_session.reset(token)


async def close_db() -> None:
    """释放数据库连接池（应用关闭时调用）。"""
    await engine.dispose()


def override_engine(new_engine: AsyncEngine) -> None:
    """替换全局 engine 与 session 工厂（测试注入用）。"""
    global engine, async_session_factory
    engine = new_engine
    async_session_factory = async_sessionmaker(new_engine, expire_on_commit=False)
