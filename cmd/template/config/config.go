package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var ServerCfg = new(ServerConfig)

type ServerConfig struct {
	ServerName string `yaml:"server_name"`
	Port       int    `yaml:"port"`
	DB         struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
		MaxIdle  int    `yaml:"max_idle,omitempty"` //设置连接池中空闲连接的最大数量
		MaxOpen  int    `yaml:"max_open,omitempty"` //设置打开数据库连接的最大数量
	} `yaml:"db"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}

// Init init global config
func Init(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Open config file error: ", err.Error())
		os.Exit(1)
	}
	if err = yaml.Unmarshal(file, ServerCfg); err != nil {
		fmt.Println("Read config file error: ", err.Error())
		os.Exit(1)
	}
}
