package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	MySQL      MySQLConfig      `yaml:"mysql"`
	Redis      RedisConfig      `yaml:"redis"`
	RabbitMQ   RabbitMQConfig   `yaml:"rabbitmq"`
	DCC        DCCConfig        `yaml:"dcc"`
	Notify     NotifyConfig     `yaml:"notify"`
	Job        JobConfig        `yaml:"job"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type MySQLConfig struct {
	DSN     string `yaml:"dsn"`
	MaxIdle int    `yaml:"max_idle"`
	MaxOpen int    `yaml:"max_open"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	URL      string `yaml:"url"`
	Exchange string `yaml:"exchange"`
	TeamSuccess struct {
		RoutingKey string `yaml:"routing_key"`
		Queue      string `yaml:"queue"`
	} `yaml:"topic_team_success"`
	TeamRefund struct {
		RoutingKey string `yaml:"routing_key"`
		Queue      string `yaml:"queue"`
	} `yaml:"topic_team_refund"`
}

type DCCConfig struct {
	DowngradeSwitch string `yaml:"downgrade_switch"`
	CutRange        string `yaml:"cut_range"`
	SCBlacklist     string `yaml:"sc_blacklist"`
	CacheSwitch     string `yaml:"cache_switch"`
}

type NotifyConfig struct {
	TopicTeamSuccess string `yaml:"topic_team_success"`
	TopicTeamRefund  string `yaml:"topic_team_refund"`
	HTTPTimeoutSec   int    `yaml:"http_timeout_sec"`
}

type JobConfig struct {
	NotifyIntervalSec        int `yaml:"notify_interval_sec"`
	TimeoutRefundIntervalSec int `yaml:"timeout_refund_interval_sec"`
}

type RateLimitConfig struct {
	IndexQPS float64 `yaml:"index_qps"` // 首页接口按 userId QPS，0 关闭
	Burst    int     `yaml:"burst"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8091
	}
	if cfg.Notify.HTTPTimeoutSec == 0 {
		cfg.Notify.HTTPTimeoutSec = 5
	}
	if cfg.Job.NotifyIntervalSec == 0 {
		cfg.Job.NotifyIntervalSec = 30
	}
	if cfg.Job.TimeoutRefundIntervalSec == 0 {
		cfg.Job.TimeoutRefundIntervalSec = 60
	}
	if cfg.RabbitMQ.URL == "" {
		cfg.RabbitMQ.URL = "amqp://admin:admin@127.0.0.1:5672/"
	}
	if cfg.RabbitMQ.Exchange == "" {
		cfg.RabbitMQ.Exchange = "group_buy_market_exchange"
	}
	if cfg.RateLimit.IndexQPS == 0 {
		cfg.RateLimit.IndexQPS = 1 // 对齐 Java permitsPerSecond=1
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 1
	}
	return &cfg, nil
}
