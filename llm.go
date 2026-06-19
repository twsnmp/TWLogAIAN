package main

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type AIAnswer struct {
	Answer string `json:"Answer"`
	Error  string `json:"Error"`
}

func (b *App) GetLLM(ctx context.Context) (llms.Model, error) {
	switch b.config.LLMProvider {
	case "ollama":
		baseURL := "http://localhost:11434"
		if b.config.LLMBaseURL != "" {
			baseURL = b.config.LLMBaseURL
		}
		return ollama.New(
			ollama.WithModel(b.config.LLMModel),
			ollama.WithServerURL(baseURL),
		)
	case "gemini", "googleai":
		opts := []googleai.Option{
			googleai.WithAPIKey(b.config.LLMAPIKey),
		}
		if b.config.LLMModel != "" {
			opts = append(opts, googleai.WithDefaultModel(b.config.LLMModel))
		}
		return googleai.New(ctx, opts...)
	case "openai":
		opts := []openai.Option{}
		if b.config.LLMModel != "" {
			opts = append(opts, openai.WithModel(b.config.LLMModel))
		}
		if b.config.LLMAPIKey != "" {
			opts = append(opts, openai.WithToken(b.config.LLMAPIKey))
		}
		if b.config.LLMBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(b.config.LLMBaseURL))
		}
		return openai.New(opts...)
	case "anthropic", "claude":
		opts := []anthropic.Option{}
		if b.config.LLMModel != "" {
			opts = append(opts, anthropic.WithModel(b.config.LLMModel))
		}
		if b.config.LLMAPIKey != "" {
			opts = append(opts, anthropic.WithToken(b.config.LLMAPIKey))
		}
		if b.config.LLMBaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(b.config.LLMBaseURL))
		}
		return anthropic.New(opts...)
	}
	return nil, fmt.Errorf("llm provider not found")
}

func (b *App) AskAIAboutLog(prompt, logStr, lang string) AIAnswer {
	r := AIAnswer{}
	ctx := b.ctx
	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("AskAIAboutLog err=%v", err)
		r.Error = err.Error()
		return r
	}
	system := `You are an expert in log analysis.
Please explain the log input by the user.`
	if lang == "ja" {
		system = `あなたはログ分析に関する専門家です。
ユーザーの入力したログについて解説してください。`
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf("Log:\n%s\n\nQuestion:\n%s", logStr, prompt)),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		OutLog("AskAIAboutLog err=%v", err)
		r.Error = err.Error()
		return r
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return r
	}
	r.Answer = resp.Choices[0].Content
	return r
}
