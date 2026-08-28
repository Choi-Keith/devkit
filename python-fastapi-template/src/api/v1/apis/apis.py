import json

from fastapi import APIRouter, Query

from models.admin import Api
from repositories.api import api_repository
from schemas import Success, SuccessExtra
from schemas.apis import ApiCreate, ApiUpdate
from schemas.response import ApiInfo, ApiListResponse, ResponseBase
from utils.cache import clear_user_perms_cache_all

router = APIRouter()


@router.get("", summary="查看API列表", response_model=ApiListResponse)
async def list_api(
    page: int = Query(1, description="页码"),
    page_size: int = Query(10, description="每页数量"),
    path: str = Query(None, description="API路径"),
    summary: str = Query(None, description="API简介"),
    tags: str = Query(None, description="API模块"),
):
    conditions = []
    if path:
        conditions.append(Api.path.contains(path))
    if summary:
        conditions.append(Api.summary.contains(summary))
    if tags:
        conditions.append(Api.tags.contains(tags))
    total, api_objs = await api_repository.list(
        page=page, page_size=page_size, search=conditions, order=["tags", "id"]
    )
    data = [obj.to_dict() for obj in api_objs]
    result = SuccessExtra(data=data, total=total, page=page, page_size=page_size)
    return json.loads(result.body)


@router.get("/{api_id}", summary="查看Api", response_model=ResponseBase[ApiInfo])
async def get_api(
    api_id: int,
):
    api_obj = await api_repository.get(id=api_id)
    data = api_obj.to_dict()
    result = Success(data=data)
    return json.loads(result.body)


@router.post("", summary="创建Api", response_model=ResponseBase[None])
async def create_api(
    api_in: ApiCreate,
):
    await api_repository.create(obj_in=api_in)
    await clear_user_perms_cache_all()
    result = Success(msg="Created Successfully")
    return json.loads(result.body)


@router.put("/{api_id}", summary="更新Api", response_model=ResponseBase[None])
async def update_api(
    api_id: int,
    api_in: ApiUpdate,
):
    api_in.id = api_id
    await api_repository.update(id=api_in.id, obj_in=api_in)
    await clear_user_perms_cache_all()
    result = Success(msg="Update Successfully")
    return json.loads(result.body)


@router.delete("/{api_id}", summary="删除Api", response_model=ResponseBase[None])
async def delete_api(
    api_id: int,
):
    await api_repository.remove(id=api_id)
    await clear_user_perms_cache_all()
    result = Success(msg="Deleted Success")
    return json.loads(result.body)


@router.post("/refresh", summary="刷新API列表", response_model=ResponseBase[None])
async def refresh_api():
    await api_repository.refresh_api()
    result = Success(msg="OK")
    return json.loads(result.body)
