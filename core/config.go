package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config structure for the database connection and server settings
type Config struct {
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		DbName   string `json:"dbname"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"database"`
	Server struct {
		Port int `json:"port"`
	} `json:"server"`
}

// loadConfig loads the configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %v", err)
	}
	return &config, nil
}
