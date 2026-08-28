import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { App } from 'supertest/types';
import { AppModule } from '../src/app.module';

/**
 * e2e 冒烟:需要真实的 MySQL(或 DB_TYPE=postgres 的 PostgreSQL),
 * 先 `docker compose up -d` 并配置 .env,再执行 `npm run test:e2e`。
 * 未设置 RUN_DB_E2E=1 时自动跳过,保证 CI/本地无依赖时不阻塞。
 */
const describeIfDb = process.env.RUN_DB_E2E === '1' ? describe : describe.skip;

describeIfDb('App (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('/health/live (GET)', () => {
    return request(app.getHttpServer() as App)
      .get('/health/live')
      .expect(200)
      .expect((res) => {
        expect(res.body.status).toBe('ok');
      });
  });
});
