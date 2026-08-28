import * as argon2 from 'argon2';
import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { UserEcode } from '../../common/ecode';
import { BusinessException } from '../../common/exceptions/business.exception';
import { Paginated, PaginationQueryDto, paginate } from '../../common/dto/pagination.dto';
import { Role } from '../../common/enums/role.enum';
import { CreateUserDto } from './dto/create-user.dto';
import { UpdateUserDto } from './dto/update-user.dto';
import { User } from './user.entity';

@Injectable()
export class UserService {
  constructor(
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
  ) {}

  async create(dto: CreateUserDto, roles: Role[] = [Role.USER]): Promise<User> {
    const exists = await this.userRepository.findOneBy({ email: dto.email });
    if (exists) {
      throw new BusinessException(UserEcode.EMAIL_EXISTS);
    }
    const user = this.userRepository.create({
      ...dto,
      password: await argon2.hash(dto.password),
      roles,
    });
    return this.userRepository.save(user);
  }

  async findAll(query: PaginationQueryDto): Promise<Paginated<User>> {
    const [list, total] = await this.userRepository.findAndCount({
      skip: query.skip,
      take: query.pageSize,
      order: { createdAt: 'DESC' },
    });
    return paginate(list, total, query);
  }

  async findOne(id: string): Promise<User> {
    const user = await this.userRepository.findOneBy({ id });
    if (!user) {
      throw new BusinessException(UserEcode.USER_NOT_FOUND);
    }
    return user;
  }

  /** 登录专用:带出 password / refreshToken 哈希 */
  async findByEmailWithSecrets(email: string): Promise<User | null> {
    return this.userRepository
      .createQueryBuilder('user')
      .addSelect(['user.password', 'user.refreshToken'])
      .where('user.email = :email', { email })
      .getOne();
  }

  async update(id: string, dto: UpdateUserDto): Promise<User> {
    const user = await this.findOne(id);
    if (dto.email && dto.email !== user.email) {
      const exists = await this.userRepository.findOneBy({ email: dto.email });
      if (exists) {
        throw new BusinessException(UserEcode.EMAIL_EXISTS);
      }
    }
    if (dto.password) {
      dto.password = await argon2.hash(dto.password);
    }
    const merged = this.userRepository.merge(user, dto);
    return this.userRepository.save(merged);
  }

  async remove(id: string): Promise<void> {
    const user = await this.findOne(id);
    await this.userRepository.remove(user);
  }

  /** 存储 / 清除 refresh token 哈希(登出时传 null) */
  async setRefreshToken(id: string, refreshToken: string | null): Promise<void> {
    const hash = refreshToken ? await argon2.hash(refreshToken) : null;
    await this.userRepository.update(id, { refreshToken: hash });
  }
}
