package main

import (
	"context"
	"fmt"

	"regexp"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/twsnmp/TWLogAIAN/pkg/ai/tensai"
	"github.com/twsnmp/TWLogAIAN/pkg/model"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
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
	case "tensai", "embedded", "local":
		modelPath, err := model.FindModel("", b.config.LLMModel)
		if err != nil {
			return nil, fmt.Errorf("local model not found: %w", err)
		}
		return tensai.NewWithOptions(modelPath, false)
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

	cleanPrompt := strings.TrimSpace(prompt)
	var system string
	var humanContent string

	if lang == "ja" {
		system = `あなたはサイバーセキュリティおよびログ分析の専門家です。
ユーザーから提示されたログを詳細に分析し、日本語で分かりやすく解説してください。
見出し、箇条書き、太字、コードブロックなどを適切に活用し、Markdown形式で整形して回答してください。`

		if cleanPrompt == "" {
			cleanPrompt = "このログの内容について詳しく解説してください。ログの意味、重要な要素、エラーや異常の有無、想定される原因と推奨される対応をわかりやすく説明してください。"
		}
		humanContent = fmt.Sprintf("対象ログ:\n```\n%s\n```\n\n質問・指示:\n%s\n\n回答は日本語のMarkdown形式で記述してください。", logStr, cleanPrompt)
	} else {
		system = `You are an expert in cybersecurity and log analysis.
Please analyze the log provided by the user in detail and provide a clear, well-structured explanation in English.
Format your response using Markdown (use headings, bullet points, bold text, code blocks, etc.).`

		if cleanPrompt == "" {
			cleanPrompt = "Please explain this log in detail. Describe its meaning, key components, whether there are anomalies or errors, potential root causes, and recommended actions."
		}
		humanContent = fmt.Sprintf("Target Log:\n```\n%s\n```\n\nQuestion/Instruction:\n%s\n\nPlease provide your answer in English using Markdown formatting.", logStr, cleanPrompt)
	}

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, humanContent),
	}

	var sb strings.Builder
	resp, err := llm.GenerateContent(ctx, history, llms.WithStreamingFunc(func(c context.Context, chunk []byte) error {
		chunkStr := string(chunk)
		sb.WriteString(chunkStr)
		if b.ctx != nil {
			wails.EventsEmit(b.ctx, "ask_ai_stream", chunkStr)
		}
		return nil
	}))
	if err != nil {
		OutLog("AskAIAboutLog err=%v", err)
		r.Error = err.Error()
		return r
	}
	if resp != nil && len(resp.Choices) > 0 && resp.Choices[0].Content != "" {
		r.Answer = resp.Choices[0].Content
	} else {
		r.Answer = sb.String()
	}
	if r.Answer == "" {
		r.Error = "no response from LLM"
	}
	return r
}

// MaskPII replaces sensitive information (IPs, MACs, emails, domains, credentials, etc.) with consistent masked placeholders.
func MaskPII(logStr string) string {
	res := logStr

	// Map to keep track of replacements consistently within a log
	ipMap := make(map[string]string)
	macMap := make(map[string]string)
	emailMap := make(map[string]string)
	hostMap := make(map[string]string)
	userMap := make(map[string]string)

	// 1. Mask Secrets/Passwords/Tokens (e.g. password=xyz, token=abc)
	secretRe := regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|api_?key|secret|auth|bearer|private_?key)\s*([:=])\s*([^\s,;]+)`)
	res = secretRe.ReplaceAllString(res, `$1$2[REDACTED]`)

	// 2. Mask Emails
	emailRe := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	res = emailRe.ReplaceAllStringFunc(res, func(m string) string {
		if val, ok := emailMap[m]; ok {
			return val
		}
		tag := fmt.Sprintf("[EMAIL_%d]", len(emailMap)+1)
		emailMap[m] = tag
		return tag
	})

	// 3. Mask MAC Addresses
	macRe := regexp.MustCompile(`\b(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`)
	res = macRe.ReplaceAllStringFunc(res, func(m string) string {
		if val, ok := macMap[m]; ok {
			return val
		}
		tag := fmt.Sprintf("[MAC_%d]", len(macMap)+1)
		macMap[m] = tag
		return tag
	})

	// 4. Mask IPv4 Addresses
	ipv4Re := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	res = ipv4Re.ReplaceAllStringFunc(res, func(m string) string {
		if val, ok := ipMap[m]; ok {
			return val
		}
		tag := fmt.Sprintf("[IP_%d]", len(ipMap)+1)
		ipMap[m] = tag
		return tag
	})

	// 5. Mask IPv6 Addresses (standard / full)
	ipv6Re := regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)
	res = ipv6Re.ReplaceAllStringFunc(res, func(m string) string {
		if val, ok := ipMap[m]; ok {
			return val
		}
		tag := fmt.Sprintf("[IP_%d]", len(ipMap)+1)
		ipMap[m] = tag
		return tag
	})

	// 6. Mask Domains / FQDNs
	domainRe := regexp.MustCompile(`\b(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+(?:com|net|org|edu|gov|io|co|jp|cn|de|uk|info|biz|me|xyz|tech|online|dev|site|internal|local|corp|domain)\b`)
	res = domainRe.ReplaceAllStringFunc(res, func(m string) string {
		if val, ok := hostMap[m]; ok {
			return val
		}
		tag := fmt.Sprintf("[HOST_%d]", len(hostMap)+1)
		hostMap[m] = tag
		return tag
	})

	// 7. Mask Users in common patterns: e.g. "user <name>", "user=<name>", "user: <name>", "for user <name>", "from user <name>"
	userRe := regexp.MustCompile(`(?i)\b(user(?:name)?|for user|from user)\s*[:=\s]\s*([a-zA-Z0-9._-]+)`)
	res = userRe.ReplaceAllStringFunc(res, func(m string) string {
		sub := userRe.FindStringSubmatch(m)
		if len(sub) == 3 {
			prefix := sub[1]
			uname := sub[2]
			if uname == "[REDACTED]" || strings.HasPrefix(uname, "[") {
				return m
			}
			tag, ok := userMap[uname]
			if !ok {
				tag = fmt.Sprintf("[USER_%d]", len(userMap)+1)
				userMap[uname] = tag
			}
			sep := " "
			if strings.Contains(m, "=") {
				sep = "="
			} else if strings.Contains(m, ":") {
				sep = ": "
			}
			return prefix + sep + tag
		}
		return m
	})

	return res
}

func (b *App) MaskPII(logStr string) string {
	return MaskPII(logStr)
}
