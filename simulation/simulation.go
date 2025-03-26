package simulation

import (
	"fmt"
	"io"
	"log"
	"time"

	"littleAPITestingTool/fileio"
	"littleAPITestingTool/requests"

	"gopkg.in/yaml.v3"
)

type simMessage struct {
	MessageFormat string `yaml:"MessageFormat"`
	MessageType   string `yaml:"MessageType"`
	Message       string `yaml:"Message"`
	Frequency     string `yaml:"frequency"`
}

func RunSimulation(path string, config fileio.Config) error {
	simMessages, err := loadSimulationConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load simulation config: %v", err)
	}

	for _, msg := range simMessages {
		duration, err := time.ParseDuration(msg.Frequency)
		if err != nil {
			log.Printf("Invalid frequency for message %s: %v", msg.MessageType, err)
			continue
		}
		// add logic for variable substitutions, dependencies, etc.
		go func(m simMessage) {
			ticker := time.NewTicker(duration)
			defer ticker.Stop()
			for {
				<-ticker.C

				if err := requests.SendRequest(config, requests.Message{
					MessageFormat: m.MessageFormat,
					MessageType:   m.MessageType,
					Message:       m.Message,
				}); err != nil {
					log.Printf("Error sending simulated message: %v", err)
				} else {
					log.Printf("Simulated message %s sent successfully", m.MessageType)
				}
			}
		}(msg)
	}

	return nil
}

func loadSimulationConfig(path string) ([]simMessage, error) {
	reader, err := fileio.GetFileIORead(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	defer reader.Close()

	var simMessages []simMessage
	decoder := yaml.NewDecoder(reader)
	for {
		if err := decoder.Decode(&simMessages); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error decoding YAML: %v", err)
		}
	}
	return simMessages, nil
}
