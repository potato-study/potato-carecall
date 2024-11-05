package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"potato-carecall/pkg/config"
)

func Request(messages []Message) (*ResponseBody, error) {
	requestBodyBytes, err := json.Marshal(RequestBody{
		Model:               config.OpenaiModel,
		Messages:            messages,
		Functions:           functionSpecs,
		FunctionCallSetting: "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("POST", config.OpenaiApiUrl, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.OpenaiApiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s %s", resp.Status, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var responseBody ResponseBody
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return &responseBody, nil
}
