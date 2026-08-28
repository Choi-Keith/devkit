from typing import Optional

from sqlalchemy import select

from core.crud import CRUDBase
from models.admin import Menu
from schemas.menus import MenuCreate, MenuUpdate


class MenuRepository(CRUDBase[Menu, MenuCreate, MenuUpdate]):
    def __init__(self):
        super().__init__(model=Menu)

    async def get_by_menu_path(self, path: str) -> Optional["Menu"]:
        result = await self.session.execute(select(Menu).where(Menu.path == path))
        return result.scalar_one_or_none()


menu_repository = MenuRepository()
