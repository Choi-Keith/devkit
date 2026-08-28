// Package main 是服务的入口。
package main

import (
	"flag"

	"github.com/Choi-Keith/golang-common/logger"
	"github.com/Choi-Keith/golang-common/server"

	"golang-simple-template/internal/config"
	"golang-simple-template/internal/container"
	"golang-simple-template/internal/router"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置(此时 logger 还未初始化,失败只能 panic)
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("starting service...")

	// 初始化依赖容器
	c, err := container.New(cfg)
	if err != nil {
		logger.Fatalf("failed to initialize container: %v", err)
	}
	defer c.Close()

	// 注册路由
	r := router.Setup(cfg, c)

	// 创建并启动 HTTP 服务(优雅退出)
	srv := server.NewHTTPServer(server.HTTPConfig{
		Host:       cfg.Server.Host,
		Port:       cfg.Server.Port,
		Production: cfg.Server.Production,
	})
	srv.OnShutdown(func() {
		logger.Info("closing resources...")
		c.Close()
	})

	logger.Infof("server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := srv.Run(r); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
