package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/15
 */

type Result []map[string]interface{}

// RowsToResult 将 sql.Rows 转换为 []map[string]interface{} 结构,自动处理类型转换
func RowsToResult(rows *sql.Rows) (Result, error) {
	// 获取列名集合
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}

	for rows.Next() {
		scanArgs := make([]interface{}, len(columns))
		for i := range scanArgs {
			scanArgs[i] = new(interface{})
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rawValue := *(scanArgs[i].(*interface{}))

			switch v := rawValue.(type) {
			case []byte:
				// 二进制数据转字符串
				rowMap[col] = string(v)
			case time.Time:
				// 时间类型标准化处理
				rowMap[col] = v.Format(time.RFC3339Nano)
			case nil:
				// 空值显式处理
				rowMap[col] = nil
			default:
				// 保留原始类型（int64/float64/string等）
				rowMap[col] = v
			}
		}

		result = append(result, rowMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r Result) MarshallJSON() []byte {
	marshal, err := json.Marshal(r)
	if err != nil {
		return []byte(fmt.Sprintf("Json marshall fail: %s", err))
	}
	return marshal
}
