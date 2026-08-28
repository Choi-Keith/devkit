import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { UserEcode } from '../../common/ecode';
import { BusinessException } from '../../common/exceptions/business.exception';
import { PaginationQueryDto } from '../../common/dto/pagination.dto';
import { Role } from '../../common/enums/role.enum';
import { User } from './user.entity';
import { UserService } from './user.service';

const mockUser = (): User => ({
  id: 'uuid-1',
  email: 'alice@example.com',
  name: 'alice',
  password: '$argon2id$hashed',
  roles: [Role.USER],
  refreshToken: null,
  createdAt: new Date(),
  updatedAt: new Date(),
});

describe('UserService', () => {
  let service: UserService;
  let repository: jest.Mocked<Repository<User>>;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        UserService,
        {
          provide: getRepositoryToken(User),
          useValue: {
            findOneBy: jest.fn(),
            findAndCount: jest.fn(),
            create: jest.fn(),
            save: jest.fn(),
            merge: jest.fn(),
            remove: jest.fn(),
            update: jest.fn(),
          },
        },
      ],
    }).compile();

    service = module.get(UserService);
    repository = module.get(getRepositoryToken(User));
  });

  describe('create', () => {
    it('应创建用户并哈希密码', async () => {
      repository.findOneBy.mockResolvedValue(null);
      repository.create.mockImplementation((entity) => entity as User);
      repository.save.mockImplementation(async (entity) => ({ id: 'uuid-1', ...entity }) as User);

      const user = await service.create({
        email: 'alice@example.com',
        name: 'alice',
        password: 'Passw0rd!',
      });

      expect(user.id).toBe('uuid-1');
      expect(user.roles).toEqual([Role.USER]);
      expect(user.password).not.toBe('Passw0rd!');
      expect(user.password).toMatch(/^\$argon2/);
    });

    it('邮箱已存在时抛出 20001', async () => {
      repository.findOneBy.mockResolvedValue(mockUser());

      await expect(
        service.create({ email: 'alice@example.com', name: 'alice', password: 'Passw0rd!' }),
      ).rejects.toMatchObject({
        code: UserEcode.EMAIL_EXISTS.code,
        message: 'email already exists',
      });
      await expect(
        service.create({ email: 'alice@example.com', name: 'alice', password: 'Passw0rd!' }),
      ).rejects.toBeInstanceOf(BusinessException);
    });
  });

  describe('findAll', () => {
    it('应返回分页结构', async () => {
      const users = [mockUser()];
      repository.findAndCount.mockResolvedValue([users, 1]);

      const query = Object.assign(new PaginationQueryDto(), { page: 2, pageSize: 10 });
      const result = await service.findAll(query);

      expect(result).toEqual({ list: users, total: 1, page: 2, pageSize: 10 });
      expect(repository.findAndCount).toHaveBeenCalledWith({
        skip: 10,
        take: 10,
        order: { createdAt: 'DESC' },
      });
    });
  });

  describe('findOne', () => {
    it('用户不存在时抛出 20000', async () => {
      repository.findOneBy.mockResolvedValue(null);

      await expect(service.findOne('missing')).rejects.toMatchObject({
        code: UserEcode.USER_NOT_FOUND.code,
        message: 'user not found',
      });
    });

    it('用户存在时返回', async () => {
      const user = mockUser();
      repository.findOneBy.mockResolvedValue(user);

      await expect(service.findOne('uuid-1')).resolves.toBe(user);
    });
  });

  describe('update', () => {
    it('更新为新邮箱且邮箱被占用时抛出 20001', async () => {
      repository.findOneBy
        .mockResolvedValueOnce(mockUser()) // findOne(id)
        .mockResolvedValueOnce({ ...mockUser(), id: 'uuid-2' }); // 占用者

      await expect(service.update('uuid-1', { email: 'taken@example.com' })).rejects.toMatchObject({
        code: UserEcode.EMAIL_EXISTS.code,
      });
    });

    it('更新密码时重新哈希', async () => {
      const user = mockUser();
      repository.findOneBy.mockResolvedValue(user);
      repository.merge.mockImplementation((target, dto) => Object.assign(target, dto));
      repository.save.mockImplementation(async (entity) => entity as User);

      const updated = await service.update('uuid-1', { password: 'NewPassw0rd!' });

      expect(updated.password).toMatch(/^\$argon2/);
      expect(updated.password).not.toBe('NewPassw0rd!');
    });
  });

  describe('remove', () => {
    it('删除不存在的用户抛出 20000', async () => {
      repository.findOneBy.mockResolvedValue(null);
      await expect(service.remove('missing')).rejects.toMatchObject({
        code: UserEcode.USER_NOT_FOUND.code,
      });
    });
  });
});
