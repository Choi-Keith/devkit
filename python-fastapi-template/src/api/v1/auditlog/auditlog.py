import json
from datetime import datetime

from fastapi import APIRouter, Query
from sqlalchemy import desc, func, select

from core.database import get_session
from models.admin import AuditLog
from schemas import SuccessExtra
from schemas.response import AuditLogListResponse

router = APIRouter()


@router.get("", summary="查看操作日志", response_model=AuditLogListResponse)
async def get_audit_log_list(
    page: int = Query(1, description="页码"),
    page_size: int = Query(10, description="每页数量"),
    username: str = Query("", description="操作人名称"),
    module: str = Query("", description="功能模块"),
    method: str = Query("", description="请求方法"),
    summary: str = Query("", description="接口描述"),
    status: int = Query(None, description="状态码"),
    start_time: datetime | None = Query(None, description="开始时间"),
    end_time: datetime | None = Query(None, description="结束时间"),
):
    conditions = []
    if username:
        conditions.append(AuditLog.username.ilike(f"%{username}%"))
    if module:
        conditions.append(AuditLog.module.ilike(f"%{module}%"))
    if method:
        conditions.append(AuditLog.method.ilike(f"%{method}%"))
    if summary:
        conditions.append(AuditLog.summary.ilike(f"%{summary}%"))
    if status:
        conditions.append(AuditLog.status == status)
    if start_time and end_time:
        conditions.append(AuditLog.created_at.between(start_time, end_time))
    elif start_time:
        conditions.append(AuditLog.created_at >= start_time)
    elif end_time:
        conditions.append(AuditLog.created_at <= end_time)

    session = get_session()
    total = await session.scalar(
        select(func.count()).select_from(AuditLog).where(*conditions)
    )
    result = await session.execute(
        select(AuditLog)
        .where(*conditions)
        .order_by(desc(AuditLog.created_at))
        .offset((page - 1) * page_size)
        .limit(page_size)
    )
    data = [audit_log.to_dict() for audit_log in result.scalars().all()]
    result = SuccessExtra(data=data, total=total or 0, page=page, page_size=page_size)
    return json.loads(result.body)
