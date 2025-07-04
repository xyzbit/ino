package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config 配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Milvus   MilvusConfig   `mapstructure:"milvus"`
	Neo4j    Neo4jConfig    `mapstructure:"neo4j"`
	Eino     EinoConfig     `mapstructure:"eino"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MilvusConfig Milvus向量数据库配置
type MilvusConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Neo4jConfig Neo4j图数据库配置
type Neo4jConfig struct {
	URI      string `mapstructure:"uri"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// EinoConfig Eino AI配置
type EinoConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

var AppConfig Config

// Init 初始化配置
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./configs")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatal("Unable to decode into struct:", err)
	}

	log.Printf("Configuration loaded successfully")
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "ino")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("milvus.host", "localhost")
	viper.SetDefault("milvus.port", 19530)

	viper.SetDefault("neo4j.uri", "bolt://localhost:7687")
	viper.SetDefault("neo4j.username", "neo4j")
	viper.SetDefault("neo4j.password", "password")

	viper.SetDefault("eino.model", "gpt-3.5-turbo")
}
