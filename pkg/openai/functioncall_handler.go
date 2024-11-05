package openai

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func HandleFunctionCall(functionCall FunctionCallResponse) error {
	switch functionCall.Name {
	case "medical_interview":
		return handleFunction("medical_interview", functionCall.Arguments, "medical_interview.json")
	case "end_conversation":
		return handleFunction("end_conversation", functionCall.Arguments, "end_conversation.json")
	case "call_emergency":
		log.Printf("Function call: %s, %s", functionCall.Name, functionCall.Arguments)
	default:
		log.Printf("Unknown function call: %s", functionCall.Name)
	}
	return nil
}

func handleFunction(name, arguments, filename string) error {
	prettyJSON, err := makePretty(arguments)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("Function call: %s , %s", name, prettyJSON)
	log.Println(name)
	if filename != "" {
		return saveJSONToFile(prettyJSON, filename)
	}
	return ErrEndConversation
}

func saveJSONToFile(data, filename string) error {
	return os.WriteFile(filename, []byte(data), 0644)
}

func makePretty(data string) (string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return string(prettyJSON), nil
}
