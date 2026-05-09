package utils

import (
	"context"
	"fmt"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"llm_ocr_api_go/config"
)

var (
	client     *openai.Client
	clientOnce sync.Once
)

// getOpenAIClient 获取或初始化 OpenAI 兼容客户端
func getOpenAIClient() *openai.Client {
	clientOnce.Do(func() {
		cfg := config.LoadConfig()
		clientConfig := openai.DefaultConfig(cfg.DashScope.ApiKey)
		clientConfig.BaseURL = cfg.DashScope.BaseURL
		client = openai.NewClientWithConfig(clientConfig)
	})
	return client
}

// CallDashScopeOCR 调用 DashScope API 进行 OCR 识别，prompt 为提示词
func CallDashScopeOCR(ctx context.Context, prompt string, imageBase64 string) (string, error) {
	cfg := config.LoadConfig()
	cli := getOpenAIClient()
	imageURL := fmt.Sprintf("data:image/png;base64,%s", imageBase64)

	resp, err := cli.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: cfg.DashScope.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: prompt,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: imageURL,
							},
						},
					},
				},
			},
			Temperature: float32(0),
		},
	)

	if err != nil {
		return "", fmt.Errorf("调用 DashScope API 失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("API 返回空响应")
	}

	return resp.Choices[0].Message.Content, nil
}
