package utils

import "strconv"

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/4
 */

type Map map[string]interface{}

func UseMap(origin map[string]interface{}) Map {
	return origin
}

func (m Map) GetInt(key string, keys ...string) int {
	s := m.GetString(key, keys...)
	atoi, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return atoi
}

func (m Map) GetString(key string, keys ...string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	// 直接处理无子键的情况
	if len(keys) == 0 {
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}
	// 处理嵌套 map
	child, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	for _, k := range keys[:len(keys)-1] { // 遍历到倒数第二个键
		val, ok := child[k]
		if !ok {
			return ""
		}
		child, ok = val.(map[string]interface{})
		if !ok {
			return ""
		}
	}
	// 处理最后一个键
	if s, ok := child[keys[len(keys)-1]].(string); ok {
		return s
	}
	return ""
}

func (m Map) GetMap(key string) Map {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	if child, ok := v.(map[string]interface{}); ok {
		return child
	}
	return make(Map)
}
