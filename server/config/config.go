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
	GrpcConfig       *GrpcConfig       `mapstructure:"grpc"`
	LogConfig        *LogConfig        `mapstructure:"log"`
	HttpConfig       *HttpConfig       `mapstructure:"http"`
	RedisConfig      *RedisConfig      `mapstructure:"redis"`
	SystemConfig     *SystemConfig     `mapstructure:"system"`
	OpenAIConfig     *OpenAIConfig     `mapstructure:"openai"`
	OSSConfig        *OSSConfig        `mapstructure:"oss"`
	KafkaConfig      *KafkaConfig      `mapstructure:"kafka"`
	UsergroupGrpc    *GrpcClientConfig `mapstructure:"usergroup_grpc"`
	TaskGrpc         *GrpcClientConfig `mapstructure:"task_grpc"`
	RewardGrpc       *GrpcClientConfig `mapstructure:"reward_grpc"`
}

type SystemConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// OpenAIConfig drives ChatGPT translation calls (optional in dev).
// APIKey comes from env OPENAI_API_KEY only; base_url/model may stay in config.yml.
type OpenAIConfig struct {
	APIKey  string `mapstructure:"-"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// OSSConfig is Aliyun OSS settings. Access keys come from env only.
type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	PublicBaseURL   string `mapstructure:"public_base_url"`
	KeyPrefix       string `mapstructure:"key_prefix"`
	AccessKeyID     string `mapstructure:"-"`
	AccessKeySecret string `mapstructure:"-"`
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

// GrpcClientConfig is an outbound gRPC dependency endpoint.
type GrpcClientConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type RedisConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
}

type KafkaConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Brokers  []string `mapstructure:"brokers"`
	GroupID  string   `mapstructure:"group_id"`
	ClientID string   `mapstructure:"client_id"`
	Topics   []string `mapstructure:"topics"`
}

func Init() {
	Config = &Conf{}
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	if workDir, err := os.Getwd(); err == nil {
		// Support both `server/` and repo-root as process cwd (e.g. VS Code launch).
		viper.AddConfigPath(workDir + "/resources")
		viper.AddConfigPath(workDir + "/server/resources")
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
	viper.SetDefault("oss.endpoint", "oss-ap-southeast-1.aliyuncs.com")
	viper.SetDefault("oss.bucket", "campaign-center-img")
	viper.SetDefault("oss.key_prefix", "campaign-center")
	viper.SetDefault("kafka.enabled", false)
	viper.SetDefault("kafka.group_id", "campaign-center-api")
	viper.SetDefault("kafka.client_id", "campaign-center-api")
	viper.SetDefault("usergroup_grpc.host", "127.0.0.1")
	viper.SetDefault("usergroup_grpc.port", 50051)
	viper.SetDefault("task_grpc.host", "127.0.0.1")
	viper.SetDefault("task_grpc.port", 50052)
	viper.SetDefault("reward_grpc.host", "127.0.0.1")
	viper.SetDefault("reward_grpc.port", 50053)

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
	applyOSSCredentialsFromEnv()
}

func applyConfigDefaults() {
	if Config.SystemConfig == nil {
		Config.SystemConfig = &SystemConfig{}
	}
	if Config.OpenAIConfig == nil {
		Config.OpenAIConfig = &OpenAIConfig{}
	}
	if Config.OSSConfig == nil {
		Config.OSSConfig = &OSSConfig{}
	}
	if Config.KafkaConfig == nil {
		Config.KafkaConfig = &KafkaConfig{}
	}
	if Config.UsergroupGrpc == nil {
		Config.UsergroupGrpc = &GrpcClientConfig{}
	}
	if Config.TaskGrpc == nil {
		Config.TaskGrpc = &GrpcClientConfig{}
	}
	if Config.RewardGrpc == nil {
		Config.RewardGrpc = &GrpcClientConfig{}
	}
	if strings.TrimSpace(Config.OSSConfig.Endpoint) == "" {
		Config.OSSConfig.Endpoint = "oss-ap-southeast-1.aliyuncs.com"
	}
	if strings.TrimSpace(Config.OSSConfig.Bucket) == "" {
		Config.OSSConfig.Bucket = "campaign-center-img"
	}
	if strings.TrimSpace(Config.OSSConfig.KeyPrefix) == "" {
		Config.OSSConfig.KeyPrefix = "campaign-center"
	}
	if strings.TrimSpace(Config.KafkaConfig.GroupID) == "" {
		Config.KafkaConfig.GroupID = "campaign-center-api"
	}
	if strings.TrimSpace(Config.KafkaConfig.ClientID) == "" {
		Config.KafkaConfig.ClientID = "campaign-center-api"
	}
	if len(Config.KafkaConfig.Topics) == 0 {
		Config.KafkaConfig.Topics = []string{"task.events.completed", "reward.distribution.result"}
	}
}

// applyOpenAIAPIKeyFromEnv loads the API key from OPENAI_API_KEY only.
func applyOpenAIAPIKeyFromEnv() {
	if Config.OpenAIConfig == nil {
		Config.OpenAIConfig = &OpenAIConfig{}
	}
	if v := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); v != "" {
		Config.OpenAIConfig.APIKey = v
	}
}

// applyOSSCredentialsFromEnv loads AK/SK from environment variables only.
func applyOSSCredentialsFromEnv() {
	if Config.OSSConfig == nil {
		Config.OSSConfig = &OSSConfig{}
	}
	if v := strings.TrimSpace(os.Getenv("OSS_ACCESS_KEY_ID")); v != "" {
		Config.OSSConfig.AccessKeyID = v
	}
	if v := strings.TrimSpace(os.Getenv("OSS_ACCESS_KEY_SECRET")); v != "" {
		Config.OSSConfig.AccessKeySecret = v
	}
}
