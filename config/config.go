package config

import (
	"automelonloaderinstallergo/internal/logger"
	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("config") // config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Log.Info("No config file found; using defaults.")
	}

	viper.SetDefault("app.theme", "dark")
}
