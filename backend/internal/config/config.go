package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv       string `env:"APP_ENV" envDefault:"development"`
	BackendPort  string `env:"BACKEND_PORT" envDefault:"8080"`
	JWTSecret    string `env:"JWT_SECRET" envDefault:"gbtreehole-dev-secret-change-me"`
	JWTExpireMin int    `env:"JWT_EXPIRE_MIN" envDefault:"10080"`
	// AliasSecret 是按帖派生树洞化名的服务端密钥；留空时回退使用 JWTSecret。
	// 更换后所有历史帖子的化名会整体变化，因此生产环境应单独固定配置。
	AliasSecret string `env:"ALIAS_SECRET" envDefault:""`

	MySQLHost     string `env:"MYSQL_HOST" envDefault:"mysql"`
	MySQLPort     string `env:"MYSQL_PORT" envDefault:"3306"`
	MySQLUser     string `env:"MYSQL_USER" envDefault:"gbtreehole"`
	MySQLPassword string `env:"MYSQL_PASSWORD" envDefault:"gbtreehole123"`
	MySQLDatabase string `env:"MYSQL_DATABASE" envDefault:"gbtreehole"`

	RedisAddr     string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:18401"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
}

// EffectiveAliasSecret 返回用于派生按帖化名的密钥，未单独配置时回退到 JWT 密钥。
func (c *Config) EffectiveAliasSecret() string {
	if c.AliasSecret != "" {
		return c.AliasSecret
	}
	return c.JWTSecret
}
