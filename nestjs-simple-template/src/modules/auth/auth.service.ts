import * as argon2 from 'argon2';
import { randomUUID } from 'node:crypto';
import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { JwtService, JwtSignOptions } from '@nestjs/jwt';
import { CommonEcode, UserEcode } from '../../common/ecode';
import { BusinessException } from '../../common/exceptions/business.exception';
import { User } from '../user/user.entity';
import { UserService } from '../user/user.service';
import { LoginDto } from './dto/login.dto';
import { RegisterDto } from './dto/register.dto';
import { JwtPayload } from './strategies/jwt.strategy';

export interface TokenPair {
  accessToken: string;
  refreshToken: string;
}

@Injectable()
export class AuthService {
  constructor(
    private readonly userService: UserService,
    private readonly jwtService: JwtService,
    private readonly config: ConfigService,
  ) {}

  async register(dto: RegisterDto): Promise<TokenPair> {
    const user = await this.userService.create(dto);
    return this.generateTokens(user);
  }

  async login(dto: LoginDto): Promise<TokenPair> {
    const user = await this.userService.findByEmailWithSecrets(dto.email);
    if (!user || !(await argon2.verify(user.password, dto.password))) {
      throw new BusinessException(UserEcode.INVALID_CREDENTIALS);
    }
    return this.generateTokens(user);
  }

  /** refresh token 轮换:校验签名与库存哈希,通过后签发新对并轮换 */
  async refresh(refreshToken: string): Promise<TokenPair> {
    let payload: JwtPayload;
    try {
      payload = await this.jwtService.verifyAsync<JwtPayload>(refreshToken, {
        secret: this.config.get<string>('jwt.refreshSecret'),
      });
    } catch {
      throw new BusinessException(CommonEcode.UNAUTHORIZED, 'invalid refresh token');
    }
    const user = await this.userService.findByEmailWithSecrets(payload.email);
    if (!user) {
      throw new BusinessException(UserEcode.USER_NOT_FOUND);
    }
    if (!user.refreshToken || !(await argon2.verify(user.refreshToken, refreshToken))) {
      throw new BusinessException(CommonEcode.UNAUTHORIZED, 'invalid refresh token');
    }
    return this.generateTokens(user);
  }

  async logout(userId: string): Promise<void> {
    await this.userService.setRefreshToken(userId, null);
  }

  private async generateTokens(user: User): Promise<TokenPair> {
    const payload: JwtPayload = { sub: user.id, email: user.email, roles: user.roles };
    const [accessToken, refreshToken] = await Promise.all([
      this.jwtService.signAsync(payload, {
        secret: this.config.get<string>('jwt.accessSecret'),
        expiresIn: this.config.get<string>('jwt.accessExpiresIn') as JwtSignOptions['expiresIn'],
      }),
      this.jwtService.signAsync(
        // jti 保证每次签发的 refresh token 唯一,否则同一秒内 payload 相同、轮换后旧 token 仍与库存哈希匹配
        { ...payload, jti: randomUUID() },
        {
          secret: this.config.get<string>('jwt.refreshSecret'),
          expiresIn: this.config.get<string>('jwt.refreshExpiresIn') as JwtSignOptions['expiresIn'],
        },
      ),
    ]);
    await this.userService.setRefreshToken(user.id, refreshToken);
    return { accessToken, refreshToken };
  }
}
