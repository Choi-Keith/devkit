import { HttpException } from '@nestjs/common';
import { Ecode } from '../ecode';

/**
 * 业务异常:携带业务 code 与 httpStatus,由 GlobalExceptionFilter 统一输出
 * { code, message, data: null }。
 */
export class BusinessException extends HttpException {
  readonly code: number;

  constructor(ec: Ecode, message?: string) {
    super(message ?? ec.message, ec.httpStatus);
    this.code = ec.code;
  }
}
