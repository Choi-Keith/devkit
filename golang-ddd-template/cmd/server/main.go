// Command server is the application entry point: it loads configuration,
// wires the dependency graph (infra -> domain -> application -> adapter)
// and runs the HTTP server with graceful shutdown.
package main

import (
	"flag"

	"go.uber.org/zap"

	"github.com/Choi-Keith/golang-common/conf"
	"github.com/Choi-Keith/golang-common/database"
	"github.com/Choi-Keith/golang-common/health"
	"github.com/Choi-Keith/golang-common/logger"
	"github.com/Choi-Keith/golang-common/server"

	httpadapter "github.com/Choi-Keith/golang-ddd-template/internal/adapter/http"
	"github.com/Choi-Keith/golang-ddd-template/internal/adapter/http/order"
	"github.com/Choi-Keith/golang-ddd-template/internal/adapter/http/user"
	"github.com/Choi-Keith/golang-ddd-template/internal/application/order"
	"github.com/Choi-Keith/golang-ddd-template/internal/application/user"
	"github.com/Choi-Keith/golang-ddd-template/internal/infra/db/order"
	"github.com/Choi-Keith/golang-ddd-template/internal/infra/db/user"
)

// Version is set via -ldflags at build time.
var Version = "dev"

// Config is the root configuration of the service, loaded from
// configs/config.yaml (overridable via environment variables).
type Config struct {
	Server   server.HTTPConfig `mapstructure:"server"`
	Log      logger.Config     `mapstructure:"log"`
	Database database.Config   `mapstructure:"database"`
}

// @title			golang-ddd-template API
// @version		1.0
// @description	基于 golang-common 的 DDD 脚手架服务
// @host			localhost:8080
// @BasePath		/api/v1
func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	if err := conf.Init(*configPath); err != nil {
		logger.Fatal("load config", zap.Error(err))
	}
	var cfg Config
	if err := conf.Unmarshal(&cfg); err != nil {
		logger.Fatal("parse config", zap.Error(err))
	}
	if err := logger.Init(&cfg.Log); err != nil {
		logger.Fatal("init logger", zap.Error(err))
	}

	// infra
	gormDB, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatal("connect database", zap.Error(err))
	}
	userRepo := userdb.NewMysqlUserRepo(gormDB)
	orderRepo := orderdb.NewMysqlOrderRepo(gormDB)

	// application -> adapter
	userSvc := userapp.NewUserService(userRepo)
	userHandler := userhttp.NewUserHandler(userSvc)
	orderSvc := orderapp.NewOrderService(orderRepo)
	orderHandler := orderhttp.NewOrderHandler(orderSvc)
	healthHandler := health.NewHealthHandler(Version).
		AddChecker("database", health.NewChecker(database.HealthChecker(gormDB)))

	router := httpadapter.NewRouter(userHandler, orderHandler, healthHandler)

	srv := server.NewHTTPServer(cfg.Server)
	srv.OnShutdown(func() {
		if err := database.Close(gormDB); err != nil {
			logger.Error("close database", zap.Error(err))
		}
	})

	if err := srv.Run(router); err != nil {
		logger.Fatal("server exited", zap.Error(err))
	}
}
