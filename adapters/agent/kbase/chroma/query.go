package chroma

import (
	"WgInspector/entities/agent"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/12
 */

type queryBuilder agent.QueryData

func (b queryBuilder) results() int32 {
	return int32(b.Results)
}

func (b queryBuilder) whereMap() map[string]interface{} {
	var filter map[string]interface{}
	if len(b.MetaData) == 0 {
		filter = nil
	} else if len(b.MetaData) == 1 {
		for k, v := range b.MetaData {
			filter = map[string]interface{}{
				k: v,
			}
		}
	} else {
		or := make([]map[string]interface{}, 0, len(b.MetaData))
		for k, v := range b.MetaData {
			or = append(or, map[string]interface{}{
				k: v,
			})
		}
		filter = map[string]interface{}{
			"$or": or,
		}
	}
	and := []map[string]interface{}{
		{
			"timestamp": map[string]interface{}{
				"$gte": b.MinTime.Unix(),
			},
		},
		{
			"timestamp": map[string]interface{}{
				"$lte": time.Now().Unix(),
			},
		},
	}
	if filter != nil {
		and = append(and, filter)
	}
	return map[string]interface{}{"$and": and}
}
func (b queryBuilder) whereDocMap() map[string]interface{} {
	if len(b.KeyWords) == 1 {
		return map[string]interface{}{
			"$contains": b.KeyWords[0],
		}
	}
	//todo:验证是否正常运行
	or := make([]map[string]interface{}, 0, len(b.KeyWords))
	for _, v := range b.KeyWords {
		or = append(or, map[string]interface{}{
			"$contains": v,
		})
	}
	return map[string]interface{}{
		"$or": or,
	}
}

func (b queryBuilder) text() []string {
	return b.KeyWords
}
