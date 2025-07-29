package main

import (
	"fmt"
	"io"
	"log"

	"littleAPITestingTool/fileio"
	"littleAPITestingTool/requests"
	"littleAPITestingTool/simulation"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

func main() {
	const configPath string = "config.json"
	const contentPath string = "data.yaml"
	const simConfigPath string = "sim.yaml"

	simulationRunning := false

	prompt := promptui.Select{
		Label: "Select an option:",
		Items: []string{"Reload endpoint configurations", "Test messages", "Start simulation", "Exit"},
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
				if err := requests.SendRequest(*config, msg); err != nil {
					log.Printf("Error sending message: %v", err)
				} else {
					fmt.Println("Message sent successfully!")
				}
			}
		case 2:
			if !simulationRunning {
				go func() {
					simulation.RunSimulation(simConfigPath, *config)
				}()
				simulationRunning = true
				fmt.Println("Simulation started.")
			} else {
				fmt.Println("Simulation already running.")
			}
		case 3:
			break mainLoop
		default:
			fmt.Println("Invalid selection.")
		}
	}
}

func readMessages(path string) ([]requests.Message, error) {
	reader, err := fileio.GetFileIORead(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer reader.Close()

	var messages []requests.Message
	decoder := yaml.NewDecoder(reader)
	for {
		var msg requests.Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error decoding YAML: %v", err)
		}
		messages = append(messages, msg)
		fmt.Println(msg)
	}
	return messages, nil
}
