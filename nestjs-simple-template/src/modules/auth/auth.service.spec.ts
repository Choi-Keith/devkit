import * as argon2 from 'argon2';
import { ConfigService } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';
import { Test, TestingModule } from '@nestjs/testing';
import { CommonEcode, UserEcode } from '../../common/ecode';
import { Role } from '../../common/enums/role.enum';
import { User } from '../user/user.entity';
import { UserService } from '../user/user.service';
import { AuthService } from './auth.service';

const configValues: Record<string, string> = {
  'jwt.accessSecret': 'access-secret',
  'jwt.accessExpiresIn': '15m',
  'jwt.refreshSecret': 'refresh-secret',
  'jwt.refreshExpiresIn': '7d',
};

describe('AuthService', () => {
  let service: AuthService;
  let userService: jest.Mocked<UserService>;
  let jwtService: jest.Mocked<JwtService>;

  const user: User = {
    id: 'uuid-1',
    email: 'alice@example.com',
    name: 'alice',
    password: '',
    roles: [Role.USER],
    refreshToken: null,
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthService,
        {
          provide: UserService,
          useValue: {
            create: jest.fn(),
            findOne: jest.fn(),
            findByEmailWithSecrets: jest.fn(),
            setRefreshToken: jest.fn(),
          },
        },
        {
          provide: JwtService,
          useValue: { signAsync: jest.fn(), verifyAsync: jest.fn() },
        },
        {
          provide: ConfigService,
          useValue: { get: jest.fn((key: string) => configValues[key]) },
        },
      ],
    }).compile();

    service = module.get(AuthService);
    userService = module.get(UserService);
    jwtService = module.get(JwtService);

    // 每次 generateTokens 依次签发 access / refresh
    jwtService.signAsync
      .mockResolvedValueOnce('access-token')
      .mockResolvedValueOnce('refresh-token');
  });

  describe('register', () => {
    it('应创建用户并返回 token 对,同时持久化 refresh token 哈希', async () => {
      userService.create.mockResolvedValue(user);

      const tokens = await service.register({
        email: user.email,
        name: user.name,
        password: 'Passw0rd!',
      });

      expect(tokens).toEqual({ accessToken: 'access-token', refreshToken: 'refresh-token' });
      expect(userService.setRefreshToken).toHaveBeenCalledWith(user.id, 'refresh-token');
    });

    it('邮箱重复时透传 20001', async () => {
      userService.create.mockRejectedValue(
        Object.assign(new Error('email already exists'), { code: UserEcode.EMAIL_EXISTS.code }),
      );

      await expect(
        service.register({ email: user.email, name: 'x', password: 'Passw0rd!' }),
      ).rejects.toMatchObject({ code: UserEcode.EMAIL_EXISTS.code });
    });
  });

  describe('login', () => {
    it('邮箱密码正确时返回 token 对', async () => {
      user.password = await argon2.hash('Passw0rd!');
      userService.findByEmailWithSecrets.mockResolvedValue(user);

      const tokens = await service.login({ email: user.email, password: 'Passw0rd!' });

      expect(tokens.accessToken).toBe('access-token');
      expect(tokens.refreshToken).toBe('refresh-token');
    });

    it('用户不存在时抛出 20002', async () => {
      userService.findByEmailWithSecrets.mockResolvedValue(null);

      await expect(
        service.login({ email: 'none@example.com', password: 'x' }),
      ).rejects.toMatchObject({ code: UserEcode.INVALID_CREDENTIALS.code });
    });

    it('密码错误时抛出 20002', async () => {
      user.password = await argon2.hash('Passw0rd!');
      userService.findByEmailWithSecrets.mockResolvedValue(user);

      await expect(service.login({ email: user.email, password: 'wrong' })).rejects.toMatchObject({
        code: UserEcode.INVALID_CREDENTIALS.code,
        message: 'invalid credentials',
      });
    });
  });

  describe('refresh', () => {
    it('refresh token 合法时轮换并返回新 token 对', async () => {
      const oldRefresh = 'old-refresh-token';
      user.refreshToken = await argon2.hash(oldRefresh);
      jwtService.verifyAsync.mockResolvedValue({
        sub: user.id,
        email: user.email,
        roles: user.roles,
      });
      userService.findByEmailWithSecrets.mockResolvedValue(user);

      const tokens = await service.refresh(oldRefresh);

      expect(tokens).toEqual({ accessToken: 'access-token', refreshToken: 'refresh-token' });
      expect(userService.setRefreshToken).toHaveBeenCalledWith(user.id, 'refresh-token');
    });

    it('签名非法时抛出 1001', async () => {
      jwtService.verifyAsync.mockRejectedValue(new Error('invalid signature'));

      await expect(service.refresh('bad-token')).rejects.toMatchObject({
        code: CommonEcode.UNAUTHORIZED.code,
      });
    });

    it('与库存哈希不匹配(已吊销/已轮换)时抛出 1001', async () => {
      user.refreshToken = await argon2.hash('another-token');
      jwtService.verifyAsync.mockResolvedValue({
        sub: user.id,
        email: user.email,
        roles: user.roles,
      });
      userService.findByEmailWithSecrets.mockResolvedValue(user);

      await expect(service.refresh('stale-token')).rejects.toMatchObject({
        code: CommonEcode.UNAUTHORIZED.code,
      });
    });
  });

  describe('logout', () => {
    it('应清除库存 refresh token', async () => {
      await service.logout(user.id);
      expect(userService.setRefreshToken).toHaveBeenCalledWith(user.id, null);
    });
  });
});
