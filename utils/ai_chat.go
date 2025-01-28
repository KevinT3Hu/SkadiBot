package utils

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sashabaranov/go-openai"
)

type AIChatter struct {
	client          *openai.Client
	chatContext     []openai.ChatCompletionMessage
	contextLock     sync.Mutex
	lastRequestTime atomic.Int64
}

func NewAiChatter() (*AIChatter, error) {
	baseUrl := os.Getenv("AI_API_URL")
	apiKeyFile := os.Getenv("AI_API_KEY_FILE")
	apiKey, err := os.ReadFile(apiKeyFile)
	if err != nil {
		return nil, err
	}

	apiKeyS := strings.TrimSpace(string(apiKey))
	config := openai.DefaultConfig(apiKeyS)
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

	c.contextLock.Lock()
	defer c.contextLock.Unlock()
	c.chatContext = append(c.chatContext, m)

	// if the context is too long, remove the oldest message
	if len(c.chatContext) > 100 {
		c.chatContext = c.chatContext[1:]
	}
}

var (
	systemHint = os.Getenv("AI_SYSTEM_HINT")
	atHint     = os.Getenv("AI_AT_HINT")
)

func (c *AIChatter) GetAtRespond(ctx context.Context, msg string) (string, error) {
	response, err := c.getRespondWithPrompt(ctx, msg, atHint)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (c *AIChatter) getRespondWithPrompt(ctx context.Context, msg string, prompt string) (string, error) {
	if !c.getIsRequestable() {
		return "", errors.New("request too frequent")
	}
	AIRequestCounter.Inc()
	c.Feed(msg)
	c.contextLock.Lock()
	messages := append([]openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}, c.chatContext...)
	c.contextLock.Unlock()
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
	content := ret.Choices[0].Message.Content
	c.chatContext = append(c.chatContext, ret.Choices[0].Message)
	if len(c.chatContext) > 100 {
		c.chatContext = c.chatContext[1:]
	}
	SLogger.Info("AI response", "response", content, "source", "ai_chatter")
	return content, nil
}

func (c *AIChatter) GetRespond(ctx context.Context, msg string) (string, error) {
	response, err := c.getRespondWithPrompt(ctx, msg, systemHint)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (c *AIChatter) getIsRequestable() bool {
	return time.Now().Unix()-c.lastRequestTime.Load() > 30
}

func (c *AIChatter) updateRequestTime() {
	c.lastRequestTime.Store(time.Now().Unix())
}

func (c *AIChatter) GetChatContextLength() int {
	c.contextLock.Lock()
	defer c.contextLock.Unlock()
	return len(c.chatContext)
}

func (c *AIChatter) ClearChatContext() {
	c.contextLock.Lock()
	defer c.contextLock.Unlock()
	c.chatContext = []openai.ChatCompletionMessage{}
}
