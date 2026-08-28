import { HttpStatus } from '@nestjs/common';
import { ecode } from './ecode.types';

/**
 * user 模块业务错误码,从 20000 开始。
 * 新业务模块按 21000 / 22000 ... 分段,每模块一个 *.ecode.ts 文件。
 */
export const UserEcode = {
  USER_NOT_FOUND: ecode(20000, 'user not found', HttpStatus.NOT_FOUND),
  EMAIL_EXISTS: ecode(20001, 'email already exists', HttpStatus.CONFLICT),
  INVALID_CREDENTIALS: ecode(20002, 'invalid credentials', HttpStatus.UNAUTHORIZED),
} as const;
