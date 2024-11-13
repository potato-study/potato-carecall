package dialogue

import (
	"fmt"
	"potato-carecall/pkg/openai"
)

// 메시지 출력
func printMessages(prefix string, message openai.Message) {
	fmt.Printf("%s: %s\n", prefix, message.Content)
}
