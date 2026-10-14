package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/twsnmp/TWLogAIAN/pkg/ai/tensai"
	"github.com/twsnmp/TWLogAIAN/pkg/model"
	"github.com/vjeantet/grok"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type utf8StreamDecoder struct {
	buf []byte
}

func (d *utf8StreamDecoder) decode(chunk []byte) string {
	d.buf = append(d.buf, chunk...)
	validLen := 0
	for validLen < len(d.buf) {
		r, size := utf8.DecodeRune(d.buf[validLen:])
		if r == utf8.RuneError && size == 1 {
			if !utf8.FullRune(d.buf[validLen:]) {
				break
			}
		}
		validLen += size
	}
	if validLen == 0 {
		return ""
	}
	validBytes := d.buf[:validLen]
	d.buf = d.buf[validLen:]
	return string(validBytes)
}

func (d *utf8StreamDecoder) flush() string {
	if len(d.buf) == 0 {
		return ""
	}
	res := string(d.buf)
	d.buf = nil
	return res
}

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
			OutLog("GetLLM tensai: local model %q not found: %v", b.config.LLMModel, err)
			return nil, fmt.Errorf("local model not found (%s): %w", b.config.LLMModel, err)
		}
		llm, err := tensai.NewWithOptions(modelPath, false)
		if err != nil {
			OutLog("GetLLM tensai: failed to load model %s: %v", modelPath, err)
			return nil, err
		}
		return llm, nil
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
	outputLang := "English"
	defaultPrompt := "Please explain this log in detail. Describe its meaning, key components, whether there are anomalies or errors, potential root causes, and recommended actions."
	if lang == "ja" {
		outputLang = "Japanese"
		defaultPrompt = "このログの内容について詳しく解説してください。ログの意味、重要な要素、エラーや異常の有無、想定される原因と推奨される対応をわかりやすく説明してください。"
	}

	if cleanPrompt == "" {
		cleanPrompt = defaultPrompt
	}

	system := fmt.Sprintf(`You are an expert in cybersecurity, system administration, and log analysis.
Analyze the target log entry provided by the user in detail and provide a clear, well-structured, actionable explanation.
Structure your analysis using Markdown (headings, bullet points, bold text, and code blocks).
You MUST write your entire response in %s.`, outputLang)

	humanContent := fmt.Sprintf("Target Log:\n```\n%s\n```\n\nQuestion / Request:\n%s\n\nPlease provide your analysis in %s using Markdown.", logStr, cleanPrompt, outputLang)

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, humanContent),
	}

	var sb strings.Builder
	decoder := &utf8StreamDecoder{}
	resp, err := llm.GenerateContent(ctx, history, llms.WithStreamingFunc(func(c context.Context, chunk []byte) error {
		chunkStr := decoder.decode(chunk)
		if chunkStr != "" {
			sb.WriteString(chunkStr)
			if b.ctx != nil {
				wails.EventsEmit(b.ctx, "ask_ai_stream", chunkStr)
			}
		}
		return nil
	}))
	if tail := decoder.flush(); tail != "" {
		sb.WriteString(tail)
		if b.ctx != nil {
			wails.EventsEmit(b.ctx, "ask_ai_stream", tail)
		}
	}
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

// AI Analysis Report types and functions

type AIAnalysisReportRequest struct {
	Prompt     string `json:"Prompt"`
	Lang       string `json:"Lang"`
	MaskPII    bool   `json:"MaskPII"`
	TopNError  int    `json:"TopNError"`
	SampleSize int    `json:"SampleSize"`
}

type AIAnalysisReportResponse struct {
	Summary  AIReportSummary `json:"Summary"`
	Markdown string          `json:"Markdown"`
	Stats    string          `json:"Stats"`
	Error    string          `json:"Error"`
}

type AIReportSummary struct {
	TotalEntries int              `json:"TotalEntries"`
	ErrorCount   int              `json:"ErrorCount"`
	WarningCount int              `json:"WarningCount"`
	StartTime    int64            `json:"StartTime"`
	EndTime      int64            `json:"EndTime"`
	TopErrors    []AIErrorPattern `json:"TopErrors"`
}

type AIErrorPattern struct {
	Pattern string `json:"Pattern"`
	Count   int    `json:"Count"`
	Example string `json:"Example"`
}

var (
	regTimestamp = regexp.MustCompile(`\b\d{4}[-/]\d{2}[-/]\d{2}[T\s]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})?\b|[A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}`)
	regUUID      = regexp.MustCompile(`\b[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\b`)
	regEmailNorm = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	regIPNorm    = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	regMACNorm   = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}[0-9a-fA-F]{2}\b`)
	regNumNorm   = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
)

func normalizeLog(msg string) string {
	res := regTimestamp.ReplaceAllString(msg, "#TIMESTAMP#")
	res = regUUID.ReplaceAllString(res, "#UUID#")
	res = regEmailNorm.ReplaceAllString(res, "#EMAIL#")
	res = regIPNorm.ReplaceAllString(res, "#IP#")
	res = regMACNorm.ReplaceAllString(res, "#MAC#")
	res = regNumNorm.ReplaceAllString(res, "#NUM#")
	return res
}

func isErrorWord(s string) bool {
	return strings.Contains(s, "error") ||
		strings.Contains(s, "fatal") ||
		strings.Contains(s, "fail") ||
		strings.Contains(s, "crit") ||
		strings.Contains(s, "alert") ||
		strings.Contains(s, "emerg") ||
		strings.Contains(s, "err ") ||
		strings.Contains(s, "err:") ||
		strings.Contains(s, "denied")
}

func getLogLevelString(l *LogEnt) string {
	if l == nil {
		return "INFO"
	}
	if l.KeyValue != nil {
		// 1. Direct _level attribute from frontend getLogLevel
		if lvl, ok := l.KeyValue["_level"].(string); ok && lvl != "" {
			switch strings.ToLower(lvl) {
			case "error":
				return "ERROR"
			case "warn":
				return "WARN"
			case "debug":
				return "DEBUG"
			default:
				return "INFO"
			}
		}

		// 2. Syslog severity / priority (PRI)
		var suverityNum float64
		hasSuverity := false
		if s, ok := l.KeyValue["suverity"].(float64); ok {
			suverityNum = s
			hasSuverity = true
		} else if p, ok := l.KeyValue["priority"].(float64); ok {
			suverityNum = p
			hasSuverity = true
		}
		if hasSuverity {
			mod := int(suverityNum) % 8
			if mod < 4 {
				return "ERROR"
			} else if mod == 4 {
				return "WARN"
			}
			return "INFO"
		}

		// 3. HTTP response code
		if code, ok := l.KeyValue["response"].(float64); ok && code > 99 {
			if code >= 400 {
				return "ERROR"
			} else if code >= 300 {
				return "WARN"
			}
			return "INFO"
		}

		// 4. Windows winLevel
		if wl, ok := l.KeyValue["winLevel"].(float64); ok {
			iwl := int(wl)
			if iwl == 1 || iwl == 2 || iwl == 100 {
				return "ERROR"
			} else if iwl == 3 || iwl == 101 {
				return "WARN"
			}
			return "INFO"
		}

		// 5. Normal / Warn / Error fields
		var normalCount, warnCount, errorCount float64
		if n, ok := l.KeyValue["Normal"].(float64); ok {
			normalCount = n
		} else if n, ok := l.KeyValue["normal"].(float64); ok {
			normalCount = n
		}
		if w, ok := l.KeyValue["Warn"].(float64); ok {
			warnCount = w
		} else if w, ok := l.KeyValue["warn"].(float64); ok {
			warnCount = w
		}
		if e, ok := l.KeyValue["Error"].(float64); ok {
			errorCount = e
		} else if e, ok := l.KeyValue["error"].(float64); ok {
			errorCount = e
		}
		if normalCount == 0 && errorCount > 0 {
			return "ERROR"
		}
		if normalCount == 0 && warnCount > 0 {
			return "WARN"
		}

		// 6. Explicit string levels
		levelFields := []string{"suverity_str", "severity", "level", "log_level", "status"}
		for _, lf := range levelFields {
			if v, ok := l.KeyValue[lf].(string); ok && v != "" {
				vLower := strings.ToLower(v)
				if isErrorWord(vLower) {
					return "ERROR"
				}
				if strings.Contains(vLower, "warn") {
					return "WARN"
				}
			}
		}
	}

	// 7. Check text of log
	allLower := strings.ToLower(l.All)
	if isErrorWord(allLower) {
		return "ERROR"
	}
	if strings.Contains(allLower, "warn") {
		return "WARN"
	}
	if strings.Contains(allLower, "debug") {
		return "DEBUG"
	}
	return "INFO"
}

func (b *App) formatAIStats(dur time.Duration, tokens int, llm llms.Model) string {
	sec := dur.Seconds()
	var speedStr string
	if tokens > 0 && sec > 0.01 {
		tokPerSec := float64(tokens) / sec
		speedStr = fmt.Sprintf("%.1f tok/s (%d tokens in %s)", tokPerSec, tokens, dur.Truncate(time.Millisecond))
	} else {
		speedStr = fmt.Sprintf("%s", dur.Truncate(time.Millisecond))
	}
	if tllm, ok := llm.(*tensai.TensaiLLM); ok {
		accType, _ := tllm.Acceleration()
		return fmt.Sprintf("%s | %s", accType, speedStr)
	}
	if b.config.LLMProvider != "" {
		providerName := b.config.LLMProvider
		if b.config.LLMModel != "" {
			providerName += fmt.Sprintf(" (%s)", b.config.LLMModel)
		}
		return fmt.Sprintf("%s | %s", providerName, speedStr)
	}
	return speedStr
}

func buildAIPromptJA(req AIAnalysisReportRequest, summary AIReportSummary, samples []string) (string, string) {
	system := "You are an expert in cybersecurity, systems administration, and log analysis. Analyze the provided log dataset in detail and generate a practical, professional analysis report in Markdown format. You MUST write your entire report in Japanese."

	var sb strings.Builder
	sb.WriteString("Analyze the provided log dataset in detail and produce a structured, actionable analysis report in Japanese using Markdown format.\n\n")

	if req.Prompt != "" {
		sb.WriteString("【特別な指示・分析リクエスト / Custom Request】\n")
		sb.WriteString(req.Prompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("【ログ概要データ / Log Summary Data】\n")
	sb.WriteString(fmt.Sprintf("- 総ログ件数: %d 件\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- エラーログ件数: %d 件\n", summary.ErrorCount))
	sb.WriteString(fmt.Sprintf("- 警告ログ件数: %d 件\n", summary.WarningCount))
	if summary.StartTime > 0 && summary.EndTime > 0 {
		st := time.Unix(0, summary.StartTime).Format("2006-01-02 15:04:05")
		et := time.Unix(0, summary.EndTime).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("- 対象期間: %s 〜 %s\n", st, et))
	}

	if len(summary.TopErrors) > 0 {
		sb.WriteString("\n【頻出エラーパターン（上位） / Top Error Patterns】\n")
		for i, ep := range summary.TopErrors {
			sb.WriteString(fmt.Sprintf("%d. 発生回数: %d 回\n   パターン: `%s`\n   具体例: `%s`\n", i+1, ep.Count, ep.Pattern, ep.Example))
		}
	}

	if len(samples) > 0 {
		sb.WriteString("\n【ログサンプル（最新・代表的な抜粋） / Log Samples】\n```\n")
		for _, s := range samples {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString(`[Report Requirements]
Structure your report using Markdown (headings, bullet points, bold text, code blocks) with the following sections:
1. **エグゼクティブ・サマリー (Executive Summary)**: ログ全体の傾向、健全性評価、重大事象の要約
2. **検出された異常・リスク (Detected Anomalies & Risks)**:
   - 種類 (error_spike / security / performance / other)
   - 重要度 (Critical / High / Medium / Low)
   - 具体的な兆候と該当ログ
3. **根本原因の考察 (Root Cause Analysis)**: エラーや異常が発生している要因の分析
4. **推奨アクション・再発防止策 (Actionable Recommendations)**:
   - 即時対応 (Immediate mitigations)
   - 恒久対策 (Long-term preventative measures)

You MUST write your entire analysis and report in Japanese.`)

	return system, sb.String()
}

func buildAIPromptEN(req AIAnalysisReportRequest, summary AIReportSummary, samples []string) (string, string) {
	system := "You are an expert in cybersecurity, systems administration, and log analysis. You must write your entire analysis and report strictly in English. Even if the log entries, user prompt, or metadata are written in Japanese or other languages, translate and provide all explanations, headings, and recommendations purely in English without using Japanese characters."

	var sb strings.Builder
	sb.WriteString("Analyze the provided log dataset in detail and produce a structured, actionable analysis report in Markdown format.\n\n")

	if req.Prompt != "" {
		sb.WriteString("[Custom Instruction / Request]\n")
		sb.WriteString(req.Prompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("[Log Summary Data]\n")
	sb.WriteString(fmt.Sprintf("- Total Log Count: %d\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- Error Count: %d\n", summary.ErrorCount))
	sb.WriteString(fmt.Sprintf("- Warning Count: %d\n", summary.WarningCount))
	if summary.StartTime > 0 && summary.EndTime > 0 {
		st := time.Unix(0, summary.StartTime).Format("2006-01-02 15:04:05")
		et := time.Unix(0, summary.EndTime).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("- Time Range: %s to %s\n", st, et))
	}

	if len(summary.TopErrors) > 0 {
		sb.WriteString("\n[Top Error Patterns]\n")
		for i, ep := range summary.TopErrors {
			sb.WriteString(fmt.Sprintf("%d. Count: %d occurrences\n   Pattern: `%s`\n   Example: `%s`\n", i+1, ep.Count, ep.Pattern, ep.Example))
		}
	}

	if len(samples) > 0 {
		sb.WriteString("\n[Recent Log Samples]\n```\n")
		for _, s := range samples {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString(`[Report Requirements]
Please structure your response using Markdown (headings, bullet points, bold text, code blocks) with the following sections:
1. **Executive Summary**: Overall trend, system health assessment, and key highlights.
2. **Detected Anomalies & Risks**:
   - Type (error_spike / security / performance / other)
   - Severity (Critical / High / Medium / Low)
   - Evidence and sample logs
3. **Root Cause Analysis**: Potential underlying causes for the errors/anomalies.
4. **Actionable Recommendations**:
   - Immediate mitigations
   - Long-term preventative measures

Please write your entire response strictly in English.`)

	return system, sb.String()
}

// CancelAIReport cancels any running AI analysis generation.
func (b *App) CancelAIReport() {
	b.aiMu.Lock()
	defer b.aiMu.Unlock()
	if b.aiCancel != nil {
		b.aiCancel()
		b.aiCancel = nil
	}
}

// GetLogSummary computes log statistics and extracts top error patterns without calling LLM.
func (b *App) GetLogSummary(req AIAnalysisReportRequest, logs []*LogEnt) AIReportSummary {
	if len(logs) == 0 {
		return AIReportSummary{
			TopErrors: []AIErrorPattern{},
		}
	}
	if req.TopNError <= 0 {
		req.TopNError = 10
	}

	summary := AIReportSummary{
		TotalEntries: len(logs),
		TopErrors:    []AIErrorPattern{},
	}
	errorLogMap := make(map[string]*AIErrorPattern)

	var minTime int64 = logs[0].Time
	var maxTime int64 = logs[0].Time

	for _, l := range logs {
		if l.Time > 0 {
			if minTime == 0 || l.Time < minTime {
				minTime = l.Time
			}
			if l.Time > maxTime {
				maxTime = l.Time
			}
		}
		level := getLogLevelString(l)
		switch level {
		case "ERROR":
			summary.ErrorCount++
			nl := normalizeLog(l.All)
			if ep, ok := errorLogMap[nl]; ok {
				ep.Count++
			} else {
				ex := l.All
				if req.MaskPII {
					ex = MaskPII(ex)
				}
				errorLogMap[nl] = &AIErrorPattern{
					Pattern: nl,
					Count:   1,
					Example: ex,
				}
			}
		case "WARN":
			summary.WarningCount++
		}
	}
	summary.StartTime = minTime
	summary.EndTime = maxTime

	// Sort error patterns by count descending
	patterns := make([]*AIErrorPattern, 0, len(errorLogMap))
	for _, ep := range errorLogMap {
		patterns = append(patterns, ep)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	topN := req.TopNError
	if len(patterns) < topN {
		topN = len(patterns)
	}
	for i := 0; i < topN; i++ {
		summary.TopErrors = append(summary.TopErrors, *patterns[i])
	}
	return summary
}

// AnalyzeLogsWithAI aggregates statistics, extracts top error patterns, and streams AI analysis.
func (b *App) AnalyzeLogsWithAI(req AIAnalysisReportRequest, logs []*LogEnt) AIAnalysisReportResponse {
	if len(logs) == 0 {
		return AIAnalysisReportResponse{
			Error: "No logs to analyze",
		}
	}

	// Cancel any previously active generation
	b.CancelAIReport()

	parentCtx := b.ctx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	b.aiMu.Lock()
	b.aiCancel = cancel
	b.aiMu.Unlock()

	defer func() {
		b.aiMu.Lock()
		cancel()
		b.aiCancel = nil
		b.aiMu.Unlock()
	}()

	if req.Lang == "" {
		req.Lang = "ja"
	}
	if req.TopNError <= 0 {
		req.TopNError = 10
	}
	if req.SampleSize <= 0 {
		req.SampleSize = 50
	}

	summary := b.GetLogSummary(req, logs)
	if b.ctx != nil {
		wails.EventsEmit(b.ctx, "ai_report_summary", summary)
	}

	// Sample logs from the end (most recent)
	sampleSize := req.SampleSize
	if len(logs) < sampleSize {
		sampleSize = len(logs)
	}
	sampleLogs := logs[len(logs)-sampleSize:]
	samples := make([]string, 0, len(sampleLogs))
	for _, l := range sampleLogs {
		msg := l.All
		if req.MaskPII {
			msg = MaskPII(msg)
		}
		tStr := ""
		if l.Time > 0 {
			tStr = time.Unix(0, l.Time).Format("2006-01-02 15:04:05") + " "
		}
		samples = append(samples, fmt.Sprintf("%s[%s] %s", tStr, getLogLevelString(l), msg))
	}

	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("AnalyzeLogsWithAI GetLLM err=%v", err)
		return AIAnalysisReportResponse{
			Summary: summary,
			Error:   err.Error(),
		}
	}

	var systemText, humanText string
	if req.Lang == "en" {
		systemText, humanText = buildAIPromptEN(req, summary, samples)
	} else {
		systemText, humanText = buildAIPromptJA(req, summary, samples)
	}

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemText),
		llms.TextParts(llms.ChatMessageTypeHuman, humanText),
	}

	var sb strings.Builder
	var tokenCount int
	st := time.Now()
	decoder := &utf8StreamDecoder{}

	resp, err := llm.GenerateContent(ctx, history, llms.WithStreamingFunc(func(c context.Context, chunk []byte) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		tokenCount++
		chunkStr := decoder.decode(chunk)
		if chunkStr != "" {
			sb.WriteString(chunkStr)
			if b.ctx != nil {
				wails.EventsEmit(b.ctx, "ai_report_stream", chunkStr)
			}
		}
		return nil
	}))
	if tail := decoder.flush(); tail != "" {
		if ctx.Err() == nil {
			sb.WriteString(tail)
			if b.ctx != nil {
				wails.EventsEmit(b.ctx, "ai_report_stream", tail)
			}
		}
	}

	dur := time.Since(st)
	stats := b.formatAIStats(dur, tokenCount, llm)

	var markdown string
	if resp != nil && len(resp.Choices) > 0 && resp.Choices[0].Content != "" {
		markdown = resp.Choices[0].Content
	} else {
		markdown = sb.String()
	}

	if err != nil {
		if ctx.Err() == context.Canceled {
			return AIAnalysisReportResponse{
				Error: "Analysis canceled",
			}
		}
		OutLog("AnalyzeLogsWithAI GenerateContent err=%v", err)
		return AIAnalysisReportResponse{
			Summary:  summary,
			Markdown: markdown,
			Stats:    stats,
			Error:    err.Error(),
		}
	}

	return AIAnalysisReportResponse{
		Summary:  summary,
		Markdown: markdown,
		Stats:    stats,
	}
}

// AIGrokResp represents the result of AI-generated Grok pattern
type AIGrokResp struct {
	Grok      string `json:"Grok"`
	TimeField string `json:"TimeField"`
	IPFields  string `json:"IPFields"`
	MACFields string `json:"MACFields"`
	ErrorMsg  string `json:"ErrorMsg"`
}

type aiGrokJSON struct {
	Grok      string `json:"grok"`
	TimeField string `json:"time_field"`
	IPFields  string `json:"ip_fields"`
	MACFields string `json:"mac_fields"`
}

// parseAIGrokResponse parses the raw LLM output into AIGrokResp
func parseAIGrokResponse(rawResp string) AIGrokResp {
	resp := AIGrokResp{}
	clean := strings.TrimSpace(rawResp)

	// 1. Try to find JSON object {...}
	startIdx := strings.Index(clean, "{")
	endIdx := strings.LastIndex(clean, "}")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonCandidate := clean[startIdx : endIdx+1]
		var parsed aiGrokJSON
		if err := json.Unmarshal([]byte(jsonCandidate), &parsed); err == nil && parsed.Grok != "" {
			resp.Grok = strings.TrimSpace(parsed.Grok)
			resp.TimeField = strings.TrimSpace(parsed.TimeField)
			resp.IPFields = strings.TrimSpace(parsed.IPFields)
			resp.MACFields = strings.TrimSpace(parsed.MACFields)
			return resp
		}
	}

	// 2. Fallback: extract from markdown code blocks
	codeBlockRe := regexp.MustCompile("(?s)```(?:grok)?\n?(.*?)\n?```")
	matches := codeBlockRe.FindStringSubmatch(clean)
	if len(matches) > 1 && strings.TrimSpace(matches[1]) != "" {
		candidate := strings.TrimSpace(matches[1])
		if strings.Contains(candidate, "%{") {
			resp.Grok = candidate
			return resp
		}
	}

	// 3. Fallback: find line containing %{...}
	for _, line := range strings.Split(clean, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.Trim(trimmed, "`\"'")
		if strings.Contains(trimmed, "%{") {
			resp.Grok = trimmed
			return resp
		}
	}

	resp.Grok = clean
	return resp
}

// AIGrok generates a Grok extraction pattern using LLM based on test log
func (b *App) AIGrok(testData, lang string) AIGrokResp {
	res := AIGrokResp{}
	if b.config.LLMProvider == "" || b.config.LLMProvider == "none" {
		res.ErrorMsg = "LLM provider is not configured"
		return res
	}

	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("AIGrok GetLLM err=%v", err)
		res.ErrorMsg = err.Error()
		return res
	}

	trimmedLog := strings.TrimSpace(testData)
	if trimmedLog == "" {
		res.ErrorMsg = "Test log data is required"
		return res
	}

	// Extract first non-empty line as main sample
	sampleLine := ""
	for _, l := range strings.Split(trimmedLog, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			sampleLine = l
			break
		}
	}

	systemPrompt := `You are an expert in log analysis and Logstash/Grok patterns.
Analyze the given target log line and generate a Logstash Grok pattern that extracts all key fields (timestamp, host, process, PID, IP, MAC, user, status, numbers, message) using %{PATTERN:field_name}.

[OUTPUT FORMAT]
Return ONLY a valid single JSON object without markdown code fences, explanation, or conversational text:
{"grok": "<grok pattern>", "time_field": "<timestamp field name or empty>", "ip_fields": "<comma-separated IP field names or empty>", "mac_fields": "<comma-separated MAC field names or empty>"}

[STRICT RULES]
1. Do NOT return the raw log text. You MUST replace extractable elements with %{PATTERN:field_name}.
2. Always assign meaningful alphanumeric field names (named captures) to extracted fields.
   Common Grok patterns:
   - Timestamps: %{SYSLOGTIMESTAMP:timestamp}, %{TIMESTAMP_ISO8601:timestamp}, %{HTTPDATE:timestamp}
   - IP addresses: %{IPV4:src_ip}, %{IPV6:src_ip}, %{IP:src_ip}, %{IP:client_ip}
   - Host/Domain: %{NOTSPACE:host}, %{HOSTNAME:host}
   - Process/PID: %{WORD:process}\[%{INT:pid}\]
   - Words/Numbers: %{WORD:action}, %{INT:status}, %{NUMBER:bytes}
   - Free text/messages: %{DATA:message}, %{GREEDYDATA:message}
3. time_field, ip_fields, and mac_fields must contain ONLY the FIELD NAMES (e.g. "timestamp", "src_ip") defined in your grok pattern, NEVER actual raw log values.
4. Output ONLY the JSON object without markdown code fences or conversational text.`

	humanContent := fmt.Sprintf("Target Log:\n%s\n\nReturn ONLY the single JSON object containing the Grok pattern.\nJSON:", sampleLine)

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, humanContent),
	}

	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		OutLog("AIGrok GenerateContent err=%v", err)
		res.ErrorMsg = err.Error()
		return res
	}

	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		res.ErrorMsg = "No response from LLM"
		return res
	}

	res = parseAIGrokResponse(resp.Choices[0].Content)
	b.cleanAndValidateAIGrok(&res, sampleLine)

	return res
}

func (b *App) cleanAndValidateAIGrok(res *AIGrokResp, sampleLine string) {
	namedCaptureRe := regexp.MustCompile(`%\{[A-Za-z0-9_]+:([A-Za-z0-9_]+)\}`)
	matches := namedCaptureRe.FindAllStringSubmatch(res.Grok, -1)

	// Check if grok pattern has captures and is syntactically valid and matches sample
	needsFallback := len(matches) == 0
	if !needsFallback {
		grokConfig := grok.Config{
			Patterns:          make(map[string]string),
			NamedCapturesOnly: true,
		}
		grokConfig.Patterns["TWLOGAIAN"] = strings.TrimSpace(res.Grok)
		g, err := grok.NewWithConfig(&grokConfig)
		if err != nil {
			needsFallback = true
		} else if sampleLine != "" {
			values, err := g.Parse("%{TWLOGAIAN}", sampleLine)
			if err != nil || len(values) == 0 {
				OutLog("AIGrok pattern failed to match sample: pat=%s, err=%v", res.Grok, err)
				needsFallback = true
			}
		}
	}

	// Fallback to AutoGrok if LLM output was invalid, empty, or couldn't match sample
	if needsFallback {
		autoRes := b.AutoGrok(sampleLine)
		if autoRes.Grok != "" {
			res.Grok = autoRes.Grok
			res.ErrorMsg = ""
		} else if res.ErrorMsg == "" {
			res.ErrorMsg = "No valid Grok pattern could be generated from log"
		}
	}

	// Parse with final Grok pattern to extract actual captured field names and values
	grokConfig := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	grokConfig.Patterns["TWLOGAIAN"] = strings.TrimSpace(res.Grok)
	g, err := grok.NewWithConfig(&grokConfig)
	if err != nil {
		if res.ErrorMsg == "" {
			res.ErrorMsg = fmt.Sprintf("Invalid Grok pattern: %v", err)
		}
		return
	}

	capturedFieldMap := make(map[string]string)
	if sampleLine != "" {
		if values, err := g.Parse("%{TWLOGAIAN}", sampleLine); err == nil {
			for k, v := range values {
				capturedFieldMap[k] = v
			}
		}
	}

	// Clean / auto-detect TimeField
	if _, ok := capturedFieldMap[res.TimeField]; !ok {
		res.TimeField = ""
		for k := range capturedFieldMap {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "time") || strings.Contains(lk, "date") {
				res.TimeField = k
				break
			}
		}
	}

	// Clean / auto-detect IPFields
	validIPFields := []string{}
	for _, f := range strings.Split(res.IPFields, ",") {
		f = strings.TrimSpace(f)
		if _, ok := capturedFieldMap[f]; ok {
			validIPFields = append(validIPFields, f)
		}
	}
	if len(validIPFields) == 0 {
		for k, v := range capturedFieldMap {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "ip") || regIPNorm.MatchString(v) {
				validIPFields = append(validIPFields, k)
			}
		}
	}
	res.IPFields = strings.Join(validIPFields, ",")

	// Clean / auto-detect MACFields
	validMACFields := []string{}
	for _, f := range strings.Split(res.MACFields, ",") {
		f = strings.TrimSpace(f)
		if _, ok := capturedFieldMap[f]; ok {
			validMACFields = append(validMACFields, f)
		}
	}
	if len(validMACFields) == 0 {
		for k, v := range capturedFieldMap {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "mac") || regMACNorm.MatchString(v) {
				validMACFields = append(validMACFields, k)
			}
		}
	}
	res.MACFields = strings.Join(validMACFields, ",")
}

// AIRegexResp represents the response of GenerateRegex
type AIRegexResp struct {
	Regex    string `json:"Regex"`
	Explain  string `json:"Explain"`
	ErrorMsg string `json:"ErrorMsg"`
}

// GenerateRegex converts a natural language prompt into a regular expression suitable for log search
func (b *App) GenerateRegex(prompt, lang string) AIRegexResp {
	res := AIRegexResp{}
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		res.ErrorMsg = "Prompt is required"
		return res
	}
	cleanedPrompt := cleanSearchPrompt(trimmedPrompt)
	if cleanedPrompt == "" {
		cleanedPrompt = trimmedPrompt
	}

	if b.config.LLMProvider == "" || b.config.LLMProvider == "none" {
		res.ErrorMsg = "LLM provider is not configured"
		return res
	}

	parentCtx := b.ctx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	b.aiMu.Lock()
	b.aiCancel = cancel
	b.aiMu.Unlock()
	defer func() {
		b.aiMu.Lock()
		b.aiCancel = nil
		b.aiMu.Unlock()
	}()

	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("GenerateRegex GetLLM err=%v", err)
		res.ErrorMsg = err.Error()
		return res
	}

	explainLang := "in English"
	if lang == "ja" {
		explainLang = "in Japanese"
	}

	systemPrompt := fmt.Sprintf(`You are a specialized regex generator for log analysis.
Your task is to convert the user's log search request into a Go regular expression (RE2 syntax) and return it ONLY as a JSON object.

[OUTPUT FORMAT]
{"regex": "<pattern>", "explain": "<brief explanation %s>"}

[RULES]
1. Extract search terms directly from the user's request. Do NOT invent unmentioned services (such as sshd, nginx, apache) or extra words.
2. Never include query action words (e.g. "を含む", "を検索", "のログ", "contains", "search for") in the regex. Match only the target terms.
3. Prepend (?i) for case-insensitive matching when searching text or keywords.
4. For multi-word phrases, use \s+ between words (e.g. invalid\s+user).
5. Escape regex meta-characters like dots when matching literal text (e.g. IP addresses).
6. Use only regular expressions valid in Go regexp package (RE2 syntax). Do NOT use lookarounds or backreferences.
7. Output ONLY the single JSON object without markdown code blocks, explanations, or any other conversational text.`, explainLang)

	humanContent := fmt.Sprintf("Search request:\n%s\n\nReturn the Go regular expression in JSON format only.\nJSON:", cleanedPrompt)

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, humanContent),
	}

	resp, err := llm.GenerateContent(ctx, history, llms.WithMaxTokens(128), llms.WithTemperature(0.1))
	if err != nil {
		OutLog("GenerateRegex GenerateContent err=%v", err)
		if ctx.Err() == context.DeadlineExceeded {
			res.ErrorMsg = "AI generation timed out (20s)"
		} else if ctx.Err() == context.Canceled {
			res.ErrorMsg = "AI generation canceled"
		} else {
			res.ErrorMsg = err.Error()
		}
		return res
	}

	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		res.ErrorMsg = "No response from LLM"
		return res
	}

	res = parseAIRegexResponse(resp.Choices[0].Content)
	if res.Regex == "" {
		res.ErrorMsg = "Failed to extract regex from LLM response"
		return res
	}

	res.Regex = cleanGeneratedRegex(res.Regex)

	// Validate regex syntax with Go regexp compiler
	if _, err := regexp.Compile(res.Regex); err != nil {
		OutLog("GenerateRegex invalid regex=%s err=%v", res.Regex, err)
		res.ErrorMsg = fmt.Sprintf("Generated regex syntax error: %v", err)
		return res
	}

	return res
}

// cleanSearchPrompt strips common conversational/query prefixes and suffixes from prompt
func cleanSearchPrompt(prompt string) string {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return ""
	}

	// English prefix stripping (case-insensitive)
	lowerP := strings.ToLower(p)
	prefixes := []string{
		"logs containing ",
		"logs with ",
		"logs of ",
		"search for ",
		"search ",
		"find ",
		"contains ",
		"contain ",
	}
	for _, pref := range prefixes {
		if strings.HasPrefix(lowerP, pref) {
			p = strings.TrimSpace(p[len(pref):])
			lowerP = strings.ToLower(p)
			break
		}
	}

	// Japanese suffix stripping (longer suffixes first)
	suffixes := []string{
		"を含んでいるログ",
		"を含んでいる",
		"を含むログ",
		"を含む",
		"を含まないログ",
		"を含まない",
		"からのアクセスログ",
		"へのアクセスログ",
		"からのアクセス",
		"へのアクセス",
		"のアクセスログ",
		"のエラーログ",
		"のエラー",
		"のログ",
		"を検索",
		"で検索",
		"の検索",
		"を抽出",
		"を探す",
		"を表示",
		"が出ている",
		"が出た",
		"がある",
		"があった",
	}
	for _, suff := range suffixes {
		if strings.HasSuffix(p, suff) && len(p) > len(suff) {
			p = strings.TrimSpace(strings.TrimSuffix(p, suff))
			break
		}
	}

	return p
}

// cleanGeneratedRegex sanitizes and normalizes the regex produced by LLM
func cleanGeneratedRegex(regex string) string {
	r := strings.TrimSpace(regex)
	if r == "" {
		return ""
	}

	actionSuffixes := []string{
		"を含んでいるログ",
		"を含んでいる",
		"を含むログ",
		"を含む",
		"のログ",
		"を検索",
		"で検索",
		"の検索",
		"を抽出",
		"を探す",
	}
	for _, suff := range actionSuffixes {
		if strings.HasSuffix(r, suff) && len(r) > len(suff) {
			r = strings.TrimSpace(strings.TrimSuffix(r, suff))
			break
		}
	}

	return r
}

func parseAIRegexResponse(content string) AIRegexResp {
	res := AIRegexResp{}
	c := strings.TrimSpace(content)

	// Strip markdown code block if present
	if strings.HasPrefix(c, "```") {
		idx := strings.Index(c, "\n")
		if idx != -1 {
			c = c[idx+1:]
		}
		if lastIdx := strings.LastIndex(c, "```"); lastIdx != -1 {
			c = c[:lastIdx]
		}
		c = strings.TrimSpace(c)
	}

	// Try direct JSON unmarshal
	type jsonSchema struct {
		Regex   string `json:"regex"`
		Explain string `json:"explain"`
	}
	var parsed jsonSchema
	if err := json.Unmarshal([]byte(c), &parsed); err == nil && parsed.Regex != "" {
		res.Regex = parsed.Regex
		res.Explain = parsed.Explain
		return res
	}

	// Fallback: search for JSON object inside content
	firstBrace := strings.Index(c, "{")
	lastBrace := strings.LastIndex(c, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		jsonSub := c[firstBrace : lastBrace+1]
		if err := json.Unmarshal([]byte(jsonSub), &parsed); err == nil && parsed.Regex != "" {
			res.Regex = parsed.Regex
			res.Explain = parsed.Explain
			return res
		}
	}

	// Fallback: search lines for backtick quoted regex or compile-able pattern
	lines := strings.Split(c, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Check for backticks: `pattern`
		if strings.Contains(line, "`") {
			parts := strings.Split(line, "`")
			if len(parts) >= 3 {
				candidate := strings.TrimSpace(parts[1])
				if candidate != "" {
					if _, err := regexp.Compile(candidate); err == nil {
						res.Regex = candidate
						return res
					}
				}
			}
		}
		// If line alone compiles as regex and doesn't look like plain sentence
		if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "*") && len(line) < 200 {
			if _, err := regexp.Compile(line); err == nil {
				res.Regex = line
				return res
			}
		}
	}

	res.Regex = c
	return res
}

