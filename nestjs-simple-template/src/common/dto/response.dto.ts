import { ApiProperty } from '@nestjs/swagger';

/**
 * 统一响应体 { code, message, data } 的 Swagger 描述。
 * 实际包装由 ResponseInterceptor 完成。
 */
export class ApiResponseDto<T> {
  @ApiProperty({ example: 0, description: '业务错误码,0 表示成功' })
  code: number;

  @ApiProperty({ example: 'success' })
  message: string;

  data: T;
}

export class ApiErrorResponseDto {
  @ApiProperty({ example: 1000, description: '业务错误码,见 README 错误码段说明' })
  code: number;

  @ApiProperty({ example: 'invalid parameters' })
  message: string;

  @ApiProperty({ example: null, nullable: true })
  data: null;
}
