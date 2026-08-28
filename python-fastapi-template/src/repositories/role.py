from sqlalchemy import func, select

from core.crud import CRUDBase
from models.admin import Api, Menu, Role
from schemas.roles import RoleCreate, RoleUpdate
from utils.cache import clear_user_perms_cache_all


class RoleRepository(CRUDBase[Role, RoleCreate, RoleUpdate]):
    def __init__(self):
        super().__init__(model=Role)

    async def is_exist(self, name: str) -> bool:
        count = await self.session.scalar(
            select(func.count()).select_from(Role).where(Role.name == name)
        )
        return bool(count)

    async def update_roles(
        self, role: Role, menu_ids: list[int], api_infos: list[dict]
    ) -> None:
        session = self.session

        # 更新菜单关联
        menus = []
        if menu_ids:
            result = await session.execute(select(Menu).where(Menu.id.in_(menu_ids)))
            menus = list(result.scalars().all())
        role.menus = menus

        # 更新API关联
        apis = []
        for item in api_infos:
            api_obj = (
                await session.execute(
                    select(Api).where(
                        Api.path == item.get("path"), Api.method == item.get("method")
                    )
                )
            ).scalar_one_or_none()
            if api_obj:
                apis.append(api_obj)
        role.apis = apis

        session.add(role)
        await session.commit()

        # 角色权限变更影响所有绑定该角色的用户，全量清除权限缓存
        await clear_user_perms_cache_all()


role_repository = RoleRepository()
