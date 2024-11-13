package openai

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type PatientInfo struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

type Lifestyle struct {
	Diet     string `json:"diet"`
	Smoking  bool   `json:"smoking"`
	Alcohol  bool   `json:"alcohol"`
	Exercise string `json:"exercise"`
}

type EndConversationParams struct {
	PatientInfo        PatientInfo `json:"patient_info"`
	MainSymptoms       string      `json:"main_symptoms"`
	MedicalHistory     string      `json:"medical_history"`
	AdditionalSymptoms string      `json:"additional_symptoms"`
	Lifestyle          Lifestyle   `json:"lifestyle"`
	InitialDiagnosis   string      `json:"initial_diagnosis"`
	RecommendedTests   string      `json:"recommended_tests"`
	Advice             string      `json:"advice"`
	Reason             string      `json:"reason"`
}

func HandleFunctionCall(functionCall FunctionCallResponse) error {
	switch functionCall.Name {
	case "end_conversation":
		var params EndConversationParams
		if err := json.Unmarshal([]byte(functionCall.Arguments), &params); err != nil {
			log.Printf("Failed to unmarshal arguments: %v", err)
			return err
		}
		// Use the params object as needed
		fmt.Printf("Function call: %s, \nParams: %+v\n", functionCall.Name, params)
		return handleFunction("end_conversation", functionCall.Arguments, "end_conversation.json")
	default:
		argument, err := makePretty(functionCall.Arguments)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("Unknown function call: %s, %s", functionCall.Name, argument)
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
