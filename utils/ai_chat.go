package utils

import (
	"context"
	"errors"
	"os"
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

func NewAiChatter(apiConfig AIApiConfig) *AIChatter {
	config := openai.DefaultConfig(apiConfig.APIKey)
	config.BaseURL = apiConfig.BaseUrl
	client := openai.NewClientWithConfig(config)

	return &AIChatter{
		client: client,
	}
}

func (c *AIChatter) NewConfig(apiConfig AIApiConfig) {
	config := openai.DefaultConfig(apiConfig.APIKey)
	config.BaseURL = apiConfig.BaseUrl
	c.client = openai.NewClientWithConfig(config)
}

// feed a message to the chat context without getting a response
func (c *AIChatter) Feed(msg string) {
	m := openai.ChatCompletionMessage{
		Role:    "user",
		Content: msg,
	}

	c.contextLock.Lock()
	c.chatContext = append(c.chatContext, m)
	c.contextLock.Unlock()

	c.truncateChatContext()
}

func (c *AIChatter) GetAtRespond(ctx context.Context, msg string) (string, error) {
	hint := GetConfig().PromptConfig.AIAtRequestPrompt
	response, err := c.getRespondWithPrompt(ctx, msg, hint)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (c *AIChatter) truncateChatContext() {
	c.contextLock.Lock()
	defer c.contextLock.Unlock()
	if len(c.chatContext) >= 100 {
		c.chatContext = c.chatContext[81:]
	}
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
	c.contextLock.Lock()
	content := ret.Choices[0].Message.Content
	c.chatContext = append(c.chatContext, ret.Choices[0].Message)
	c.contextLock.Unlock()
	c.truncateChatContext()
	SLogger.Info("AI response", "response", content, "source", "ai_chatter")
	return content, nil
}

func (c *AIChatter) GetRespond(ctx context.Context, msg string) (string, error) {
	systemHint := GetConfig().PromptConfig.AIRequestPrompt
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
