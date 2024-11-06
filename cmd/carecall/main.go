package main

import (
	"errors"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"log"
	"potato-carecall/pkg/dialogue"
	"potato-carecall/pkg/openai"
	"potato-carecall/pkg/recorder"
)

const maxDialogue = 20

func main() {
	// 녹음 장치 초기화
	if err := portaudio.Initialize(); err != nil {
		log.Fatalf("Failed to initialize PortAudio: %v", err)
	}
	defer portaudio.Terminate()

	if err := recorder.InitRecordDevice(); err != nil {
		panic(err)
	}

	// 대화 메시지 초기화
	messages := []openai.Message{openai.PromptMessage()}

	// 함수 명세 가져오기
	fmt.Printf("\n\n ===== %d번째 대화 =====\n", 0)
	dialogue.ShowWelcomeMessage()

	for turns := 0; turns < maxDialogue; turns++ {
		if err := dialogue.RunDialogue(&messages); errors.Is(err, openai.ErrEndConversation) {
			log.Println("finished conversation")
			break
		} else if err != nil {
			log.Printf("error processing turn: %v", err)
			break
		}
	}

	log.Println("대화를 종료합니다.")
}
