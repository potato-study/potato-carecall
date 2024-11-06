package dialogue

import (
	"errors"
	"fmt"
	"log"
	"potato-carecall/pkg/nestclient"
	"potato-carecall/pkg/openai"
	"potato-carecall/pkg/recorder"
)

// 환영 메시지 출력
func ShowWelcomeMessage() {
	printMessages("의사", openai.Message{
		Role:    "assistant",
		Content: "안녕하세요, 저는 담당 의사입니다. 먼저 증상과 성별, 나이를 알려주시겠어요? (예: '저는 두통이 있어요. 여성, 30대')",
	})
}

func RunDialogue(messages *[]openai.Message) error {
	// 오디오 녹음
	wavData, err := recorder.RecordAudio()
	if err != nil {
		return fmt.Errorf("error recording audio: %v", err)
	}

	client := nestclient.NewSTTClient()
	client.Params["assessment"] = "false" // 발음평가
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

// 응답 처리
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
