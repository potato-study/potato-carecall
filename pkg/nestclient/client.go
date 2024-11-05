package nestclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"potato-carecall/pkg/config"
)

// STTClient 구조체 정의
type STTClient struct {
	APIKey string
	Lang   string
	Params map[string]string
}

type STTResponse struct {
	Text  string `json:"text"`
	Quota int    `json:"quota"`
}

// NewSTTClient 함수: STTClient 인스턴스를 생성합니다.
func NewSTTClient() *STTClient {
	return &STTClient{
		APIKey: config.ClovaSpeechApiKey,
		Lang:   "Kor",
		Params: make(map[string]string),
	}
}

// Recognize 함수: 오디오 데이터를 받아서 STT 결과를 반환합니다.
func (c *STTClient) Recognize(wavData []byte) (*STTResponse, error) {
	params := url.Values{"lang": {c.Lang}}
	for key, value := range c.Params {
		params.Add(key, value)
	}

	fullURL := fmt.Sprintf("%s?%s", config.ClovaSpeechUrl, params.Encode())
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(wavData))
	if err != nil {
		return nil, fmt.Errorf("요청 생성 중 오류 발생: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-CLOVASPEECH-API-KEY", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 중 오류 발생: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("API 오류: Response status", resp)
		return nil, fmt.Errorf("API 오류: Response status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 중 오류 발생: %v", err)
	}

	var sttResponse STTResponse
	if err := json.Unmarshal(body, &sttResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return &sttResponse, nil
}
