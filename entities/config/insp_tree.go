package config

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
	Parent    Identity                 `gorm:"type:jsonb"`
	Children  map[Identity]*InspConfig `gorm:"-"`
	//AlertID   Identity
	//AlertFunc func(alerter.Content) error //包括检查是否符合报警条件，并且发送报警
}

func (InspIndex) GetIdentity() Identity {
	return Identity{
		ID:   0,
		Name: "insp_tree",
	}
}

type InspIndex map[Identity]*InspConfig

func NewInspIndex(nodes []InspConfig) InspIndex {
	idx := make(InspIndex, len(nodes))
	for i := range nodes {
		idx[nodes[i].Identity] = &nodes[i]
	}

	// 遍历所有节点，建立父子关系
	for i := range nodes {
		if nodes[i].Parent != (Identity{}) {
			// 查找父节点
			if parent, exists := idx[nodes[i].Parent]; exists {
				if parent.Children == nil {
					parent.Children = make(map[Identity]*InspConfig)
				}
				parent.Children[nodes[i].Identity] = &nodes[i]
			}
		}
	}

	return idx
}

func (idx InspIndex) Get(baseID Identity, ids ...Identity) []*InspConfig {
	// 合并所有查询ID（包括baseID和可变参数ids）
	allIDs := append([]Identity{baseID}, ids...)

	var (
		result  []*InspConfig                 // 最终结果
		visited = make(map[Identity]struct{}) // 全局去重记录
		queue   []*InspConfig                 // BFS队列
	)

	// 初始化队列：添加所有存在的目标节点
	for _, id := range allIDs {
		if node, exists := idx[id]; exists {
			// 如果节点未访问过且SQL有效，直接加入结果（父节点自身）
			if _, seen := visited[id]; !seen && node.SQL != "" {
				result = append(result, node)
				visited[id] = struct{}{}
			}
			queue = append(queue, node)
		}
	}

	// BFS遍历
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, child := range current.Children {
			// 去重检查
			if _, exists := visited[child.Identity]; exists {
				continue
			}
			visited[child.Identity] = struct{}{} // 标记已访问

			// 过滤有效节点
			if child.SQL != "" {
				result = append(result, child)
			}

			// 加入队列继续遍历
			queue = append(queue, child)
		}
	}

	return result
}
