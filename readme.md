# Carecall-Helloworld (NEST + OPENAI)

## Introduction
PortAudio를 통해 마이크로 입력을 받아들이고, NEST를 통해 STT를 수행하고, OPENAI를 통해 대화를 수행하는 프로젝트입니다.
이 프로젝트를 통해 대화 종료시 function call이 수행되는 과정을 이해 할 수 있습니다.

### Installation - ubuntu
```bash
sudo apt-get install portaudio19-dev
```

### Installation - mac
```bash
brew install portaudio
```

### golang project 생성
```bash
go mod init carecall-helloworld
go get github.com/gordonklaus/portaudio
```

### 실행환경 설정
- OPENAI_API_KEY: https://platform.openai.com/docs/guides/authentication 에서 API key를 발급
- CLOVA_SPEECH_API_KEY: https://guide.ncloud-docs.com/docs/clovaspeech-builder-short 를 참고하여 builder에서 key를 발급

```bash
export OPENAI_API_KEY=YOUR_OPENAI_API_KEY
export CLOVA_SPEECH_API_KEY=YOUR_CLOVA_SPEECH_API_KEY
```

## 흐름 설명 
- 오디오 입력을 받아들이고, STT를 수행하고, 대화를 수행하는 메인 함수입니다.
- 오디오 입력은 PortAudio를 통해 받아들이고, STT는 NEST를 통해 수행합니다.
- 대화는 OPENAI를 통해 수행하고 대화 결과에 따라 function call을 수행합니다.


```go
func processTurn(messages *[]openai.Message) error {
	// 오디오 녹음
	wavData, err := recorder.RecordAudio()                                    // 1. 오디오 녹음
	if err != nil {
		return fmt.Errorf("error recording audio: %v", err)
	}

	client := nestclient.NewSTTClient()
	client.Params["assessment"] = "false" // 발음평가
	client.Params["graph"] = "false"      // 그래프

	userInput, err := client.Recognize(wavData)                              // 2. STT 수행
	if err != nil {
		return fmt.Errorf("error getting user input: %v", err)
	}

	userMessage := openai.Message{Role: "user", Content: userInput.Text}
	*messages = append(*messages, userMessage)
	printMessages("환자", userMessage)

	fmt.Printf("\n\n====%d====================\n", len(*messages))
	response, err := openai.Request(*messages)                               // 3. 대화 수행
	if err != nil {
		return fmt.Errorf("error sending request to OpenAI: %v", err)
	}

	message, err := processResponse(response)                               // 4. 대화 결과 처리
	if errors.Is(err, openai.ErrEndConversation) {
		return openai.ErrEndConversation                                    // 5. 대화 종료
	}

	printMessages("의사", message)                                           // 6. 응답 출력
	*messages = append(*messages, message)

	return nil
}
``` 

## NEST 연동
- NEST를 통해 STT를 수행합니다. Recognize 함수를 통해 STT를 수행하고 결과를 반환합니다.
- octet-stream으로 header를 설정하고, request body에 wav 데이터를 넣어 요청합니다.

```go
req.Header.Set("Content-Type", "application/octet-stream")
req.Header.Set("X-CLOVASPEECH-API-KEY", c.APIKey)

resp, err := http.DefaultClient.Do(req)
````
- 응답은 json 형태로 반환되며, text에 STT 결과가 있습니다.

```go
type STTResponse struct {
  Text  string `json:"text"`
  Quota int    `json:"quota"`
}

```
## OPENAI 연동 
- function call을 수행하기 위해서는 Setting에 FunctionCallSetting을 추가하고, RequestBody에 FunctionSpec을 추가해야 합니다.
```go
requestBodyBytes, err := json.Marshal(RequestBody{
    Model:               config.OpenaiModel,
    Messages:            messages,
    Functions:           functionSpecs,
    FunctionCallSetting: "auto", // function call을 수행하도록 설정
})
```

- functions에 functionSpec을 추가합니다. 이는 function의 이름과 설명 parameter들의 명세입니다.
```go
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

```

- processResponse 함수를 통해 대화 결과를 처리합니다.
- OPENAI는 결과에 Message와 FunctionCall을 둘 중 하나를 반환합니다.function call이 있는 경우, message는 없습니다.

```go
func processResponse(response *openai.ResponseBody) (openai.Message, error) {
	if len(response.Choices) < 1 {
		log.Println("No response from the API.")
		return openai.Message{}, fmt.Errorf("no response from the API")
	}

	responseContent := response.Choices[0].Message.Content
	functionCall := response.Choices[0].Message.FunctionCall

	if functionCall.Name != "" {
		openai.HandleFunctionCall(functionCall)
		return openai.Message{Role: "assistant", Content: "function call"}, openai.ErrEndConversation
	}

	return openai.Message{Role: "assistant", Content: responseContent}, nil
}

```

- name과 arguments를 가지는 FunctionCallResponse 구조체를 정의합니다.

```go
type FunctionCallResponse struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type RequestBody struct {
	Model               string         `json:"model"`
	Messages            []Message      `json:"messages"`
	Functions           []FunctionSpec `json:"functions,omitempty"`
	FunctionCallSetting string         `json:"function_call,omitempty"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Content      string               `json:"content,omitempty"`
			FunctionCall FunctionCallResponse `json:"function_call,omitempty"`
		} `json:"message"`
	} `json:"choices"`
}

```

- 반환된 결과중 functionCall이 있는 경우, HandleFunctionCall 함수를 통해 function call을 수행합니다.

```go
func HandleFunctionCall(functionCall FunctionCallResponse) error {
	switch functionCall.Name {
	case "medical_interview":
		return handleFunct****ion("medical_interview", functionCall.Arguments, "medical_interview.json")
	case "end_conversation":
		return handleFunction("end_conversation", functionCall.Arguments, "end_conversation.json")
	case "call_emergency":
		log.Printf("Function call: %s, %s", functionCall.Name, functionCall.Arguments)
	default:
		log.Printf("Unknown function call: %s", functionCall.Name)
	}
	return nil
}
```

## 요약
- NEST를 통해 STT를 수행하고, 결과중에 text를 추출합니다.
- OPENAI로 Request 할때 function call specification을 전달합니다.
- 응답 결과에 function call이 있는 경우, HandleFunctionCall 함수를 통해 function call을 수행합니다.
- function call 중에 end_conversation이 있는 경우 대화를 종료합니다.
