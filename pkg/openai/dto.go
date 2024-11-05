package openai

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type FunctionSpec struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

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
