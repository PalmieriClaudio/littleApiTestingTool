package requests

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"littleAPITestingTool/fileio"

	"gopkg.in/yaml.v3"
)

type Message struct {
	MessageFormat string
	MessageType   string
	Message       string
}

func SendRequest(config fileio.Config, msg Message) error {
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
	apiEndpoint := config.APIEndpoint
	if config.DynamicPath {
		apiEndpoint = apiEndpoint + "/" + msg.MessageType
	}

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating POST request: %v", err)
	}

	switch strings.ToLower(config.AuthType) {
	case "basic":
		authHeaderValue = "basic " + base64.StdEncoding.EncodeToString([]byte(config.Username+":"+config.Password))
		req.Header.Add("Authorization", authHeaderValue)
	case "Oauth":
		fmt.Println("Oauth not yet implemented.")
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
