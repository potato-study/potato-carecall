# Function Call을 이해해 보자

## 소개
이 프로젝트는 PortAudio를 통해 마이크 입력을 받고, NEST로 STT를 수행하며, OpenAI를 통해 대화를 진행합니다.
이를 통해 대화 종료 시 function call이 어떻게 수행되는지 이해할 수 있습니다.

### 설치 방법 - Ubuntu
```bash
sudo apt-get install portaudio19-dev
```

### 설치 방법 - Mac
```bash
brew install portaudio
```

### 실행 환경 설정
- **OPENAI_API_KEY**: [API 키 발급](https://platform.openai.com/docs/guides/authentication)에서 OpenAI API 키를 발급받습니다.
- **CLOVA_SPEECH_API_KEY**: [CLOVA Speech Builder](https://guide.ncloud-docs.com/docs/clovaspeech-builder-short)를 참고하여 키를 발급받습니다.

```bash
export OPENAI_API_KEY=YOUR_OPENAI_API_KEY
export CLOVA_SPEECH_API_KEY=YOUR_CLOVA_SPEECH_API_KEY
```

## 흐름 설명
- 오디오 입력을 받아 STT를 수행하고, 대화를 진행하는 메인 함수입니다.
- 오디오 입력은 **PortAudio**를 통해 받고, STT는 **NEST**를 통해 수행합니다.
- 대화는 **OpenAI**를 통해 진행하며, 대화 결과에 따라 function call을 수행합니다.

# 실습

## OpenAI와 연동할 주요 함수 작성


### `dialogue.RunDialogue` 함수:
- 대화의 주요 흐름을 관리하는 함수입니다.
- 오디오를 녹음하고 이를 텍스트로 변환한 후, OpenAI API에 요청을 보냅니다.
- 응답을 처리하고 대화 메시지를 업데이트합니다.

`/pkg/dialogue/dialogue.go`
```go
func RunDialogue(messages *[]openai.Message) error {
	// 오디오 녹음
	wavData, err := recorder.RecordAudio()
	if err != nil {
		return fmt.Errorf("error recording audio: %v", err)
	}

	client := nestclient.NewSTTClient()
	client.Params["assessment"] = "false" // 발음 평가
	client.Params["graph"] = "false"      // 그래프

	userInput, err := client.Recognize(wavData)
	if err != nil {
		return fmt.Errorf("error getting user input: %v", err)
	}

	userMessage := openai.Message{Role: "user", Content: userInput.Text}
	*messages = append(*messages, userMessage)
	printMessages("환자", userMessage)

	fmt.Printf("\n\n ===== %d번째 대화 =====\n", len(*messages)/2)
	response, err := openai.Request(*messages)
	if err != nil {
		return fmt.Errorf("error sending request to OpenAI: %v", err)
	}

	message, err := processResponse(response)
	if errors.Is(err, openai.ErrEndConversation) {
		return openai.ErrEndConversation
	}

	printMessages("의사", message)
	*messages = append(*messages, message)

	return nil
}
```

### `dialogue.processResponse` 함수:
- OpenAI API의 응답을 처리하는 함수입니다.
- 응답에서 메시지를 추출하고, 필요한 경우 함수 호출을 처리합니다.

`/pkg/dialogue/dialogue.go`
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

### `openai.Request` 함수:
- OpenAI API에 요청을 보내고 응답을 받는 함수입니다.
- 요청 본문을 JSON으로 직렬화하고, HTTP 요청을 생성하여 API에 보냅니다.
- 응답을 받아 JSON으로 역직렬화합니다.

`/pkg/openai/client.go`
```go
func Request(messages []Message) (*ResponseBody, error) {
	request := RequestBody{
		Model:               config.OpenaiModel,
		Messages:            messages,
		Functions:           loadFunctions(),
		FunctionCallSetting: "auto",
	}

	requestBodyBytes, err := json.Marshal(request)

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
```

### Function Call 스펙 작성:
- `openai.loadFunctions` 함수를 통해 function call을 처리할 수 있는 함수 목록을 로드합니다.
- 환자의 정보를 저장하고, 대화를 종료하는 함수입니다. OpenAI의 응답을 통해 function call을 수행합니다.

`/config/function/end_conversation.go`
```json
{
  "type": "object",
  "properties": {
    "patient_info": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "환자의 이름"
        },
        "age": {
          "type": "integer",
          "description": "환자의 나이"
        },
        "gender": {
          "type": "string",
          "description": "환자의 성별"
        }
      },
      "required": ["name", "age", "gender"]
    },
    "main_symptoms": {
      "type": "string",
      "description": "환자가 호소하는 주요 증상"
    },
    "medical_history": {
      "type": "string",
      "description": "환자의 과거 병력"
    },
    "additional_symptoms": {
      "type": "string",
      "description": "환자의 추가 증상"
    },
    "lifestyle": {
      "type": "object",
      "properties": {
        "diet": {
          "type": "string",
          "description": "식습관"
        },
        "smoking": {
          "type": "boolean",
          "description": "흡연 여부"
        },
        "alcohol": {
          "type": "boolean",
          "description": "음주 여부"
        },
        "exercise": {
          "type": "string",
          "description": "운동 빈도"
        }
      }
    },
    "initial_diagnosis": {
      "type": "string",
      "description": "초기 진단"
    },
    "recommended_tests": {
      "type": "string",
      "description": "추가로 추천되는 검사"
    },
    "advice": {
      "type": "string",
      "description": "환자를 위한 추가 조언"
    },
    "reason": {
      "type": "string",
      "description": "대화를 종료하려는 이유"
    }
  },
  "required": [
    "patient_info",
    "main_symptoms",
    "medical_history",
    "additional_symptoms",
    "lifestyle",
    "initial_diagnosis",
    "recommended_tests",
    "advice",
    "reason"
  ]
}
```