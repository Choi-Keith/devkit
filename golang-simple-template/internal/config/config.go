// Package config 提供配置加载与定义。
package config

import (
	"github.com/Choi-Keith/golang-common/cache"
	"github.com/Choi-Keith/golang-common/conf"
	"github.com/Choi-Keith/golang-common/database"
	"github.com/Choi-Keith/golang-common/logger"
)

// Config 服务配置
type Config struct {
	Server   ServerConfig      `mapstructure:"server"`
	Database database.Config   `mapstructure:"database"`
	Redis    cache.RedisConfig `mapstructure:"redis"`
	Log      logger.Config     `mapstructure:"log"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Production bool   `mapstructure:"production"`
}

// Load 从指定路径加载配置,未设置的项使用组件默认值
func Load(path string) (*Config, error) {
	if err := conf.Init(path); err != nil {
		return nil, err
	}

	cfg := &Config{
		Database: database.DefaultConfig(),
		Redis:    cache.DefaultRedisConfig(),
		Log:      *logger.DefaultConfig(),
	}
	if err := conf.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
