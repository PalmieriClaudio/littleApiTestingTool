package simulation

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	"littleAPITestingTool/fileio"
	"littleAPITestingTool/requests"

	"gopkg.in/yaml.v3"
)

type simMessage struct {
	MessageFormat string               `yaml:"MessageFormat"`
	MessageType   string               `yaml:"MessageType"`
	Message       string               `yaml:"Message"`
	Variables     map[string]variables `yaml:"Variables"`
	Frequency     string               `yaml:"Frequency"`
}

type variables struct {
	Type  string `yaml:"Type"`
	Value string `yaml:"Value"`
}

// type parsedVars map[string]string

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

		go func(m simMessage) {
			ticker := time.NewTicker(duration)
			defer ticker.Stop()
			tmpl, _ := template.New("parsedMessage").Delims("{{", "}}").Parse(msg.Message)

			parsedVars, err := varsParser(msg.Variables)
			if err != nil {
				log.Printf("error in variable parsing: %v", err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, parsedVars); err != nil {
				log.Printf("error in template execution: %v", err)
				return
			}

			if err := requests.SendRequest(config, requests.Message{
				MessageFormat: m.MessageFormat,
				MessageType:   m.MessageType,
				Message:       buf.String(),
			}); err != nil {
				log.Printf("Error sending %v message: %v", m.MessageType, err)
			} else {
				log.Printf("Message %s sent successfully", m.MessageType)
			}

			<-ticker.C
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

func varsParser(vars map[string]variables) (map[string]any, error) {
	pVars := map[string]any{}
	for key, value := range vars {
		switch value.Type {
		case "static":
			pVars[key] = value.Value
		case "sequence":
			temp, err := parseSequence(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = temp
		case "range":
			temp, err := parseRange(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = temp
		case "random":
			temp, err := parseRandInRange(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = temp
		}
	}
	return pVars, nil
}

func parseSequence(s string) ([]string, error) {
	trimmed := strings.Trim(s, "[]")
	splitted := strings.Split(trimmed, ",")
	for i, str := range splitted {
		splitted[i] = strings.TrimSpace(str)
	}
	return splitted, nil
}

func parseRange(s string) ([]string, error) {
	vals, err := parseSequence(s)
	if err != nil {
		return nil, err
	}

	if len(vals) != 2 {
		return nil, fmt.Errorf("exactly 2 values are expected in a range variable")
	}

	start, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse start value: %v", err)
	}
	end, err := strconv.Atoi(vals[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse end value: %v", err)
	}

	filled := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		filled = append(filled, strconv.Itoa(i))
	}
	return filled, nil
}

func parseRandInRange(s string) (string, error) {
	vals, err := parseSequence(s)
	if err != nil {
		return "", err
	}

	if len(vals) != 2 {
		return "", fmt.Errorf("exactly 2 values are expected in a range variable")
	}

	start, err := strconv.Atoi(vals[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse start value: %v", err)
	}
	end, err := strconv.Atoi(vals[1])
	if err != nil {
		return "", fmt.Errorf("failed to parse end value: %v", err)
	}

	return strconv.Itoa(rand.Intn(end-start+1) + start), nil
}
