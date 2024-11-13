package dialogue

import (
	"potato-carecall/pkg/openai"
)

// 환영 메시지 출력
func ShowWelcomeMessage() {
	printMessages("의사", openai.Message{
		Role:    "assistant",
		Content: "안녕하세요, 저는 담당 의사입니다. 먼저 증상과 성별, 나이를 알려주시겠어요? (예: '저는 두통이 있어요. 여성, 30대')",
	})
}

func RunDialogue(messages *[]openai.Message) error {
	// 아래라인을 지우고 여기에 대화 로직을 작성하세요.
	panic("not implemented")
}

// 응답 처리
func processResponse(response *openai.ResponseBody) (openai.Message, error) {
	// 아래라인을 지우고 여기에 응답 처리 로직을 작성하세요.
	panic("not implemented")
}
