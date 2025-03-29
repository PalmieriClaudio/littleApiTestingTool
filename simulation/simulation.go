package simulation

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
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
	var wg sync.WaitGroup
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

		ticker := time.NewTicker(duration)
		wg.Add(1)
		defer ticker.Stop()
		go func(wg *sync.WaitGroup, m simMessage) {
			tmpl, err := template.New("parsedMessage").Delims("{{", "}}").Parse(msg.Message)
			if err != nil {
				fmt.Printf("error: %v", err)
				return
			}
			parsedVars, err := varsParser(msg.Variables)
			if err != nil {
				log.Printf("error in variable parsing: %v", err)
			}
			composeAndSendRequest(parsedVars, tmpl, config, m)
			for range ticker.C {
				composeAndSendRequest(parsedVars, tmpl, config, m)
			}
			wg.Done()
		}(&wg, msg)
	}
	wg.Wait()
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
			pVars[key] = []any{value.Value, false}
		case "sequence":
			temp, err := parseSequence(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = []any{temp, true}
		case "range":
			temp, err := parseRange(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = []any{temp, true}
		case "random":
			temp, err := parseRandInRange(value.Value)
			if err != nil {
				return nil, err
			}
			pVars[key] = []any{temp, false}
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

func composeAndSendRequest(parsedVars map[string]any, tmpl *template.Template, config fileio.Config, m simMessage) {
	for _, v := range parsedVars {
		t := reflect.TypeOf(v.([]any)[0]).Kind()
		if (t == reflect.Slice) && len(v.([]any)[0].([]string)) == 0 {
			return
		}
	}
	var buf bytes.Buffer
	// create a new struct that only contains the first values for slices
	tmplVars := make(map[string]any, len(parsedVars))
	for k, v := range parsedVars {
		w := v.([]any)[0]
		t := reflect.TypeOf(w).Kind()
		if t == reflect.Slice || t == reflect.Array {
			tmplVars[k] = w.([]string)[0]
		} else {
			tmplVars[k] = w
		}
	}

	if err := tmpl.Execute(&buf, tmplVars); err != nil {
		log.Printf("error in template execution: %v", err)
		return
	}
	content := buf.String()
	fmt.Printf("%v\n", content)
	if err := requests.SendRequest(config, requests.Message{
		MessageFormat: m.MessageFormat,
		MessageType:   m.MessageType,
		Message:       buf.String(),
	}); err != nil {
		log.Printf("Error sending %v message: %v\n", m.MessageType, err)
	} else {
		log.Printf("Message %s sent successfully", m.MessageType)
	}
	for k, v := range parsedVars {
		if v.([]any)[1] == true {
			parsedVars[k].([]any)[0] = parsedVars[k].([]any)[0].([]string)[1:]
		}
	}
}
