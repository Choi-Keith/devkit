// Package configs 定义应用配置结构并负责加载(Viper + 环境变量覆盖)。
package configs

import (
	"fmt"
	"time"

	"github.com/Choi-Keith/golang-common/cache"
	"github.com/Choi-Keith/golang-common/conf"
	"github.com/Choi-Keith/golang-common/database"
	"github.com/Choi-Keith/golang-common/logger"
	"github.com/Choi-Keith/golang-common/server"
)

// Config 应用根配置,加载自 configs/config.yaml,环境变量可覆盖
// (如 SERVER_PORT 覆盖 server.port)。
type Config struct {
	Server   server.HTTPConfig `mapstructure:"server"`
	Log      logger.Config     `mapstructure:"log"`
	Database database.Config   `mapstructure:"database"`
	Redis    cache.RedisConfig `mapstructure:"redis"`
	JWT      JWTConfig         `mapstructure:"jwt"`
}

// JWTConfig 认证配置。access token 默认 2 小时,refresh token 默认 7 天。
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	Issuer            string        `mapstructure:"issuer"`
	AccessExpiration  time.Duration `mapstructure:"accessExpiration"`
	RefreshExpiration time.Duration `mapstructure:"refreshExpiration"`
}

// Load 读取配置文件并反序列化为 Config,同时填充默认值。
func Load(path string) (*Config, error) {
	if err := conf.Init(path); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	var cfg Config
	if err := conf.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.JWT.AccessExpiration == 0 {
		cfg.JWT.AccessExpiration = 2 * time.Hour
	}
	if cfg.JWT.RefreshExpiration == 0 {
		cfg.JWT.RefreshExpiration = 7 * 24 * time.Hour
	}
	return &cfg, nil
}
