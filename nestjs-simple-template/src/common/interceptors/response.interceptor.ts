import { CallHandler, ExecutionContext, Injectable, NestInterceptor } from '@nestjs/common';
import { Request } from 'express';
import { Observable, map } from 'rxjs';
import { CommonEcode } from '../ecode';

/** 不包装的路径前缀:健康检查保持原始结构,便于探针与 terminus 语义 */
const SKIP_PREFIXES = ['/health'];

/**
 * 统一成功响应为 { code: 0, message: 'success', data }。
 * Swagger 中对应结构见 common/dto/response.dto.ts。
 */
@Injectable()
export class ResponseInterceptor<T> implements NestInterceptor<T, unknown> {
  intercept(context: ExecutionContext, next: CallHandler<T>): Observable<unknown> {
    const req = context.switchToHttp().getRequest<Request>();
    const skip = SKIP_PREFIXES.some((prefix) => req.path.startsWith(prefix));
    return next.handle().pipe(
      map((data) => {
        if (skip) {
          return data;
        }
        return {
          code: CommonEcode.SUCCESS.code,
          message: CommonEcode.SUCCESS.message,
          data: data ?? null,
        };
      }),
    );
  }
}
