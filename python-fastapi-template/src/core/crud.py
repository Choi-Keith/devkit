from typing import Any, Generic, NewType, TypeVar

from pydantic import BaseModel as PydanticBaseModel
from sqlalchemy import desc, func, select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.sql.elements import ColumnElement

from core.database import get_session
from core.exceptions import NotFoundError
from models.base import BaseModel

Total = NewType("Total", int)
ModelType = TypeVar("ModelType", bound=BaseModel)
CreateSchemaType = TypeVar("CreateSchemaType", bound=PydanticBaseModel)
UpdateSchemaType = TypeVar("UpdateSchemaType", bound=PydanticBaseModel)


class CRUDBase(Generic[ModelType, CreateSchemaType, UpdateSchemaType]):
    def __init__(self, model: type[ModelType]):
        self.model = model

    @property
    def session(self) -> AsyncSession:
        return get_session()

    async def get(self, id: int) -> ModelType:
        obj = await self.session.get(self.model, id)
        if obj is None:
            raise NotFoundError(f"{self.model.__name__} id={id} 不存在")
        return obj

    async def list(
        self,
        page: int,
        page_size: int,
        search: list[ColumnElement] | None = None,
        order: list[str] | None = None,
    ) -> tuple[Total, list[ModelType]]:
        conditions = search or []
        session = self.session
        total = await session.scalar(
            select(func.count()).select_from(self.model).where(*conditions)
        )
        stmt = select(self.model).where(*conditions)
        for field in order or []:
            if field.startswith("-"):
                stmt = stmt.order_by(desc(getattr(self.model, field[1:])))
            else:
                stmt = stmt.order_by(getattr(self.model, field))
        stmt = stmt.offset((page - 1) * page_size).limit(page_size)
        result = await session.execute(stmt)
        return Total(total or 0), list(result.scalars().all())

    async def create(self, obj_in: CreateSchemaType) -> ModelType:
        if isinstance(obj_in, dict):
            obj_dict = obj_in
        else:
            obj_dict = obj_in.model_dump()
        # 过滤掉 schema 中的非模型字段（如 role_ids）
        obj_dict = {k: v for k, v in obj_dict.items() if hasattr(self.model, k)}
        obj = self.model(**obj_dict)
        session = self.session
        session.add(obj)
        await session.commit()
        await session.refresh(obj)
        return obj

    async def update(
        self, id: int, obj_in: UpdateSchemaType | dict[str, Any]
    ) -> ModelType:
        if isinstance(obj_in, dict):
            obj_dict = obj_in
        else:
            obj_dict = obj_in.model_dump(exclude_unset=True, exclude={"id"})
        obj = await self.get(id=id)
        obj.update_from_dict(obj_dict)
        session = self.session
        await session.commit()
        await session.refresh(obj)
        return obj

    async def remove(self, id: int) -> None:
        obj = await self.get(id=id)
        session = self.session
        await session.delete(obj)
        await session.commit()
