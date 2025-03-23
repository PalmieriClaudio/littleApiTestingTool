package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"littleAPITestingTool/fileio"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

type Message struct {
	MessageFormat string `yaml:"MessageFormat"`
	MessageType   string `yaml:"MessageType"`
	Message       string `yaml:"Message"`
}

func main() {
	const configPath string = "config.json"
	const contentPath string = "data.yaml"

	prompt := promptui.Select{
		Label: "Select an option:",
		Items: []string{"Reload endpoint configurations", "Test messages", "Exit"},
	}
	config, err := fileio.ParseConfig(configPath)
	if err != nil {
		log.Fatalf("Error unmarshalling config.json: %v", err)
	}
	fmt.Printf("API Endpoint: %s\nAuthenticationType: %s\nUsername: %s\nPassword: %s\n", config.APIEndpoint, config.AuthType, config.Username, config.Password)

mainLoop:
	for {
		switch index, _, _ := prompt.Run(); index {
		case 0:
			config, err = fileio.ParseConfig(configPath)
			if err != nil {
				log.Fatalf("Error unmarshalling config.json: %v", err)
			}
			fmt.Printf("API Endpoint: %s\nAuthenticationType: %s\nUsername: %s\nPassword: %s\n", config.APIEndpoint, config.AuthType, config.Username, config.Password)
		case 1:
			messages, err := readMessages(contentPath)
			if err != nil {
				log.Fatalf("Error reading messages: %v", err)
			}
			for _, msg := range messages {
				if err := sendRequest(*config, msg); err != nil {
					log.Printf("Error sending message: %v", err)
				} else {
					fmt.Println("Message sent successfully!")
				}
			}
		case 2:
			break mainLoop
		default:
			fmt.Println("Invalid selection.")
		}
	}
}

func readMessages(path string) ([]Message, error) {
	reader, err := fileio.GetFileIORead(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer reader.Close()

	var messages []Message
	decoder := yaml.NewDecoder(reader)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error decoding YAML: %v", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func sendRequest(config fileio.Config, msg Message) error {
	data, err := yaml.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	var contentType string
	var authHeaderValue string
	switch strings.ToLower(msg.MessageFormat) {
	case "xml":
		contentType = "application/xml"
	case "json":
		contentType = "application/json"
	default:
		return fmt.Errorf("MessageFormat %s not managed", msg.MessageFormat)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", config.APIEndpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating POST request: %v", err)
	}

	switch strings.ToLower(config.AuthType) {
	case "basic":
		authHeaderValue = "basic " + base64.StdEncoding.EncodeToString([]byte(config.Username+":"+config.Password))
		req.Header.Add("Authorization", authHeaderValue)
	case "none", "", "anonymous":
	default:
		return fmt.Errorf("unhandled auth type")
	}

	req.Header.Add("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP POST request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response: %s\n", resp.Status)
	return nil
}
