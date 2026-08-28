from sqlalchemy import delete, select

from core.crud import CRUDBase
from log import logger
from models.admin import Dept, DeptClosure
from schemas.depts import DeptCreate, DeptUpdate


class DeptRepository(CRUDBase[Dept, DeptCreate, DeptUpdate]):
    def __init__(self):
        super().__init__(model=Dept)

    async def get_dept_tree(self, name):
        # 获取所有未被软删除的部门
        conditions = [Dept.is_deleted.is_(False)]
        if name:
            conditions.append(Dept.name.contains(name))
        result = await self.session.execute(
            select(Dept).where(*conditions).order_by(Dept.order)
        )
        all_depts = list(result.scalars().all())

        # 辅助函数，用于递归构建部门树
        def build_tree(parent_id):
            return [
                {
                    "id": dept.id,
                    "name": dept.name,
                    "desc": dept.desc,
                    "order": dept.order,
                    "parent_id": dept.parent_id,
                    "children": build_tree(dept.id),  # 递归构建子部门
                }
                for dept in all_depts
                if dept.parent_id == parent_id
            ]

        # 从顶级部门（parent_id=0）开始构建部门树
        dept_tree = build_tree(0)
        return dept_tree

    async def get_dept_info(self):
        pass

    async def update_dept_closure(self, obj: Dept):
        result = await self.session.execute(
            select(DeptClosure).where(DeptClosure.descendant == obj.parent_id)
        )
        parent_depts = list(result.scalars().all())
        for i in parent_depts:
            logger.debug(
                f"Processing dept closure: ancestor={i.ancestor}, descendant={i.descendant}"
            )
        dept_closure_objs: list[DeptClosure] = []
        # 插入父级关系
        for item in parent_depts:
            dept_closure_objs.append(
                DeptClosure(
                    ancestor=item.ancestor,
                    descendant=obj.id,
                    level=item.level + 1,
                )
            )
        # 插入自身x
        dept_closure_objs.append(
            DeptClosure(ancestor=obj.id, descendant=obj.id, level=0)
        )
        # 创建关系
        self.session.add_all(dept_closure_objs)
        await self.session.flush()

    async def create_dept(self, obj_in: DeptCreate):
        session = self.session
        try:
            # 创建
            if obj_in.parent_id != 0:
                await self.get(id=obj_in.parent_id)
            new_obj = Dept(
                **{
                    k: v
                    for k, v in obj_in.model_dump().items()
                    if hasattr(Dept, k)
                }
            )
            session.add(new_obj)
            await session.flush()  # 获取自增 ID
            await self.update_dept_closure(new_obj)
            await session.commit()
        except Exception:
            await session.rollback()
            raise

    async def update_dept(self, obj_in: DeptUpdate):
        session = self.session
        try:
            dept_obj = await self.get(id=obj_in.id)
            # 更新部门关系
            if dept_obj.parent_id != obj_in.parent_id:
                await session.execute(
                    delete(DeptClosure).where(DeptClosure.ancestor == dept_obj.id)
                )
                await session.execute(
                    delete(DeptClosure).where(DeptClosure.descendant == dept_obj.id)
                )
                await self.update_dept_closure(dept_obj)
            # 更新部门信息
            dept_obj.update_from_dict(obj_in.model_dump(exclude_unset=True))
            await session.commit()
        except Exception:
            await session.rollback()
            raise

    async def delete_dept(self, dept_id: int):
        session = self.session
        try:
            # 删除部门
            obj = await self.get(id=dept_id)
            obj.is_deleted = True
            # 删除关系
            await session.execute(
                delete(DeptClosure).where(DeptClosure.descendant == dept_id)
            )
            await session.commit()
        except Exception:
            await session.rollback()
            raise


dept_repository = DeptRepository()
