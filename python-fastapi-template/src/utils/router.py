"""路由遍历工具。

FastAPI 新版（惰性 include_router）中 app.routes 可能包含 _IncludedRouter
而非 APIRoute，需通过 effective_candidates() 展开为带完整前缀与合并依赖的
有效路由（_EffectiveRouteContext，属性与 APIRoute 同名）。
"""

from collections.abc import Iterator

from starlette.routing import BaseRoute


def iter_api_routes(routes: list[BaseRoute]) -> Iterator:
    """展开路由列表，产出所有有效路由（APIRoute 或 _EffectiveRouteContext）。"""
    for route in routes:
        if hasattr(route, "effective_candidates"):
            # FastAPI 惰性挂载的 included router，递归展开
            yield from iter_api_routes(list(route.effective_candidates()))
        elif hasattr(route, "path_format"):
            yield route


def route_dependencies(route) -> list:
    """取路由的依赖列表（普通 starlette Route 没有该属性）。"""
    return getattr(route, "dependencies", []) or []
