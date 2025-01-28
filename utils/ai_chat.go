package utils

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"time"

	"github.com/sashabaranov/go-openai"
)

type AIChatter struct {
	client          *openai.Client
	chatContext     []openai.ChatCompletionMessage
	lastRequestTime atomic.Int64
}

func NewAiChatter() (*AIChatter, error) {
	baseUrl := os.Getenv("AI_API_URL")
	apiKeyFile := os.Getenv("AI_API_KEY_FILE")
	apiKey, err := os.ReadFile(apiKeyFile)
	if err != nil {
		return nil, err
	}

	config := openai.DefaultConfig(string(apiKey))
	config.BaseURL = baseUrl
	client := openai.NewClientWithConfig(config)

	return &AIChatter{
		client: client,
	}, nil
}

// feed a message to the chat context without getting a response
func (c *AIChatter) Feed(msg string) {
	m := openai.ChatCompletionMessage{
		Role:    "user",
		Content: msg,
	}

	c.chatContext = append(c.chatContext, m)

	// if the context is too long, remove the oldest message
	if len(c.chatContext) > 100 {
		c.chatContext = c.chatContext[1:]
	}
}

var systemHint = openai.ChatCompletionMessage{
	Role:    "system",
	Content: os.Getenv("AI_SYSTEM_HINT"),
}

func (c *AIChatter) GetRespond(ctx context.Context, msg string) (string, error) {
	if !c.getIsRequestable() {
		return "", errors.New("request too frequent")
	}
	AIRequestCounter.Inc()
	c.Feed(msg)
	messages := append([]openai.ChatCompletionMessage{systemHint}, c.chatContext...)
	req := openai.ChatCompletionRequest{
		Model:               os.Getenv("AI_MODEL"),
		Messages:            messages,
		MaxCompletionTokens: 100,
	}
	c.updateRequestTime()
	reqStartTime := time.Now()
	ret, err := c.client.CreateChatCompletion(ctx, req)
	reqTime := time.Since(reqStartTime).Milliseconds()
	AIRequestLatency.Observe(float64(reqTime))
	if err != nil {
		return "", err
	}
	return ret.Choices[0].Message.Content, nil
}

func (c *AIChatter) getIsRequestable() bool {
	return time.Now().Unix()-c.lastRequestTime.Load() > 30
}

func (c *AIChatter) updateRequestTime() {
	c.lastRequestTime.Store(time.Now().Unix())
}
