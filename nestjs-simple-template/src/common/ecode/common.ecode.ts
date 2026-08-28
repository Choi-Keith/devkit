import { HttpStatus } from '@nestjs/common';
import { ecode } from './ecode.types';

/**
 * 通用错误码,占 1000-19999 段(0 为成功)。
 */
export const CommonEcode = {
  SUCCESS: ecode(0, 'success', HttpStatus.OK),
  PARAM_INVALID: ecode(1000, 'invalid parameters', HttpStatus.BAD_REQUEST),
  UNAUTHORIZED: ecode(1001, 'unauthorized', HttpStatus.UNAUTHORIZED),
  FORBIDDEN: ecode(1002, 'forbidden', HttpStatus.FORBIDDEN),
  NOT_FOUND: ecode(1003, 'resource not found', HttpStatus.NOT_FOUND),
  TOO_MANY_REQUESTS: ecode(1004, 'too many requests', HttpStatus.TOO_MANY_REQUESTS),
  INTERNAL_ERROR: ecode(1500, 'internal server error', HttpStatus.INTERNAL_SERVER_ERROR),
} as const;
