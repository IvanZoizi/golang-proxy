package configGolang

import (
	"encoding/json"
	"github.com/bytedance/gopkg/util/logger"
	"os"
)

type Config struct {
	ServerPort string `json:"serverPort"`
	AppName    string `json:"appName"`
	Env        string `json:"env"`
	SecretKey  string `json:"secretKey"`
}

func CreateConfig(filePath string) *Config {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Not found file by path %s, %s", filePath, err)
		return &Config{
			":9000",
			"Go Proxy",
			"development",
			"secret-key",
		}
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		logger.Errorf("Not found file by path %s, %s", filePath, err)
		return &Config{
			":9000",
			"Go Proxy",
			"development",
			"secret-key",
		}
	}

	return &config
}
