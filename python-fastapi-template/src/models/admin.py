from datetime import datetime

from sqlalchemy import (
    JSON,
    BigInteger,
    Boolean,
    Column,
    ForeignKey,
    Integer,
    String,
    Table,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base, BaseModel, BigIntCompat, TimestampMixin

# 多对多关联表
user_role = Table(
    "user_role",
    Base.metadata,
    Column(
        "user_id",
        BigIntCompat,
        ForeignKey("user.id", ondelete="CASCADE"),
        primary_key=True,
    ),
    Column(
        "role_id",
        BigIntCompat,
        ForeignKey("role.id", ondelete="CASCADE"),
        primary_key=True,
    ),
)

role_menu = Table(
    "role_menu",
    Base.metadata,
    Column(
        "role_id",
        BigIntCompat,
        ForeignKey("role.id", ondelete="CASCADE"),
        primary_key=True,
    ),
    Column(
        "menu_id",
        BigIntCompat,
        ForeignKey("menu.id", ondelete="CASCADE"),
        primary_key=True,
    ),
)

role_api = Table(
    "role_api",
    Base.metadata,
    Column(
        "role_id",
        BigIntCompat,
        ForeignKey("role.id", ondelete="CASCADE"),
        primary_key=True,
    ),
    Column(
        "api_id",
        BigIntCompat,
        ForeignKey("api.id", ondelete="CASCADE"),
        primary_key=True,
    ),
)


class User(BaseModel, TimestampMixin):
    __tablename__ = "user"

    username: Mapped[str] = mapped_column(
        String(20), unique=True, index=True, comment="用户名称"
    )
    alias: Mapped[str | None] = mapped_column(
        String(30), nullable=True, index=True, comment="姓名"
    )
    email: Mapped[str] = mapped_column(
        String(255), unique=True, index=True, comment="邮箱"
    )
    phone: Mapped[str | None] = mapped_column(
        String(20), nullable=True, index=True, comment="电话"
    )
    password: Mapped[str | None] = mapped_column(
        String(128), nullable=True, comment="密码"
    )
    is_active: Mapped[bool] = mapped_column(
        Boolean, default=True, index=True, comment="是否激活"
    )
    is_superuser: Mapped[bool] = mapped_column(
        Boolean, default=False, index=True, comment="是否为超级管理员"
    )
    last_login: Mapped[datetime | None] = mapped_column(
        nullable=True, index=True, comment="最后登录时间"
    )
    roles: Mapped[list["Role"]] = relationship(
        secondary=user_role, back_populates="user_roles", lazy="selectin"
    )
    dept_id: Mapped[int | None] = mapped_column(
        Integer, nullable=True, index=True, comment="部门ID"
    )


class Role(BaseModel, TimestampMixin):
    __tablename__ = "role"

    name: Mapped[str] = mapped_column(
        String(20), unique=True, index=True, comment="角色名称"
    )
    desc: Mapped[str | None] = mapped_column(
        String(500), nullable=True, comment="角色描述"
    )
    user_roles: Mapped[list[User]] = relationship(
        secondary=user_role, back_populates="roles", lazy="selectin"
    )
    menus: Mapped[list["Menu"]] = relationship(
        secondary=role_menu, back_populates="role_menus", lazy="selectin"
    )
    apis: Mapped[list["Api"]] = relationship(
        secondary=role_api, back_populates="role_apis", lazy="selectin"
    )


class Api(BaseModel, TimestampMixin):
    __tablename__ = "api"

    path: Mapped[str] = mapped_column(String(100), index=True, comment="API路径")
    method: Mapped[str] = mapped_column(String(10), index=True, comment="请求方法")
    summary: Mapped[str] = mapped_column(String(500), comment="请求简介")
    tags: Mapped[str] = mapped_column(String(100), comment="API标签")
    role_apis: Mapped[list[Role]] = relationship(
        secondary=role_api, back_populates="apis", lazy="selectin"
    )


class Menu(BaseModel, TimestampMixin):
    __tablename__ = "menu"

    name: Mapped[str] = mapped_column(String(20), comment="菜单名称")
    remark: Mapped[dict | None] = mapped_column(JSON, nullable=True, comment="保留字段")
    menu_type: Mapped[str | None] = mapped_column(
        String(20), nullable=True, comment="菜单类型"
    )
    icon: Mapped[str | None] = mapped_column(
        String(100), nullable=True, comment="菜单图标"
    )
    path: Mapped[str] = mapped_column(String(100), index=True, comment="菜单路径")
    order: Mapped[int] = mapped_column(Integer, default=0, index=True, comment="排序")
    parent_id: Mapped[int] = mapped_column(
        Integer, default=0, index=True, comment="父菜单ID"
    )
    is_hidden: Mapped[bool] = mapped_column(Boolean, default=False, comment="是否隐藏")
    component: Mapped[str] = mapped_column(String(100), comment="组件")
    keepalive: Mapped[bool] = mapped_column(Boolean, default=True, comment="存活")
    redirect: Mapped[str | None] = mapped_column(
        String(100), nullable=True, comment="重定向"
    )
    role_menus: Mapped[list[Role]] = relationship(
        secondary=role_menu, back_populates="menus", lazy="selectin"
    )


class Dept(BaseModel, TimestampMixin):
    __tablename__ = "dept"

    name: Mapped[str] = mapped_column(
        String(20), unique=True, index=True, comment="部门名称"
    )
    desc: Mapped[str | None] = mapped_column(String(500), nullable=True, comment="备注")
    is_deleted: Mapped[bool] = mapped_column(
        Boolean, default=False, index=True, comment="软删除标记"
    )
    order: Mapped[int] = mapped_column(Integer, default=0, index=True, comment="排序")
    parent_id: Mapped[int] = mapped_column(
        Integer, default=0, index=True, comment="父部门ID"
    )


class DeptClosure(BaseModel, TimestampMixin):
    __tablename__ = "dept_closure"

    ancestor: Mapped[int] = mapped_column(Integer, index=True, comment="父代")
    descendant: Mapped[int] = mapped_column(Integer, index=True, comment="子代")
    level: Mapped[int] = mapped_column(Integer, default=0, index=True, comment="深度")


class AuditLog(BaseModel, TimestampMixin):
    __tablename__ = "audit_log"

    user_id: Mapped[int] = mapped_column(Integer, index=True, comment="用户ID")
    username: Mapped[str] = mapped_column(
        String(64), default="", index=True, comment="用户名称"
    )
    module: Mapped[str] = mapped_column(String(64), default="", comment="功能模块")
    summary: Mapped[str] = mapped_column(String(128), default="", comment="请求描述")
    method: Mapped[str] = mapped_column(String(10), default="", comment="请求方法")
    path: Mapped[str] = mapped_column(String(255), default="", comment="请求路径")
    status: Mapped[int] = mapped_column(
        Integer, default=-1, index=True, comment="状态码"
    )
    response_time: Mapped[int] = mapped_column(
        Integer, default=0, index=True, comment="响应时间(单位ms)"
    )
    request_args: Mapped[dict | None] = mapped_column(
        JSON, nullable=True, comment="请求参数"
    )
    response_body: Mapped[dict | None] = mapped_column(
        JSON, nullable=True, comment="返回数据"
    )


class FileMapping(BaseModel, TimestampMixin):
    """文件映射模型 - 管理文件ID和文件信息的映射关系"""

    __tablename__ = "file_mapping"

    file_id: Mapped[str] = mapped_column(
        String(255), unique=True, index=True, comment="文件ID"
    )
    original_filename: Mapped[str] = mapped_column(String(255), comment="原始文件名")
    file_type: Mapped[str] = mapped_column(String(50), comment="文件类型")
    file_size: Mapped[int | None] = mapped_column(
        BigInteger, nullable=True, comment="文件大小(字节)"
    )
    upload_user_id: Mapped[int] = mapped_column(Integer, index=True, comment="上传用户ID")
    file_path: Mapped[str | None] = mapped_column(
        String(500), nullable=True, comment="本地文件路径"
    )
