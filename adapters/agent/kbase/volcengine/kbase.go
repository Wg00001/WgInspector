package volcengine

import (
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/volcengine/volc-sdk-golang/base"
)

type KBaseVolcengine struct {
	Config config.KnowledgeBaseConfig
	AK     string
	SK     string
}

func (k *KBaseVolcengine) Init(config config.KnowledgeBaseConfig) (agent.KnowledgeBase, error) {
	k.Config = config
	var ok bool
	k.AK, ok = config.Option["ak"]
	if !ok {
		return nil, fmt.Errorf("missing ak in config options")
	}
	k.SK, ok = config.Option["sk"]
	if !ok {
		return nil, fmt.Errorf("missing sk in config options")
	}
	return k, nil
}

func (k *KBaseVolcengine) WriteIn(docs []agent.Document) error {
	for _, doc := range docs {
		collectionName, ok := k.Config.Option["collection_name"]
		if !ok {
			return fmt.Errorf("collection_name is required in config options")
		}
		project := k.Config.Option["project"]

		addType, ok := doc.Metadata["add_type"].(string)
		if !ok {
			return fmt.Errorf("add_type is required in document metadata")
		}
		docName, ok := doc.Metadata["doc_name"].(string)
		if !ok {
			return fmt.Errorf("doc_name is required in document metadata")
		}
		docType, ok := doc.Metadata["doc_type"].(string)
		if !ok {
			return fmt.Errorf("doc_type is required in document metadata")
		}

		var urlStr, content string
		switch addType {
		case "url":
			if val, ok := doc.Metadata["url"].(string); ok {
				urlStr = val
			} else {
				urlStr = doc.Content
			}
		case "text":
			content = doc.Content
		default:
			return fmt.Errorf("unsupported add_type: %s", addType)
		}

		meta := make([]map[string]interface{}, 0)
		for key, value := range doc.Metadata {
			if isReservedField(key) {
				continue
			}

			fieldType := inferFieldType(value)
			meta = append(meta, map[string]interface{}{
				"field_name":  key,
				"field_type":  fieldType,
				"field_value": value,
			})
		}

		requestBody := map[string]interface{}{
			"collection_name": collectionName,
			"project":         project,
			"add_type":        addType,
			"doc_id":          doc.ID,
			"doc_name":        docName,
			"doc_type":        docType,
			"meta":            meta,
		}

		if addType == "url" {
			requestBody["url"] = urlStr
		} else {
			requestBody["content"] = content
		}

		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request failed: %v", err)
		}

		req, err := k.prepareRequest("POST", "/api/knowledge/doc/add", bodyBytes)
		if err != nil {
			return fmt.Errorf("prepare request failed: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("API request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		var result struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("decode response failed: %v", err)
		}
		if result.Code != 0 {
			return fmt.Errorf("API error: %s", result.Message)
		}
	}
	return nil
}

func (k *KBaseVolcengine) Search(queries agent.QueryData) ([]agent.Document, error) {
	collectionName, ok := k.Config.Option["collection_name"]
	if !ok {
		return nil, fmt.Errorf("collection_name is required in config options")
	}

	queryParam := make(map[string]interface{})
	for k, v := range queries.MetaData {
		queryParam[k] = v
	}

	requestBody := map[string]interface{}{
		"name":         collectionName,
		"query":        strings.Join(queries.KeyWords, " "),
		"limit":        queries.Results,
		"query_param":  queryParam,
		"dense_weight": 0.5,
		"pre_processing": map[string]interface{}{
			"need_instruction":   true,
			"rewrite":            true,
			"messages":           []interface{}{},
			"return_token_usage": true,
		},
		"post_processing": map[string]interface{}{
			"rerank_switch":       false,
			"rerank_model":        "m3-v2-rerank",
			"rerank_only_chunk":   false,
			"retrieve_count":      25,
			"endpoint_id":         "ep",
			"chunk_group":         false,
			"get_attachment_link": false,
		},
	}

	if dw, ok := k.Config.Option["dense_weight"]; ok {
		if denseWeight, err := strconv.ParseFloat(dw, 64); err == nil {
			requestBody["dense_weight"] = denseWeight
		}
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	req, err := k.prepareRequest("POST", "/api/knowledge/collection/search_knowledge", bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("prepare request failed: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Documents []struct {
				DocID   string                   `json:"doc_id"`
				Content string                   `json:"content"`
				Meta    []map[string]interface{} `json:"meta"`
				Score   float64                  `json:"score"`
			} `json:"documents"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	documents := make([]agent.Document, 0, len(response.Data.Documents))
	for _, doc := range response.Data.Documents {
		metadata := make(map[string]interface{})
		for _, metaItem := range doc.Meta {
			if fieldName, ok := metaItem["field_name"].(string); ok {
				metadata[fieldName] = metaItem["field_value"]
			}
		}

		documents = append(documents, agent.Document{
			ID:        doc.DocID,
			Content:   doc.Content,
			Metadata:  metadata,
			Embedding: nil,
		})
	}

	return documents, nil
}

func (k *KBaseVolcengine) prepareRequest(method, path string, body []byte) (*http.Request, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api-knowledgebase.mlp.cn-beijing.volces.com",
		Path:   path,
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", u.Host)

	credential := base.Credentials{
		AccessKeyID:     k.AK,
		SecretAccessKey: k.SK,
		Service:         "air",
		Region:          "cn-north-1",
	}

	return credential.Sign(req), nil
}

func isReservedField(field string) bool {
	reserved := map[string]bool{
		"add_type": true,
		"doc_name": true,
		"doc_type": true,
		"url":      true,
		"content":  true,
	}
	return reserved[field]
}

func inferFieldType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "bool"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "number"
	case string:
		return "string"
	default:
		return "string"
	}
}
func (k *KBaseVolcengine) Embedding(query string) ([]float32, error) {
	//TODO implement me
	panic("implement me")
}

var _ agent.KnowledgeBase = (*KBaseVolcengine)(nil)
