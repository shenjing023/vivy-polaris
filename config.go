package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

var Cfg = new(Config)

type Config struct {
	Protobuf struct {
		SourceDir    string   `yaml:"source_dir"`
		DstDir       string   `yaml:"dst_dir"`
		CompileFiles []string `yaml:"compile_files"`
	} `yaml:"protobuf"`
	ModName    string `yaml:"mod_name"`
	ServerName string `yaml:"server_name"`
}

// init global config
func init() {
	file, err := ioutil.ReadFile("gen.yaml")
	if err != nil {
		fmt.Println("Open config file error: ", err.Error())
		os.Exit(1)
	}
	if err = yaml.Unmarshal(file, Cfg); err != nil {
		fmt.Println("Read config file error: ", err.Error())
		os.Exit(1)
	}
}
