import { ArgumentsHost, Catch, ExceptionFilter, HttpException, HttpStatus } from '@nestjs/common';
import { ThrottlerException } from '@nestjs/throttler';
import { Response } from 'express';
import { PinoLogger } from 'nestjs-pino';
import { CommonEcode } from '../ecode';
import { BusinessException } from '../exceptions/business.exception';

interface ErrorBody {
  code: number;
  message: string;
  data: null;
}

/**
 * 全局异常过滤器:统一输出 { code, message, data: null }。
 * - BusinessException:使用其携带的业务 code 与 httpStatus
 * - ThrottlerException:映射为 1004
 * - HttpException(含 class-validator 的 BadRequestException):按 HTTP 状态映射通用码,提取校验消息
 * - 未知错误:兜底 1500,生产环境不泄露内部细节
 */
@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  constructor(private readonly logger: PinoLogger) {
    this.logger.setContext(GlobalExceptionFilter.name);
  }

  catch(exception: unknown, host: ArgumentsHost): void {
    const res = host.switchToHttp().getResponse<Response>();
    const body = this.toErrorBody(exception);
    res.status(this.httpStatusOf(body.code, exception)).json(body);
  }

  private toErrorBody(exception: unknown): ErrorBody {
    if (exception instanceof BusinessException) {
      return { code: exception.code, message: exception.message, data: null };
    }
    if (exception instanceof ThrottlerException) {
      return {
        code: CommonEcode.TOO_MANY_REQUESTS.code,
        message: CommonEcode.TOO_MANY_REQUESTS.message,
        data: null,
      };
    }
    if (exception instanceof HttpException) {
      const status = exception.getStatus();
      return {
        code: this.codeOf(status),
        message: this.messageOf(exception),
        data: null,
      };
    }
    this.logger.error(exception instanceof Error ? exception.stack : String(exception));
    return {
      code: CommonEcode.INTERNAL_ERROR.code,
      message:
        process.env.NODE_ENV === 'production'
          ? CommonEcode.INTERNAL_ERROR.message
          : exception instanceof Error
            ? exception.message
            : CommonEcode.INTERNAL_ERROR.message,
      data: null,
    };
  }

  private httpStatusOf(code: number, exception: unknown): number {
    if (exception instanceof HttpException) {
      return exception.getStatus();
    }
    return code === CommonEcode.SUCCESS.code ? HttpStatus.OK : HttpStatus.INTERNAL_SERVER_ERROR;
  }

  private codeOf(status: number): number {
    switch (status) {
      case HttpStatus.BAD_REQUEST:
        return CommonEcode.PARAM_INVALID.code;
      case HttpStatus.UNAUTHORIZED:
        return CommonEcode.UNAUTHORIZED.code;
      case HttpStatus.FORBIDDEN:
        return CommonEcode.FORBIDDEN.code;
      case HttpStatus.NOT_FOUND:
        return CommonEcode.NOT_FOUND.code;
      case HttpStatus.TOO_MANY_REQUESTS:
        return CommonEcode.TOO_MANY_REQUESTS.code;
      default:
        return CommonEcode.INTERNAL_ERROR.code;
    }
  }

  /** class-validator 的 BadRequestException 响应体中 message 为数组,拼接后返回 */
  private messageOf(exception: HttpException): string {
    const body = exception.getResponse();
    if (typeof body === 'object' && body !== null && 'message' in body) {
      const message = (body as { message: unknown }).message;
      if (Array.isArray(message)) {
        return message.join('; ');
      }
      if (typeof message === 'string') {
        return message;
      }
    }
    return exception.message;
  }
}
