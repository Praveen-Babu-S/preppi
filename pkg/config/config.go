package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string
	GRPCPort    int
	HTTPPort    int

	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}

	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}

	NATS struct {
		URL string
	}

	Storage struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
	}

	JWT struct {
		PublicKey  string
		PrivateKey string
		AccessTTL  int // minutes
		RefreshTTL int // hours
	}

	LogLevel string
}

func Load(serviceName string) *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../")

	viper.SetEnvPrefix(strings.ToUpper(serviceName))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()

	cfg := &Config{}
	cfg.ServiceName = serviceName

	cfg.GRPCPort = viper.GetInt("grpc.port")
	cfg.HTTPPort = viper.GetInt("http.port")
	cfg.LogLevel = viper.GetString("log.level")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	cfg.Database.Host = viper.GetString("database.host")
	cfg.Database.Port = viper.GetInt("database.port")
	cfg.Database.User = viper.GetString("database.user")
	cfg.Database.Password = viper.GetString("database.password")
	cfg.Database.Name = viper.GetString("database.name")
	cfg.Database.SSLMode = viper.GetString("database.sslmode")

	cfg.Redis.Host = viper.GetString("redis.host")
	cfg.Redis.Port = viper.GetInt("redis.port")
	cfg.Redis.Password = viper.GetString("redis.password")
	cfg.Redis.DB = viper.GetInt("redis.db")

	cfg.NATS.URL = viper.GetString("nats.url")

	cfg.Storage.Endpoint = viper.GetString("storage.endpoint")
	cfg.Storage.AccessKey = viper.GetString("storage.access_key")
	cfg.Storage.SecretKey = viper.GetString("storage.secret_key")
	cfg.Storage.Bucket = viper.GetString("storage.bucket")
	cfg.Storage.UseSSL = viper.GetBool("storage.use_ssl")

	cfg.JWT.PublicKey = viper.GetString("jwt.public_key")
	cfg.JWT.PrivateKey = viper.GetString("jwt.private_key")
	cfg.JWT.AccessTTL = viper.GetInt("jwt.access_ttl")
	cfg.JWT.RefreshTTL = viper.GetInt("jwt.refresh_ttl")

	return cfg
}

func (c *Config) DatabaseDSN() string {
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name, c.Database.SSLMode,
	)
}

func (c *Config) RedisAddr() string {
	if c.Redis.Port == 0 {
		c.Redis.Port = 6379
	}
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
