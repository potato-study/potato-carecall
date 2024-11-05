package main

import (
	"errors"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"log"
	"potato-carecall/pkg/nestclient"
	"potato-carecall/pkg/openai"
	"potato-carecall/pkg/recorder"
)

const maxTurns = 10

func main() {
	initializePortAudio()
	defer portaudio.Terminate()

	messages := []openai.Message{openai.PromptMessage()}
	functions := openai.GetFunctionSpec()

	if err := recorder.InitRecordDevice(); err != nil {
		panic(err)
	}

	fmt.Printf("\n\n ===== %d번째 대화 =====\n", 0)
	showWelcomeMessage()

	for turns := 0; turns < maxTurns; turns++ {
		if err := processTurn(&messages, functions); errors.Is(err, openai.ErrEndConversation) {
			log.Println("finished conversation")
			break
		} else if err != nil {
			log.Printf("error processing turn: %v", err)
			break
		}
	}

	log.Println("대화를 종료합니다.")
}

// PortAudio 초기화
func initializePortAudio() {
	if err := portaudio.Initialize(); err != nil {
		log.Fatalf("Failed to initialize PortAudio: %v", err)
	}
}

func processTurn(messages *[]openai.Message, functions []openai.FunctionSpec) error {
	wavData, err := recorder.RecordAudio()
	if err != nil {
		return fmt.Errorf("error recording audio: %v", err)
	}

	client := nestclient.NewSTTClient()
	client.Params["assessment"] = "false"
	client.Params["utterance"] = "안녕하세요."
	client.Params["graph"] = "false"

	userInput, err := client.Recognize(wavData)
	if err != nil {
		return fmt.Errorf("error getting user input: %v", err)
	}

	userMessage := openai.Message{Role: "user", Content: userInput.Text}
	*messages = append(*messages, userMessage)
	printMessages("환자", userMessage)

	fmt.Printf("\n\n====%d====================\n", len(*messages))
	response, err := openai.Request(*messages, functions)
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

// 메시지 출력
func printMessages(prefix string, message openai.Message) {
	fmt.Printf("%s: %s\n", prefix, message.Content)
}

// 환영 메시지 출력
func showWelcomeMessage() {
	printMessages("의사", openai.Message{
		Role:    "assistant",
		Content: "안녕하세요, 저는 담당 의사입니다. 먼저 증상과 성별, 나이를 알려주시겠어요? (예: '저는 두통이 있어요. 여성, 30대')",
	})
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
