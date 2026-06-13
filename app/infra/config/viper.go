package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// AppConfig mirrors the settings described in .env.example and exposes
// additional structured sections for downstream consumers.
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Session  SessionConfig  `mapstructure:"session"`
	Node     NodeConfig     `mapstructure:"node"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Runner   RunnerConfig   `mapstructure:"runner"`
	AWS      AWSConfig      `mapstructure:"aws"`
}

// ServerConfig controls how the HTTP server binds to the network.
type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	FrontendDomain string `mapstructure:"frontend_domain"`
}

// DatabaseConfig holds datasource configuration derived from DATABASE_* envs.
type DatabaseConfig struct {
	Type            string        `mapstructure:"type"`
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// SessionConfig stores cookie/session information from SESSION_* envs.
type SessionConfig struct {
	Secret         string `mapstructure:"secret"`
	MaxAgeMS       int    `mapstructure:"max_age"`
	CookieSecure   bool   `mapstructure:"cookie_secure"`
	CookieSameSite string `mapstructure:"cookie_same_site"`
}

type NodeConfig struct {
	Roles   []string `mapstructure:"roles"`
	Name    string   `mapstructure:"name"`
	NodeKey string   `mapstructure:"node_key"`
}

type AgentConfig struct {
	ServerURL         string        `mapstructure:"server_url"`
	Token             string        `mapstructure:"token"`
	DataDir           string        `mapstructure:"data_dir"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	PollInterval      time.Duration `mapstructure:"poll_interval"`
}

type RunnerConfig struct {
	DefaultVersion string `mapstructure:"default_version"`
}

// AWSConfig describes the AWS credentials defined in .env.example.
type AWSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	S3Bucket        string `mapstructure:"s3_bucket"`
}

// NewViper configures a Viper instance that loads defaults, optional .env files,
// config/{config.yaml}, and environment variables that match .env.example.
func NewViper() (*viper.Viper, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	if err := bindEnvVariables(v); err != nil {
		return nil, err
	}

	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("./configs")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if configFile != "" || !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	return v, nil
}

// NewAppConfig unmarshals the loaded configuration into a strongly typed struct.
func NewAppConfig(v *viper.Viper) (*AppConfig, error) {
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
	}
	if cfg.Database.URL == "" {
		return nil, errors.New("database.url is required (see .env.example)")
	}

	if cfg.Session.MaxAgeMS <= 0 {
		cfg.Session.MaxAgeMS = int((14 * 24 * time.Hour) / time.Millisecond)
	}

	return &cfg, nil
}

func loadDotEnv() error {
	if err := gotenv.Load(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.frontend_domain", "")
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.url", "./data.db")
	v.SetDefault("database.max_open_conns", 5)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", time.Minute*30)
	v.SetDefault("session.secret", "change-me")
	v.SetDefault("session.max_age", int((14*time.Hour*24)/time.Millisecond))
	v.SetDefault("session.cookie_secure", false)
	v.SetDefault("session.cookie_same_site", "lax")
	v.SetDefault("node.roles", []string{"server"})
	v.SetDefault("agent.data_dir", "./merc-agent")
	v.SetDefault("agent.heartbeat_interval", 15*time.Second)
	v.SetDefault("agent.poll_interval", 5*time.Second)
	v.SetDefault("runner.default_version", "2.329.0")
	v.SetDefault("aws.region", "us-east-1")
}

func bindEnvVariables(v *viper.Viper) error {
	envBindings := map[string]string{
		"server.host":              "HOST",
		"server.port":              "PORT",
		"server.frontend_domain":   "FRONTEND_DOMAIN",
		"database.type":            "DATABASE_TYPE",
		"database.url":             "DATABASE_URL",
		"session.secret":           "SESSION_SECRET",
		"session.max_age":          "SESSION_MAX_AGE",
		"session.cookie_secure":    "SESSION_COOKIE_SECURE",
		"session.cookie_same_site": "SESSION_COOKIE_SAME_SITE",
		"node.name":                "NODE_NAME",
		"node.node_key":            "NODE_KEY",
		"agent.server_url":         "AGENT_SERVER_URL",
		"agent.token":              "AGENT_TOKEN",
		"agent.data_dir":           "AGENT_DATA_DIR",
		"agent.heartbeat_interval": "AGENT_HEARTBEAT_INTERVAL",
		"agent.poll_interval":      "AGENT_POLL_INTERVAL",
		"runner.default_version":   "RUNNER_DEFAULT_VERSION",
		"aws.access_key_id":        "AWS_ACCESS_KEY_ID",
		"aws.secret_access_key":    "AWS_SECRET_ACCESS_KEY",
		"aws.region":               "AWS_REGION",
		"aws.s3_bucket":            "AWS_S3_BUCKET",
	}

	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			return fmt.Errorf("bind env %s: %w", env, err)
		}
	}

	return nil
}
