package chroma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/amikos-tech/chroma-go/types"
	"io"
	"net/http"
	"net/url"
)

func NewWithModel(ak string) FeishuEmbedding {
	return FeishuEmbedding{
		AK: ak,
	}
}

type FeishuEmbedding struct {
	AK string
}

// 共享的 API 请求结构
type embeddingRequest struct {
	EncodingFormat string   `json:"encoding_format"`
	Input          []string `json:"input"`
	Model          string   `json:"model"`
}

// 共享的 API 响应结构
type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (f FeishuEmbedding) EmbedDocuments(ctx context.Context, texts []string) ([]*types.Embedding, error) {
	return f.embedBatch(ctx, texts)
}

func (f FeishuEmbedding) EmbedQuery(ctx context.Context, text string) (*types.Embedding, error) {
	embeddings, err := f.embedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// 核心嵌入处理方法
func (f FeishuEmbedding) embedBatch(ctx context.Context, texts []string) ([]*types.Embedding, error) {
	// 1. 构建请求体
	reqBody := embeddingRequest{
		EncodingFormat: "float",
		Input:          texts,
		Model:          "doubao-embedding-text-240715", // 根据实际情况调整
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 2. 创建签名请求
	req, err := f.prepareRequest("/api/v3/embeddings", jsonBody)
	if err != nil {
		return nil, fmt.Errorf("prepare request failed: %w", err)
	}
	req = req.WithContext(ctx)

	// 3. 发送 HTTP 请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. 处理响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
	}

	// 5. 解析响应
	var apiResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// 6. 错误处理
	if apiResp.Error.Message != "" {
		return nil, fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	// 7. 转换为目标结构
	result := make([]*types.Embedding, len(apiResp.Data))
	for i, item := range apiResp.Data {
		vector := make([]float32, len(item.Embedding))
		for j, v := range item.Embedding {
			vector[j] = float32(v)
		}
		result[i] = &types.Embedding{
			ArrayOfFloat32: &vector,
			ArrayOfInt32:   nil, // 明确设置为 nil
		}
	}

	return result, nil
}

// 签名方法（已提供）
func (f FeishuEmbedding) prepareRequest(path string, body []byte) (*http.Request, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "ark.cn-beijing.volces.com",
		Path:   path,
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", f.AK)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", u.Host)

	return req, nil
}

func (f FeishuEmbedding) EmbedRecords(ctx context.Context, records []*types.Record, force bool) error {
	//TODO implement me
	panic("implement me")
}
