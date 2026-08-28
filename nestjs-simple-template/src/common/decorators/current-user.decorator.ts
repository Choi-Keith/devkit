import { ExecutionContext, createParamDecorator } from '@nestjs/common';
import { Role } from '../enums/role.enum';

export interface AuthUser {
  id: string;
  email: string;
  roles: Role[];
}

/** 获取当前登录用户(JwtStrategy.validate 的返回值) */
export const CurrentUser = createParamDecorator(
  (data: keyof AuthUser | undefined, ctx: ExecutionContext): AuthUser | unknown => {
    const req = ctx.switchToHttp().getRequest<{ user: AuthUser }>();
    return data ? req.user?.[data] : req.user;
  },
);
