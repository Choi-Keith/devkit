from datetime import datetime
from uuid import UUID, uuid4

from sqlalchemy import BigInteger, DateTime, Integer, Uuid
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

from settings import settings

# SQLite 只有 INTEGER PRIMARY KEY 才自增，BigInteger 需降级变体
BigIntCompat = BigInteger().with_variant(Integer, "sqlite")


class Base(DeclarativeBase):
    pass


class BaseModel(Base):
    __abstract__ = True

    id: Mapped[int] = mapped_column(
        BigIntCompat, primary_key=True, autoincrement=True, index=True
    )

    def to_dict(
        self, m2m: bool = False, exclude_fields: list[str] | None = None
    ) -> dict:
        """序列化为字典；m2m=True 时附带已加载（selectin）的多对多关系。"""
        if exclude_fields is None:
            exclude_fields = []

        d = {}
        for attr in self.__mapper__.column_attrs:
            field = attr.key
            if field not in exclude_fields:
                value = getattr(self, field)
                if isinstance(value, datetime):
                    value = value.strftime(settings.DATETIME_FORMAT)
                d[field] = value

        if m2m:
            for rel in self.__mapper__.relationships:
                if rel.key in exclude_fields:
                    continue
                related = getattr(self, rel.key)
                if related is None:
                    continue
                if rel.uselist:
                    d[rel.key] = [
                        item.to_dict(exclude_fields=exclude_fields) for item in related
                    ]
                else:
                    d[rel.key] = related.to_dict(exclude_fields=exclude_fields)

        return d

    def update_from_dict(self, data: dict) -> "BaseModel":
        """按字典更新字段（仅更新模型上已有的属性）。"""
        for key, value in data.items():
            if hasattr(self, key):
                setattr(self, key, value)
        return self


class UUIDModel:
    uuid: Mapped[UUID] = mapped_column(
        Uuid, unique=True, index=True, default=uuid4
    )


class TimestampMixin:
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.now, index=True
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.now, onupdate=datetime.now, index=True
    )
