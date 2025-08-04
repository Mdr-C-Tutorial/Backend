package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	TTL           int64  `mapstructure:"ttl"`
	RememberMeTTL int64  `mapstructure:"remember_me_ttl"`
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"from_email"`
}

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
}

var AppConfig Config

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	
	// 打印 viper 读取到的所有配置
	fmt.Printf("All viper configs: %+v\n", viper.AllSettings())

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
