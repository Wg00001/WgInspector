package agent

import (
	"WgInspector/entities/config"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/5
 */

type Document struct {
	ID        string
	Content   string
	Embedding []float32
	Metadata  map[string]interface{}
}

type KnowledgeBase interface {
	Init(config *config.KnowledgeBaseConfig) (KnowledgeBase, error)
	WriteIn(docs []Document) error //将文档写入知识库
	Search(queries QueryData) ([]Document, error)
	Embedding(query string) ([]float32, error)
	//SimilaritySearch(topK int, embedding []float32) ([]*Document, error)
}

type QueryData struct {
	Results  int //number of results
	MinTime  time.Time
	KeyWords []string
	MetaData map[string]string
}

func NewQueryData() QueryData {
	return QueryData{Results: 3, MinTime: time.Now().AddDate(0, 0, -15)}
}

func (d QueryData) WithKeyWords(keywords ...string) QueryData {
	d.KeyWords = keywords
	return d
}

func (d QueryData) WithMetaData(metadata map[string]string) QueryData {
	d.MetaData = metadata
	return d
}
