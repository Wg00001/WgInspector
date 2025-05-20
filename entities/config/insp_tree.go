package config

import (
	"encoding/json"
	"fmt"
)

/**
 * @description: insp的树
 * @author Wg
 * @date 2025/1/19
 */

// InspConfig 分为Insp节点和索引节点，Insp节点也是叶子节点
type InspConfig struct {
	Identity
	SQL       string
	AlertWhen string
	Parent    IdKey   `gorm:"type:jsonb"`
	Children  InspMap `gorm:"-"`
}

type InspMap map[IdKey]*InspConfig

func (m InspMap) MarshalJSON() ([]byte, error) {
	filtered := make(map[string]*InspConfig)
	if m == nil {
		return []byte("{}"), nil
	}
	for idKey, config := range m {
		if config == nil || config.SQL == "" {
			continue
		}
		keyBytes, err := json.Marshal(idKey)
		if err != nil {
			return nil, fmt.Errorf("序列化键失败: %w", err)
		}
		var keyStr Identity
		if err := json.Unmarshal(keyBytes, &keyStr); err != nil {
			return nil, fmt.Errorf("键转换失败: %w", err)
		}
		filtered[keyStr.ToString()] = config
	}
	return json.Marshal(filtered)
}

type InspIndex struct {
	data   InspMap // 存储所有节点的索引
	forest InspMap // 仅包含根节点
}

// 创建新索引
func NewInspIndex(nodes []InspConfig) *InspIndex {
	idx := &InspIndex{
		data:   make(map[IdKey]*InspConfig, len(nodes)),
		forest: make(map[IdKey]*InspConfig),
	}

	// 第一阶段：填充所有节点到data
	for i := range nodes {
		idx.data[nodes[i].IdKey()] = &nodes[i]
	}

	// 第二阶段：建立父子关系并识别根节点
	for _, node := range idx.data {
		if node.Parent.ID == 0 { //父节点进森林
			idx.forest[node.IdKey()] = node
		} else if parent, ok := idx.data[node.Parent]; ok { // 子节点建立父子关系
			if parent.Children == nil {
				parent.Children = make(map[IdKey]*InspConfig)
			}
			parent.Children[node.IdKey()] = node
		}
	}

	return idx
}

// 获取指定节点及其递归子节点（带过滤和去重）
func (idx *InspIndex) Get(baseID Identity, ids ...Identity) []*InspConfig {
	allIDs := append([]Identity{baseID}, ids...)
	var (
		result  []*InspConfig
		visited = make(map[IdKey]struct{}) // 改用IdKey类型去重
		queue   []*InspConfig
	)

	// 初始化队列
	for _, id := range allIDs {
		idKey := id.IdKey()
		node := idx.data[idKey]
		if node == nil {
			continue
		}

		// 处理当前节点自身
		if node.SQL != "" && !exists(visited, idKey) {
			result = append(result, node)
			visited[idKey] = struct{}{}
		}
		queue = append(queue, node)
	}

	// BFS遍历子节点
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, child := range current.Children {
			childKey := child.IdKey()

			// 去重检查
			if exists(visited, childKey) {
				continue
			}
			visited[childKey] = struct{}{}

			// SQL过滤
			if child.SQL != "" {
				result = append(result, child)
			}

			queue = append(queue, child)
		}
	}
	return result
}

func (idx *InspIndex) Put(config InspConfig) error {
	if idx == nil {
		return fmt.Errorf("InspIndex is nil")
	}
	if idx.forest == nil {
		idx.forest = make(map[IdKey]*InspConfig)
	}
	if _, ok := idx.data[config.IdKey()]; ok {
		return fmt.Errorf("InspConfig is already exists")
	}
	if config.Parent.ID != 0 {
		if _, ok := idx.data[config.Parent]; !ok {
			return fmt.Errorf("InspConfig parent not exists")
		}
	}
	idx.data[config.Identity.IdKey()] = &config
	if config.Parent.ID != 0 {
		if idx.data[config.Parent].Children == nil {
			idx.data[config.Parent].Children = make(map[IdKey]*InspConfig)
		}
		idx.data[config.Parent].Children[config.IdKey()] = &config
	}
	return nil
}

func (idx *InspIndex) GetForestRootList() []InspConfig {
	res := make([]InspConfig, 0, len(idx.forest))
	for _, node := range idx.forest {
		res = append(res, *node)
	}
	return res
}

// 辅助函数检查键是否存在
func exists(m map[IdKey]struct{}, key IdKey) bool {
	_, ok := m[key]
	return ok
}
