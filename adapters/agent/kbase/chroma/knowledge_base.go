package chroma

import (
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	"WgInspector/usecase/agent/kbase"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"context"
	"fmt"
	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/amikos-tech/chroma-go/pkg/embeddings/ollama"
	"github.com/amikos-tech/chroma-go/types"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/5
 */

func init() {
	kbase.RegisterDriver("chroma", KBaseChroma{})
}

type KBaseChroma struct {
	Config           *config.KnowledgeBaseConfig
	Path             string
	Collection       string //chroma的collection类似于库
	Tenant           string //chroma需要指定租户
	Database         string
	EmbeddingDriver  string
	EmbeddingBaseUrl string
	EmbeddingModel   string
	EmbeddingApikey  string
	Efunc            types.EmbeddingFunction //进行向量计算的函数
}

var _ agent.KnowledgeBase = (*KBaseChroma)(nil)

func (k KBaseChroma) Init(cfg *config.KnowledgeBaseConfig) (_ agent.KnowledgeBase, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("kbase: chroma init fail - panic: %s\n", r)
		}
	}()
	k.Config = cfg
	value := utils.UseMap(cfg.Value)
	k.Path = value.GetString("path")
	k.Collection = value.GetString("collection")
	k.Tenant = value.GetString("tenant")
	k.Database = value.GetString("database")
	// 获取 embedding 子 map
	embedding := value.GetMap("embedding")
	k.EmbeddingBaseUrl = embedding.GetString("baseurl")
	k.EmbeddingModel = embedding.GetString("model")
	k.EmbeddingApikey = embedding.GetString("apikey")
	k.EmbeddingDriver = embedding.GetString("driver")

	switch k.EmbeddingDriver {
	case "ollama":
		k.Efunc, err = ollama.NewOllamaEmbeddingFunction(
			ollama.WithBaseURL(k.EmbeddingBaseUrl),
			ollama.WithModel(k.EmbeddingModel))
		if err != nil {
			return
		}
	case "openai":
	default:
		//agentConfig := config2.GetAgentConfig()
		agentConfig, _ := config2.Get[config.AgentConfig](config.AgentConfig{})
		k.Efunc, err = openai.NewOpenAIEmbeddingFunction(
			agentConfig.ApiKey,
			func(c *openai.OpenAIClient) error {
				c.Model = agentConfig.Model
				return nil
			},
			func(c *openai.OpenAIClient) error {
				c.BaseURL = agentConfig.Url
				return nil
			})
		if err != nil {
			return k, fmt.Errorf("agent - kbase: chroma Error creating OpenAI embedding function: %v\n", err)
		}
	}

	return k, nil
}

func (k KBaseChroma) WriteIn(docs []agent.Document) error {
	if docs == nil || len(docs) == 0 {
		return fmt.Errorf("agent - kbase: chroma write in fail: can't write nil document")
	}
	ctx := context.Background()
	collection, err := k.connect(ctx)
	if err != nil {
		return err
	}

	metaData := make([]map[string]interface{}, 0, len(docs))
	document := make([]string, 0, len(docs))
	ids := make([]string, 0, len(docs))
	ems := make([]*types.Embedding, 0, len(docs))
	for _, d := range docs {
		metaData = append(metaData, d.Metadata)
		document = append(document, d.Content)
		ids = append(ids, d.ID)
		if d.Embedding != nil {
			ems = append(ems, &types.Embedding{ArrayOfFloat32: &d.Embedding})
		}
	}

	_, err = collection.Add(ctx, ems, metaData, document, ids)
	if err != nil {
		return fmt.Errorf("agent - kbase: chroma Error adding documents: %v\n", err)
	}
	return nil
}

func (k KBaseChroma) Search(queries agent.QueryData) ([]agent.Document, error) {
	ctx := context.Background()

	collection, err := k.connect(ctx)
	if err != nil {
		return nil, err
	}
	query := queryBuilder(queries)

	//todo：将入参转成where metadata和where document格式
	results, err := collection.QueryWithOptions(
		ctx,
		types.WithQueryTexts(query.text()),
		types.WithNResults(query.results()),
		types.WithWhereMap(query.whereMap()),
		types.WithWhereDocumentMap(query.whereDocMap()),
		types.WithInclude(types.IDocuments, types.IMetadatas),
	)
	if err != nil {
		return nil, err
	}
	return parseQueryResults(results), nil
}

func (k KBaseChroma) SimilaritySearch(topK int, embedding []float32) ([]agent.Document, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	ctx := context.Background()
	collection, err := k.connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// 构建嵌入查询向量
	queryEmb := &types.Embedding{
		ArrayOfFloat32: &embedding,
	}

	results, err := collection.QueryWithOptions(
		ctx,
		types.WithQueryEmbeddings([]*types.Embedding{queryEmb}),
		types.WithNResults(int32(topK)),
		types.WithInclude(types.IDocuments, types.IMetadatas),
	)
	if err != nil {
		return nil, fmt.Errorf("similarity query failed: %w", err)
	}

	return parseQueryResults(results), nil
}

func (k KBaseChroma) Embedding(query string) ([]float32, error) {
	embedQuery, err := k.Efunc.EmbedQuery(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	return *embedQuery.ArrayOfFloat32, nil
}

func (k KBaseChroma) connect(ctx context.Context) (*chromago.Collection, error) {
	var cliopt []chromago.ClientOption
	if k.Tenant != "" {
		cliopt = append(cliopt, chromago.WithTenant(k.Tenant))
	}
	if k.Database != "" {
		cliopt = append(cliopt, chromago.WithDatabase(k.Database))
	}
	client, err := chromago.NewClient(k.Path, cliopt...)
	if err != nil {
		return nil, fmt.Errorf("agent - kbase: chroma Failed to create client: %v\n", err)
	}
	collection, err := client.CreateCollection(ctx, k.Collection, map[string]interface{}{}, true, k.Efunc, types.L2)
	if err != nil {
		return nil, fmt.Errorf("agent - kbase: chroma Failed to create or get collection: \n    %v\n", err)
	}

	return collection, nil
}

func parseQueryResults(results *chromago.QueryResults) []agent.Document {
	var docs []agent.Document
	if results == nil {
		return docs
	}

	for qIdx := range results.Ids {
		for i := 0; i < len(results.Ids[qIdx]); i++ {
			// 安全索引检查
			if i >= len(results.Documents[qIdx]) ||
				i >= len(results.Metadatas[qIdx]) {
				continue
			}

			doc := agent.Document{
				ID:       results.Ids[qIdx][i],
				Content:  results.Documents[qIdx][i],
				Metadata: results.Metadatas[qIdx][i],
			}
			docs = append(docs, doc)
		}
	}
	return docs
}
