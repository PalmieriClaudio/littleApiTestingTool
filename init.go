package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

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

	var apiEndpoint, username, password string
	var err error
	apiEndpoint, _, _, err = fileio.ParseConfig(configPath)
	if err != nil {
		log.Fatalf("Error unmarshalling config.json: %v", err)
	}

	prompt := promptui.Select{
		Label: "Select an option:",
		Items: []string{"Reload endpoint configurations", "Test messages", "Exit"},
	}

mainLoop:
	for {
		switch index, _, _ := prompt.Run(); index {
		case 0:
			apiEndpoint, username, password, err = fileio.ParseConfig(configPath)
			if err != nil {
				log.Fatalf("Error unmarshalling config.json: %v", err)
			}
			fmt.Printf("API Endpoint: %s\nUsername: %s\nPassword: %s\n", apiEndpoint, username, password)
		case 1:
			messages, err := readMessages(contentPath)
			if err != nil {
				log.Fatalf("Error reading messages: %v", err)
			}
			for _, msg := range messages {
				if err := sendRequest(apiEndpoint, msg); err != nil {
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

func sendRequest(apiEndpoint string, msg Message) error {
	data, err := yaml.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	resp, err := http.Post(apiEndpoint, "application/x-yaml", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error making HTTP POST request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response: %s\n", resp.Status)
	return nil
}
