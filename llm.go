package main

import (
	"context"
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
		strings.Contains(s, "err:")
}

func getLogLevelString(l *LogEnt) string {
	if l == nil {
		return "INFO"
	}
	if l.KeyValue != nil {
		if sev, ok := l.KeyValue["severity_str"].(string); ok && sev != "" {
			sevLower := strings.ToLower(sev)
			if isErrorWord(sevLower) {
				return "ERROR"
			}
			if strings.Contains(sevLower, "warn") {
				return "WARN"
			}
		}
		if lv, ok := l.KeyValue["level"].(string); ok && lv != "" {
			lvLower := strings.ToLower(lv)
			if isErrorWord(lvLower) {
				return "ERROR"
			}
			if strings.Contains(lvLower, "warn") {
				return "WARN"
			}
		}
	}
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

func buildAIPromptJA(req AIAnalysisReportRequest, summary AIReportSummary, samples []string) string {
	var sb strings.Builder
	sb.WriteString("あなたは高度なサイバーセキュリティおよびログ分析の専門家です。\n")
	sb.WriteString("提示されたログデータを詳細に分析し、運用管理者が迅速に対処できる実践的な日本語の分析レポート（Markdown形式）を作成してください。\n\n")

	if req.Prompt != "" {
		sb.WriteString("【特別な指示・分析リクエスト】\n")
		sb.WriteString(req.Prompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("【ログ概要データ】\n")
	sb.WriteString(fmt.Sprintf("- 総ログ件数: %d 件\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- エラーログ件数: %d 件\n", summary.ErrorCount))
	sb.WriteString(fmt.Sprintf("- 警告ログ件数: %d 件\n", summary.WarningCount))
	if summary.StartTime > 0 && summary.EndTime > 0 {
		st := time.Unix(0, summary.StartTime).Format("2006-01-02 15:04:05")
		et := time.Unix(0, summary.EndTime).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("- 対象期間: %s 〜 %s\n", st, et))
	}

	if len(summary.TopErrors) > 0 {
		sb.WriteString("\n【頻出エラーパターン（上位）】\n")
		for i, ep := range summary.TopErrors {
			sb.WriteString(fmt.Sprintf("%d. 発生回数: %d 回\n   パターン: `%s`\n   具体例: `%s`\n", i+1, ep.Count, ep.Pattern, ep.Example))
		}
	}

	if len(samples) > 0 {
		sb.WriteString("\n【ログサンプル（最新・代表的な抜粋）】\n```\n")
		for _, s := range samples {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString(`【レポート構成要件】
以下の構成で、見出し、箇条書き、太字、コードブロックなどを効果的に用いてMarkdownで記述してください:
1. **エグゼクティブ・サマリー**: ログ全体の傾向、健全性評価、特に警戒すべき事象の要約。
2. **検出された異常・リスク（Anomalies）**:
   - 種類（エラー急増 / セキュリティ侵害・不正アクセス / パフォーマンス・障害 / その他）
   - 重要度（Critical / High / Medium / Low）
   - 具体的な兆候と該当ログ
3. **根本原因の考察（Root Cause Analysis）**: なぜこのエラーや異常が発生しているのかの分析。
4. **推奨アクション・再発防止策（Recommendations）**:
   - 即時対応（緊急対応）
   - 恒久対策（監視強化・設定変更・保守手順など）

必ず日本語で回答してください。`)

	return sb.String()
}

func buildAIPromptEN(req AIAnalysisReportRequest, summary AIReportSummary, samples []string) string {
	var sb strings.Builder
	sb.WriteString("You are an expert in cybersecurity, systems administration, and log analysis.\n")
	sb.WriteString("Analyze the provided log dataset in detail and produce a structured, actionable analysis report in Markdown format.\n\n")

	if req.Prompt != "" {
		sb.WriteString("【Custom Instruction / Request】\n")
		sb.WriteString(req.Prompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("【Log Summary Data】\n")
	sb.WriteString(fmt.Sprintf("- Total Log Count: %d\n", summary.TotalEntries))
	sb.WriteString(fmt.Sprintf("- Error Count: %d\n", summary.ErrorCount))
	sb.WriteString(fmt.Sprintf("- Warning Count: %d\n", summary.WarningCount))
	if summary.StartTime > 0 && summary.EndTime > 0 {
		st := time.Unix(0, summary.StartTime).Format("2006-01-02 15:04:05")
		et := time.Unix(0, summary.EndTime).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("- Time Range: %s to %s\n", st, et))
	}

	if len(summary.TopErrors) > 0 {
		sb.WriteString("\n【Top Error Patterns】\n")
		for i, ep := range summary.TopErrors {
			sb.WriteString(fmt.Sprintf("%d. Count: %d occurrences\n   Pattern: `%s`\n   Example: `%s`\n", i+1, ep.Count, ep.Pattern, ep.Example))
		}
	}

	if len(samples) > 0 {
		sb.WriteString("\n【Recent Log Sample】\n```\n")
		for _, s := range samples {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString(`【Report Requirements】
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

	return sb.String()
}

// AnalyzeLogsWithAI aggregates statistics, extracts top error patterns, and streams AI analysis.
func (b *App) AnalyzeLogsWithAI(req AIAnalysisReportRequest, logs []*LogEnt) AIAnalysisReportResponse {
	if len(logs) == 0 {
		return AIAnalysisReportResponse{
			Error: "No logs to analyze",
		}
	}

	if req.Lang == "" {
		req.Lang = "ja"
	}
	if req.TopNError <= 0 {
		req.TopNError = 10
	}
	if req.SampleSize <= 0 {
		req.SampleSize = 50
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

	ctx := b.ctx
	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("AnalyzeLogsWithAI GetLLM err=%v", err)
		return AIAnalysisReportResponse{
			Summary: summary,
			Error:   err.Error(),
		}
	}

	var promptText string
	if req.Lang == "ja" {
		promptText = buildAIPromptJA(req, summary, samples)
	} else {
		promptText = buildAIPromptEN(req, summary, samples)
	}

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, promptText),
	}

	var sb strings.Builder
	var tokenCount int
	st := time.Now()
	decoder := &utf8StreamDecoder{}

	resp, err := llm.GenerateContent(ctx, history, llms.WithStreamingFunc(func(c context.Context, chunk []byte) error {
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
		sb.WriteString(tail)
		if b.ctx != nil {
			wails.EventsEmit(b.ctx, "ai_report_stream", tail)
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
