import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Inject, Injectable } from '@nestjs/common';
import { Cache } from 'cache-manager';

/**
 * 注入式缓存服务:读写失败时静默降级(返回 null / 忽略),不影响主流程。
 */
@Injectable()
export class CacheService {
  constructor(@Inject(CACHE_MANAGER) private readonly cache: Cache) {}

  async get<T>(key: string): Promise<T | null> {
    try {
      return (await this.cache.get<T>(key)) ?? null;
    } catch {
      return null;
    }
  }

  async set<T>(key: string, value: T, ttl?: number): Promise<void> {
    try {
      await this.cache.set(key, value, ttl);
    } catch {
      // 缓存写入失败不影响业务
    }
  }

  async del(key: string): Promise<void> {
    try {
      await this.cache.del(key);
    } catch {
      // 缓存删除失败不影响业务
    }
  }
}
