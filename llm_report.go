package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// LLMExplainReportRequest contains parameters for explaining a report using LLM.
type LLMExplainReportRequest struct {
	ReportType string            `json:"ReportType"` // "ranking", "time", "time3d", "cluster", "histogram", "fft", "world", "globe", "graph", "heatmap", "interval", "delay", "email", "memo", "result"
	Title      string            `json:"Title"`      // Title of the report
	Field      string            `json:"Field"`      // Primary field name
	SubField   string            `json:"SubField"`   // Secondary field name (if applicable)
	Params     map[string]string `json:"Params"`     // Additional parameters (mode, unit, etc.)
	Columns    []string          `json:"Columns"`    // Table header columns
	Data       [][]interface{}   `json:"Data"`       // Data rows (clipped to max 100 rows)
	Summary    string            `json:"Summary"`    // Optional pre-aggregated summary text
	Lang       string            `json:"Lang"`       // "ja" or "en"
}

// LLMExplainReport generates an AI explanation for any report screen based on displayed data.
func (b *App) LLMExplainReport(req LLMExplainReportRequest) AIAnswer {
	r := AIAnswer{}
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	llm, err := b.GetLLM(ctx)
	if err != nil {
		OutLog("LLMExplainReport GetLLM err=%v", err)
		r.Error = err.Error()
		return r
	}

	targetLang := "Japanese"
	if req.Lang == "en" {
		targetLang = "English"
	}

	systemPrompt, userPrompt := buildReportPrompt(req, targetLang)

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		OutLog("LLMExplainReport GenerateContent err=%v", err)
		r.Error = err.Error()
		return r
	}

	if resp != nil && len(resp.Choices) > 0 && resp.Choices[0].Content != "" {
		r.Answer = strings.TrimSpace(resp.Choices[0].Content)
	} else {
		r.Error = "no response from LLM"
	}
	return r
}

func buildReportPrompt(req LLMExplainReportRequest, targetLang string) (string, string) {
	// Base instructions are written in clear English to best accommodate smaller / local LLMs.
	var system strings.Builder
	system.WriteString("You are an expert in cybersecurity, system administration, and log data analytics. ")
	system.WriteString("Analyze the provided report dataset in detail and provide a clear, professional, well-structured analytical explanation in Markdown format. ")
	system.WriteString(fmt.Sprintf("You MUST write your entire response, analysis, headings, bullet points, and recommendations strictly in %s.\n\n", targetLang))

	system.WriteString("Report Domain Focus:\n")
	switch strings.ToLower(req.ReportType) {
	case "ranking":
		system.WriteString("Analyze item frequencies, concentration ratios, outliers, and anomalous top-ranking entities. Identify whether heavy hitters indicate normal behavior, operational issues, or potential security threats.")
	case "histogram":
		system.WriteString("Analyze value distributions, central tendencies, spreads, skewness, and extreme outliers. Identify potential latency bottlenecks, data surges, or abnormal metric spikes.")
	case "time":
		system.WriteString("Analyze time-series trends, cyclical variations, activity peaks, surge timeframes, and unusual off-hour activities.")
	case "time3d":
		system.WriteString("Analyze multi-dimensional distributions across time and attributes. Highlight clusters of activity, rare coordinate occurrences, and temporal anomalies.")
	case "cluster":
		system.WriteString("Analyze log cluster patterns, volume distribution among clusters, and specifically examine rare/singleton clusters as potential security anomalies or system bugs.")
	case "fft":
		system.WriteString("Analyze frequency spectrum characteristics and periodicity. Identify periodic activities such as scheduled cron jobs, automated bot polling, or C2 beaconing signals.")
	case "world", "globe":
		system.WriteString("Analyze geographic IP distribution, regional access concentrations, unexpected foreign connections, and accesses from high-risk locations.")
	case "graph":
		system.WriteString("Analyze network topology, relationship graphs, high-degree hub nodes, isolated clusters, and indicators of lateral movement or unauthorized connectivity.")
	case "heatmap":
		system.WriteString("Analyze day-of-week and hourly density matrices. Highlight periods of intense activity, unusual off-hours or weekend operations, and regularity.")
	case "interval":
		system.WriteString("Analyze event intervals and deltas. Detect beaconing (consistent intervals), automated bursts, or abnormal inter-arrival times.")
	case "delay":
		system.WriteString("Analyze response time and latency statistics. Identify performance degradation, tail latency, timeout risks, and root causes.")
	case "email":
		system.WriteString("Analyze email transaction logs, sender/recipient volumes, delivery failures, SPF/DKIM issues, and potential spam or data exfiltration.")
	case "sigma":
		system.WriteString("Analyze detected security threats and compliance audit findings identified by Sigma and Wazuh rules. Assess incident severity, MITRE ATT&CK tactics/techniques, compliance impacts (PCI-DSS, NIST, GDPR, CIS), potential attack vectors, affected assets, and provide clear, prioritized incident response and remediation steps.")
	case "result":
		system.WriteString("Provide a comprehensive summary of search results, key observations, data distribution, and recommended next steps for log investigation.")
	case "memo":
		system.WriteString("Review saved analytical metrics and dashboard memos, providing a consolidated analytical overview and actionable maintenance advice.")
	default:
		system.WriteString("Analyze the provided tabular summary data, highlight key trends, anomalies, and operational recommendations.")
	}

	var user strings.Builder
	user.WriteString(fmt.Sprintf("# Report Analysis Request: %s\n\n", req.Title))
	user.WriteString(fmt.Sprintf("- Report Type: %s\n", req.ReportType))
	if req.Field != "" {
		user.WriteString(fmt.Sprintf("- Primary Field: %s\n", req.Field))
	}
	if req.SubField != "" {
		user.WriteString(fmt.Sprintf("- Secondary Field: %s\n", req.SubField))
	}
	for k, v := range req.Params {
		user.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	if req.Summary != "" {
		user.WriteString(fmt.Sprintf("- Summary: %s\n", req.Summary))
	}
	user.WriteString("\n")

	// Render table data into Markdown format
	if len(req.Columns) > 0 && len(req.Data) > 0 {
		user.WriteString("## Displayed Data (Sample / Top Entries):\n\n")
		// Header row
		user.WriteString("| " + strings.Join(req.Columns, " | ") + " |\n")
		// Separator row
		seps := make([]string, len(req.Columns))
		for i := range seps {
			seps[i] = "---"
		}
		user.WriteString("| " + strings.Join(seps, " | ") + " |\n")

		maxRows := 60
		if len(req.Data) < maxRows {
			maxRows = len(req.Data)
		}
		for i := 0; i < maxRows; i++ {
			row := req.Data[i]
			cells := make([]string, len(req.Columns))
			for colIdx := range req.Columns {
				if colIdx < len(row) {
					cells[colIdx] = fmt.Sprintf("%v", row[colIdx])
				} else {
					cells[colIdx] = ""
				}
			}
			user.WriteString("| " + strings.Join(cells, " | ") + " |\n")
		}
		if len(req.Data) > maxRows {
			user.WriteString(fmt.Sprintf("\n*(Total entries: %d, showing top %d)*\n", len(req.Data), maxRows))
		}
		user.WriteString("\n")
	}

	user.WriteString(fmt.Sprintf(`Please structure your analysis using Markdown with the following sections (written in %s):
1. **Overview**: Executive summary of what this report data represents.
2. **Key Findings & Trends**: Notable observations, patterns, and high-frequency elements.
3. **Anomalies & Security/Operational Risks**: Suspicious indicators, outliers, bottlenecks, or potential failures.
4. **Actionable Recommendations**: Clear, prioritized next steps or mitigation measures for administrators.

You MUST write the entire response in %s.`, targetLang, targetLang))

	return system.String(), user.String()
}
