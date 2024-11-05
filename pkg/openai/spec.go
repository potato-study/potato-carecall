package openai

import (
	"log"
	"os"
)

var functionSpecs = GetFunctionSpec()

func PromptMessage() Message {
	content, err := os.ReadFile("prompt")
	if err != nil {
		log.Fatalf("Failed to read prompt file: %v", err)
	}
	return Message{
		Role:    "system",
		Content: string(content),
	}
}

func GetFunctionSpec() []FunctionSpec {
	return []FunctionSpec{
		getCallEmergencySpec(),
		getMedicalInterviewSpec(),
		getEndConversation(),
	}
}

func getCallEmergencySpec() FunctionSpec {
	return FunctionSpec{
		Name:        "call_emergency",
		Description: "위급상황",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"emergency_type": map[string]interface{}{
					"type":        "string",
					"description": "응급 상황의 유형",
				},
				"call_number_kor": map[string]interface{}{
					"type":        "string",
					"description": "긴급전화번호",
				},
				"location": map[string]interface{}{
					"type":        "string",
					"description": "응급 상황 발생 위치",
				},
			},
		},
	}
}

func getMedicalInterviewSpec() FunctionSpec {
	return FunctionSpec{
		Name:        "medical_interview",
		Description: "medical interview",
		Parameters:  getCommonParameters(),
	}
}

func getEndConversation() FunctionSpec {
	commonParams := getCommonParameters()
	commonParams["properties"].(map[string]interface{})["reason"] = map[string]interface{}{
		"type":        "string",
		"description": "대화를 종료하려는 이유",
	}
	commonParams["required"] = append(commonParams["required"].([]string), "reason")

	return FunctionSpec{
		Name:        "end_conversation",
		Description: "End the conversation gracefully and generate the final patient chart.",
		Parameters:  commonParams,
	}
}

func getCommonParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"patient_info": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "환자의 이름",
					},
					"age": map[string]interface{}{
						"type":        "integer",
						"description": "환자의 나이",
					},
					"gender": map[string]interface{}{
						"type":        "string",
						"description": "환자의 성별",
					},
				},
				"required": []string{"name", "age", "gender"},
			},
			"main_symptoms": map[string]interface{}{
				"type":        "string",
				"description": "환자가 호소하는 주요 증상",
			},
			"medical_history": map[string]interface{}{
				"type":        "string",
				"description": "환자의 과거 병력",
			},
			"additional_symptoms": map[string]interface{}{
				"type":        "string",
				"description": "환자의 추가 증상",
			},
			"lifestyle": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"diet": map[string]interface{}{
						"type":        "string",
						"description": "식습관",
					},
					"smoking": map[string]interface{}{
						"type":        "boolean",
						"description": "흡연 여부",
					},
					"alcohol": map[string]interface{}{
						"type":        "boolean",
						"description": "음주 여부",
					},
					"exercise": map[string]interface{}{
						"type":        "string",
						"description": "운동 빈도",
					},
				},
			},
			"initial_diagnosis": map[string]interface{}{
				"type":        "string",
				"description": "초기 진단",
			},
			"recommended_tests": map[string]interface{}{
				"type":        "string",
				"description": "추가로 추천되는 검사",
			},
			"advice": map[string]interface{}{
				"type":        "string",
				"description": "환자를 위한 추가 조언",
			},
		},
		"required": []string{
			"patient_info",
			"main_symptoms",
			"medical_history",
			"additional_symptoms",
			"lifestyle",
			"initial_diagnosis",
			"recommended_tests",
			"advice",
		},
	}
}
