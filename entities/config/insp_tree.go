package config

import (
	"fmt"
	"sort"
	"strings"
)

/**
 * @description: insp的树
 * @author Wg
 * @date 2025/1/19
 */

// InspNode 分为Insp节点和索引节点，Insp节点也是叶子节点
type InspNode struct {
	Identity
	SQL       string
	Children  Map
	AlertWhen string
	//AlertID   Identity
	//AlertFunc func(alerter.Content) error //包括检查是否符合报警条件，并且发送报警
}

type InspTree struct {
	Roots   Map
	Num     int         //Insp节点数量
	AllInsp []*InspNode //所有的Insp节点
}

func (InspTree) GetIdentity() string {
	return "InspTree"
}

type Map map[string]*InspNode

func NewTree() *InspTree {
	return &InspTree{
		Roots:   make(Map),
		AllInsp: []*InspNode{},
	}
}

// GetNode 获取单个子节点
func (t *InspTree) GetNode(path string) *InspNode {
	if len(path) == 0 {
		return nil
	}
	paths := strings.Split(path, ".")

	var cur *InspNode
	if val, ok := t.Roots[paths[0]]; ok {
		cur = val
	} else {
		return nil
	}

	for i := 1; i < len(paths); i++ {
		if val, ok := cur.Children[paths[i]]; ok {
			cur = val
		} else {
			return nil
		}
	}
	return cur
}

// Arr 获取Map的第一层子节点，会按照字典序排序
func (m Map) Arr() []*InspNode {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	res := make([]*InspNode, 0, len(m))
	for _, k := range keys {
		res = append(res, m[k])
	}
	return res
}

func (n *InspNode) IsInsp() bool {
	return n.SQL != ""
}

// GetAllInsp 获取该节点的所有insp节点，会寻找至叶子节点
func (n *InspNode) GetAllInsp() []*InspNode {
	var res, idx []*InspNode
	idx = n.Children.Arr()
	for len(idx) != 0 {
		var nextIdx []*InspNode
		for _, v := range idx {
			if v.IsInsp() {
				res = append(res, v)
			} else {
				nextIdx = append(nextIdx, v)
			}
		}
		idx = nextIdx
	}
	return res
}

func (n *InspNode) AddChild(node *InspNode) error {
	if n == nil {
		return fmt.Errorf("insp parent is nil")
	}
	if node == nil {
		return fmt.Errorf("new insp child is nil")
	}
	if n.Children == nil {
		n.Children = make(Map)
	}
	n.Children[node.Identity.Str()] = node
	return nil
}

func (t *InspTree) AddChild(path string, node *InspNode) error {
	if node == nil {
		return fmt.Errorf("Insp tree new child node is nil, path: %s\n", path)
	}
	if node.IsInsp() {
		t.AllInsp = append(t.AllInsp, node)
		t.Num++
	}
	//根目录层
	if path == "" {
		t.Roots[node.Identity.Str()] = node
		return nil
	}

	//子目录
	n := t.GetNode(path)
	if n == nil {
		return fmt.Errorf("Insp tree err: path is not exist: %s\n", path)
	}
	if _, ok := n.Children[node.Identity.Str()]; ok {
		return fmt.Errorf("node is already exist, path: %s, name: %s \n", path, node.Identity)
	}
	return n.AddChild(node)
}
