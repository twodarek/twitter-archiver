package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ConsumerKey    string `mapstructure:"API_KEY"`
	ConsumerSecret string `mapstructure:"API_KEY_SECRET"`
	AccessToken    string `mapstructure:"ACCESS_TOKEN"`
	AccessSecret   string `mapstructure:"ACCESS_TOKEN_SECRET"`
	APIBearerToken string `mapstructure:"BEARER_TOKEN"`

	DatabaseUser string `mapstructure:"DB_USER"`
	DatabasePass string `mapstructure:"DB_PASS"`
	DatabaseHost string `mapstructure:"DB_HOST"`
	DatabasePort string `mapstructure:"DB_PORT"`
	DatabaseName string `mapstructure:"DB_NAME"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
