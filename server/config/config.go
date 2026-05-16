package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

var (
	Config = &Conf{}
)

type Conf struct {
	GrpcConfig   *GrpcConfig   `mapstructure:"grpc"`
	LogConfig    *LogConfig    `mapstructure:"log"`
	HttpConfig   *HttpConfig   `mapstructure:"http"`
	RedisConfig  *RedisConfig  `mapstructure:"redis"`
	SystemConfig *SystemConfig `mapstructure:"system"`
	OpenAIConfig *OpenAIConfig `mapstructure:"openai"`
}

type SystemConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// OpenAIConfig drives ChatGPT translation calls (optional in dev).
// APIKey is loaded from config then overridden by OPENAI_API_KEY if set.
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

type HttpConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
}

type GrpcConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ConnectTimeout int    `mapstructure:"connect_timeout"`
	MaxPoolSize    int    `mapstructure:"max_pool_size"`
}

type RedisConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
}

func Init() {
	Config = &Conf{}
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	if workDir, err := os.Getwd(); err == nil {
		viper.AddConfigPath(workDir + "/resources")
		viper.AddConfigPath(workDir)
	}
	viper.SetEnvPrefix("CAMPAIGN_CENTER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 9090)
	viper.SetDefault("grpc.connect_timeout", 5)
	viper.SetDefault("grpc.max_pool_size", 100)
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("openai.model", "gpt-4o-mini")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
	if err := viper.Unmarshal(Config); err != nil {
		panic(err)
	}
	applyConfigDefaults()
	applyOpenAIAPIKeyFromEnv()
}

func applyConfigDefaults() {
	if Config.SystemConfig == nil {
		Config.SystemConfig = &SystemConfig{}
	}
	if Config.OpenAIConfig == nil {
		Config.OpenAIConfig = &OpenAIConfig{}
	}
}

// applyOpenAIAPIKeyFromEnv sets OpenAI API key from OPENAI_API_KEY when non-empty
// (overrides config file / CAMPAIGN_CENTER_OPENAI_API_KEY from viper).
func applyOpenAIAPIKeyFromEnv() {
	if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		Config.OpenAIConfig.APIKey = v
	}
}
