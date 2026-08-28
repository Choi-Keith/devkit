import json

from fastapi import APIRouter, Query

from repositories.dept import dept_repository
from schemas import Success
from schemas.depts import DeptCreate, DeptUpdate
from schemas.response import DeptDetailResponse, DeptListResponse, ResponseBase

router = APIRouter()


@router.get("", summary="查看部门列表", response_model=DeptListResponse)
async def list_dept(
    name: str = Query(None, description="部门名称"),
):
    dept_tree = await dept_repository.get_dept_tree(name)
    result = Success(data=dept_tree)
    return json.loads(result.body)


@router.get("/{dept_id}", summary="查看部门", response_model=DeptDetailResponse)
async def get_dept(
    dept_id: int,
):
    dept_obj = await dept_repository.get(id=dept_id)
    data = dept_obj.to_dict()
    result = Success(data=data)
    return json.loads(result.body)


@router.post("", summary="创建部门", response_model=ResponseBase[None])
async def create_dept(
    dept_in: DeptCreate,
):
    await dept_repository.create_dept(obj_in=dept_in)
    result = Success(msg="Created Successfully")
    return json.loads(result.body)


@router.put("/{dept_id}", summary="更新部门", response_model=ResponseBase[None])
async def update_dept(
    dept_id: int,
    dept_in: DeptUpdate,
):
    dept_in.id = dept_id
    await dept_repository.update_dept(obj_in=dept_in)
    result = Success(msg="Update Successfully")
    return json.loads(result.body)


@router.delete("/{dept_id}", summary="删除部门", response_model=ResponseBase[None])
async def delete_dept(
    dept_id: int,
):
    await dept_repository.delete_dept(dept_id=dept_id)
    result = Success(msg="Deleted Success")
    return json.loads(result.body)
