import { HttpStatus } from '@nestjs/common';

/**
 * 业务错误码条目:业务 code + 默认英文消息 + 默认 HTTP 状态映射。
 * 统一响应体为 { code, message, data },HTTP 状态与 code 保持语义一致。
 */
export interface Ecode {
  code: number;
  message: string;
  httpStatus: HttpStatus;
}

export const ecode = (code: number, message: string, httpStatus: HttpStatus): Ecode => ({
  code,
  message,
  httpStatus,
});
