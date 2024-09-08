package configs

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
)

func Initialize() error {
	_ = os.Setenv("ENV", "env")
	env := os.Getenv("ENV")
	fileName := env

	viper.SetConfigName(fileName)
	viper.SetConfigType("json")
	viper.AddConfigPath("./config/")
	viper.AutomaticEnv() // if a key is present in the json file and environment, the environment value will be used.

	fileName += ".json"

	slog.Info("Reading configs from ./configs/" + fileName)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	slog.Info("Configs read successfully from ./configs/" + fileName)

	return nil
}