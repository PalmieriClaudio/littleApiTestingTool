package fileio

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	APIEndpoint string `json:"APIEndpoint"`
	DynamicPath bool   `json:"dynamicPath"`
	AuthType    string `json:"AuthType"`
	Username    string `json:"Username"`
	Password    string `json:"Password"`
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return data, nil
}

func ParseConfig(path string) (config *Config, err error) {
	// var config Config

	data, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	return config, nil
}

func GetFileIORead(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	return file, nil
}
