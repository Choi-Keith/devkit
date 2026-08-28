import json
import logging

from fastapi import APIRouter, Query
from sqlalchemy import func, select

from core.database import get_session
from models.admin import Menu
from repositories.menu import menu_repository
from schemas.base import Fail, Success, SuccessExtra
from schemas.menus import *
from schemas.response import (
    MenuDetailResponse,
    MenuListResponse,
    ResponseBase,
)

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("", summary="查看菜单列表", response_model=MenuListResponse)
async def list_menu(
    page: int = Query(1, description="页码"),
    page_size: int = Query(10, description="每页数量"),
):
    async def get_menu_with_children(menu_id: int):
        menu = await menu_repository.get(id=menu_id)
        menu_dict = menu.to_dict()
        session = get_session()
        child_menus = (
            (
                await session.execute(
                    select(Menu).where(Menu.parent_id == menu_id).order_by(Menu.order)
                )
            )
            .scalars()
            .all()
        )
        menu_dict["children"] = [
            await get_menu_with_children(child.id) for child in child_menus
        ]
        return menu_dict

    session = get_session()
    parent_menus = (
        (
            await session.execute(
                select(Menu).where(Menu.parent_id == 0).order_by(Menu.order)
            )
        )
        .scalars()
        .all()
    )
    res_menu = [await get_menu_with_children(menu.id) for menu in parent_menus]
    result = SuccessExtra(
        data=res_menu, total=len(res_menu), page=page, page_size=page_size
    )
    return json.loads(result.body)


@router.get("/{menu_id}", summary="查看菜单", response_model=MenuDetailResponse)
async def get_menu(
    menu_id: int,
):
    result_data = await menu_repository.get(id=menu_id)
    result = Success(data=result_data)
    return json.loads(result.body)


@router.post("", summary="创建菜单", response_model=ResponseBase[None])
async def create_menu(
    menu_in: MenuCreate,
):
    await menu_repository.create(obj_in=menu_in)
    result = Success(msg="Created Success")
    return json.loads(result.body)


@router.put("/{menu_id}", summary="更新菜单", response_model=ResponseBase[None])
async def update_menu(
    menu_id: int,
    menu_in: MenuUpdate,
):
    menu_in.id = menu_id
    await menu_repository.update(id=menu_in.id, obj_in=menu_in)
    result = Success(msg="Updated Success")
    return json.loads(result.body)


@router.delete("/{menu_id}", summary="删除菜单", response_model=ResponseBase[None])
async def delete_menu(
    menu_id: int,
):
    child_menu_count = await get_session().scalar(
        select(func.count()).select_from(Menu).where(Menu.parent_id == menu_id)
    )
    if child_menu_count > 0:
        result = Fail(msg="Cannot delete a menu with child menus")
        return json.loads(result.body)
    await menu_repository.remove(id=menu_id)
    result = Success(msg="Deleted Success")
    return json.loads(result.body)
