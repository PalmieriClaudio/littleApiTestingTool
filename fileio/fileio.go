package fileio

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	APIEndpoint string `json:"api_endpoint"`
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

func ParseConfig(path string) (string, string, string, error) {
	var config Config

	data, err := ReadFile(path)
	if err != nil {
		return "", "", "", err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open file %s: %v", path, err)
	}
	return config.APIEndpoint, config.Username, config.Password, nil
}

func GetFileIORead(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	return file, nil
}
