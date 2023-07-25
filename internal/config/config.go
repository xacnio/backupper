package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	DateFormat string        `json:"dateFormat"`
	LogLevel   *string       `json:"logLevel"`
	Backups    []interface{} `json:"backups"`
	Timezone   *string       `json:"timezone"`
}

var config *Config

func Get() *Config {
	return config
}

func ReadConfig() Config {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var configTmp Config
	err = json.Unmarshal(b, &configTmp)
	if err != nil {
		panic(err)
	}

	config = &configTmp

	return configTmp
}
