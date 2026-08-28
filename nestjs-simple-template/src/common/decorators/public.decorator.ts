import { SetMetadata } from '@nestjs/common';

export const IS_PUBLIC_KEY = 'isPublic';

/** 标记免认证接口(全局 JwtAuthGuard 会跳过) */
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true);
