package agent

import (
	"WgInspector/entities/agent"
	alerter2 "WgInspector/entities/alerter"
	"WgInspector/entities/config"
	"WgInspector/usecase/agent/analyzer"
	"WgInspector/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/wg00001/wgo-sdk/wg"
	"strings"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/6
 */

type InspLogAnalysis struct {
	Metrics    string `json:"metrics"`
	Belongs    string `json:"belongs"`
	Status     string `json:"status"`
	Trend      string `json:"trend"`
	Problem    string `json:"problem"`
	Suggestion string `json:"suggestion"`
}

func (a *InspLogAnalysis) String() string {
	res, _ := json.Marshal(a)
	return string(res)
}

type AnalysisReport struct {
	InspectLogAnalyze []InspLogAnalysis `json:"inspect_log_analysis"`
	ExtendSuggestion  []string          `json:"extend_suggestion"`
}

func buildAiAlertContent(t *AgentTask, res string) *alerter2.Content {
	return &alerter2.Content{
		Success:   true,
		TimeStamp: time.Now(),
		TaskName:  t.Identity(),
		TaskID:    time.Now().Format("20060504_150201"),
		InspName:  config.Identity{Name: logFilterString(t.LogFilter)},
		Result:    []map[string]any{map[string]any{"Result": res}},
	}
}

func logFilterString(filter config.LogFilter) string {
	marshal, err := json.Marshal(filter)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func generateQueryEmbedding(query string, base agent.KnowledgeBase) ([]float32, error) {
	// 生成嵌入向量（假设有嵌入生成器）
	embedding, err := base.Embedding(query)
	if err != nil {
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}
	return embedding, nil
}

// 使用Ai进行分析获取搜索关键词（并解析JSON
func (t *AgentTask) generateQueryWithAI(logContent *string) (*agent.QueryData, error) {
	kba, err := analyzer.Get(t.KbaseAgentID.Identity())
	if err != nil {
		return nil, err
	}
	res, err := kba.Analyze(&agent.AnalyzeContent{
		SystemMsg: utils.DefaultKBaseSystemMessage,
		UserMsg:   *logContent,
	})
	if err != nil {
		return nil, err
	}
	l := strings.Index(res, "{")
	r := strings.LastIndex(res, "}")
	var q agent.QueryData
	err = json.Unmarshal([]byte(res[l:r+1]), &q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// 混合检索 - 结合关键词和向量检索
// todo: 入参改为Query结构体
func hybridSearch(kbase agent.KnowledgeBase, query string, embedding []float32, topK int) ([]agent.Document, error) {
	// 并行执行两种检索
	keywordResults, _ := kbase.Search(agent.NewQueryData())
	//vectorResults, _ := kbase.Search(topK/2, embedding)
	vectorResults, _ := kbase.Search(agent.NewQueryData())
	// 结果去重合并
	idx := wg.SliceToSet(keywordResults,
		func(document agent.Document) string {
			return document.ID
		})
	for _, vr := range vectorResults {
		if _, ok := idx[vr.ID]; !ok {
			keywordResults = append(keywordResults, vr)
		}
	}
	return keywordResults, nil
}

// 生成知识上下文（辅助函数）
func formatKBaseContent(docs []agent.Document, maxLen int) *string {
	var buf strings.Builder
	currentLen := 0

	for _, doc := range docs {
		content := fmt.Sprintf("条目 %s:\n%s\n\n", doc.ID, doc.Content)
		if currentLen+len(content) > maxLen {
			break
		}
		buf.WriteString(content)
		currentLen += len(content)
	}

	if buf.Len() == 0 {
		return nil
	}
	s := buf.String()
	return &s
}
func (r *AnalysisReport) String() string {
	var buf strings.Builder

	// 报告标题
	buf.WriteString("数据库检查分析报告\n")
	buf.WriteString("========================================\n\n")

	// 检查项明细
	if len(r.InspectLogAnalyze) > 0 {
		buf.WriteString("检查项明细：\n\n")
		for i, item := range r.InspectLogAnalyze {
			// 条目头
			buf.WriteString(fmt.Sprintf("检查项 %d: [%s] @ %s\n", i+1, item.Metrics, item.Belongs))

			// 基本信息
			buf.WriteString(fmt.Sprintf("  状态: %s\n", item.Status))
			buf.WriteString(fmt.Sprintf("  趋势: %s\n", item.Trend))

			// 问题描述（保留原始换行）
			buf.WriteString("  问题描述:\n")
			writeIndented(&buf, item.Problem, 4)

			// 处理建议（保留原始换行）
			buf.WriteString("  处理建议:\n")
			writeIndented(&buf, item.Suggestion, 4)

			// 分隔线（最后一条不显示）
			if i < len(r.InspectLogAnalyze)-1 {
				buf.WriteString("\n" + strings.Repeat("-", 50) + "\n\n")
			}
		}
	} else {
		buf.WriteString("未发现需要处理的检查项\n")
	}

	// 扩展建议
	if len(r.ExtendSuggestion) > 0 {
		buf.WriteString("\n扩展建议：\n\n")
		for i, suggestion := range r.ExtendSuggestion {
			buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(suggestion)))
			if i < len(r.ExtendSuggestion)-1 {
				buf.WriteString("\n")
			}
		}
	}

	return buf.String()
}

// 带缩进的文本写入（保持原始换行）
func writeIndented(buf *strings.Builder, text string, indent int) {
	space := strings.Repeat(" ", indent)
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			buf.WriteString(space + line + "\n")
		}
	}
}
