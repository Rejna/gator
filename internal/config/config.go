package config

import (
	"encoding/json"
	"os"
	"errors"
	"fmt"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl string						`json:"db_url"`
	CurrentUserName string	`json:"current_user_name"`
}

func Read() Config {
	path, err := getConfigFilePath()
	if err != nil {
		fmt.Println(err)
		return Config{}
	}

	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return Config{}
	}
	defer jsonFile.Close()

	var config Config
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println(err)
		return Config{}
	}

	return config
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	if err := write(*c); err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homedir, _ := os.UserHomeDir()
	path := fmt.Sprintf("%s/%s", homedir, configFileName)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("config file doesn't exist")
	}
	return path, nil
}

func write(cfg Config) error {
	jsonString, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	os.WriteFile(path, jsonString, os.ModePerm)
	return nil
}