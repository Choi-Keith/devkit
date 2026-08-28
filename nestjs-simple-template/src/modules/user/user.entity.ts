import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm';
import { Role } from '../../common/enums/role.enum';

@Entity('users')
export class User {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  email: string;

  @Column()
  name: string;

  /** argon2id 哈希,默认查询不返回 */
  @Column({ select: false })
  password: string;

  @Column({ type: 'json' })
  roles: Role[];

  /** refresh token 的 argon2 哈希,用于轮换校验;null 表示已登出 */
  @Column({ type: 'varchar', length: 255, nullable: true, select: false })
  refreshToken: string | null;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
