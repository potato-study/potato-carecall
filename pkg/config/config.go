package config

import "os"

var (
	OpenaiModel       string
	OpenaiApiUrl      string
	OpenaiApiKey      string
	ClovaSpeechApiKey string
	ClovaSpeechUrl    string
)

func init() {
	OpenaiModel = "gpt-4o-2024-08-06"
	OpenaiApiUrl = "https://api.openai.com/v1/chat/completions"
	OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
	ClovaSpeechUrl = "https://clovaspeech-gw.ncloud.com/recog/v1/stt"
	ClovaSpeechApiKey = os.Getenv("CLOVA_SPEECH_API_KEY")
}
