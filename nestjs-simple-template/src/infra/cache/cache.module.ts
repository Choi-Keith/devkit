import KeyvRedis from '@keyv/redis';
import { CacheModule as NestCacheModule } from '@nestjs/cache-manager';
import { Global, Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import Keyv from 'keyv';
import type { RedisClientOptions } from 'redis';
import { CacheService } from './cache.service';

const REDIS_CONNECT_TIMEOUT_MS = 2000;

/**
 * 全局缓存模块:默认使用 Redis(keyv + @keyv/redis),连不上时降级为内存缓存,启动不崩。
 * 业务代码注入 CacheService(已内置容错)即可,无需关心底层实现。
 *
 * 注:cache-manager v6 生态已迁移到 Keyv,`cache-manager-redis-yet` 已废弃,
 * 且与 @nestjs/cache-manager v3 的 Keyv 包装不兼容(Invalid storage adapter),故使用官方 @keyv/redis。
 */
@Global()
@Module({
  imports: [
    NestCacheModule.registerAsync({
      inject: [ConfigService],
      useFactory: async (config: ConfigService) => {
        const ttl = config.get<number>('redis.ttl') ?? 60000;
        const host = config.get<string>('redis.host') ?? 'localhost';
        const port = config.get<number>('redis.port') ?? 6379;
        const password = config.get<string>('redis.password');
        const db = config.get<number>('redis.db') ?? 0;
        try {
          const redisOptions: RedisClientOptions = {
            socket: {
              host,
              port,
              connectTimeout: REDIS_CONNECT_TIMEOUT_MS,
              reconnectStrategy: () => false,
            },
            password: password || undefined,
            database: db,
          };
          const keyv = new Keyv({
            store: new KeyvRedis(redisOptions),
            ttl,
            namespace: 'cache',
          });
          keyv.on('error', (err) => {
            // 运行期 Redis 抖动只告警,不中断业务
            console.warn(`[CacheModule] redis error: ${err.message}`);
          });
          // 启动时探测连通性,失败则降级内存缓存
          const probe = await Promise.race([
            keyv.set('cache:ping', '1', REDIS_CONNECT_TIMEOUT_MS),
            new Promise<boolean>((resolve) =>
              setTimeout(() => resolve(false), REDIS_CONNECT_TIMEOUT_MS + 1000),
            ),
          ]);
          if (!probe) {
            throw new Error('redis probe failed');
          }
          return { stores: [keyv], ttl };
        } catch (err) {
          console.warn(
            `[CacheModule] redis unavailable (${err instanceof Error ? err.message : String(err)}), fallback to in-memory cache`,
          );
          return { ttl };
        }
      },
    }),
  ],
  providers: [CacheService],
  exports: [NestCacheModule, CacheService],
})
export class CacheModule {}
