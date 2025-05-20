package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// Filter 根据表达式过滤Result，返回符合条件的项
func (r Result) Filter(expression string) (Result, error) {
	if len(expression) == 0 {
		return nil, nil
	}
	fieldName, op, compValStr, err := parseExpression(expression)
	if err != nil {
		return nil, err
	}

	compNum, compStr, isNumber := parseComparisonValue(compValStr)

	var filtered Result
	for _, row := range r {
		fieldValue, exists := row[fieldName]
		if !exists {
			continue // 字段不存在则跳过
		}

		match, err := evaluateCondition(fieldValue, op, compNum, compStr, isNumber)
		if err != nil {
			return nil, err
		}
		if match {
			filtered = append(filtered, row)
		}
	}

	return filtered, nil
}

// parseExpression 解析表达式为字段名、操作符和比较值
func parseExpression(expr string) (string, string, string, error) {
	re := regexp.MustCompile(`^\s*(\w+)\s*(==|!=|>=|<=|>|<)\s*(.+?)\s*$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return "", "", "", fmt.Errorf("invalid expression format")
	}
	return matches[1], matches[2], matches[3], nil
}

// parseComparisonValue 解析比较值为数值或字符串
func parseComparisonValue(compVal string) (float64, string, bool) {
	num, err := strconv.ParseFloat(compVal, 64)
	if err == nil {
		return num, "", true
	}
	return 0, compVal, false
}

// evaluateCondition 判断字段值是否符合条件
func evaluateCondition(fieldValue interface{}, op string, compNum float64, compStr string, compIsNumber bool) (bool, error) {
	// 处理字段值为nil的情况
	if fieldValue == nil {
		switch op {
		case "==":
			return compStr == "nil" || compStr == "null", nil
		case "!=":
			return !(compStr == "nil" || compStr == "null"), nil
		default:
			return false, nil // 非等于/不等操作符对nil无效
		}
	}

	var fieldNum float64
	var fieldStr string
	var fieldIsNumber bool
	_ = fieldStr

	switch v := fieldValue.(type) {
	case int:
		fieldNum = float64(v)
		fieldIsNumber = true
	case int64:
		fieldNum = float64(v)
		fieldIsNumber = true
	case float64:
		fieldNum = v
		fieldIsNumber = true
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err == nil {
			fieldNum = num
			fieldIsNumber = true
		} else {
			fieldStr = v
			fieldIsNumber = false
		}
	default:
		fieldStr = fmt.Sprintf("%v", v)
		fieldIsNumber = false
	}

	switch op {
	case ">", "<", ">=", "<=":
		if !compIsNumber || !fieldIsNumber {
			return false, nil
		}
		switch op {
		case ">":
			return fieldNum > compNum, nil
		case "<":
			return fieldNum < compNum, nil
		case ">=":
			return fieldNum >= compNum, nil
		case "<=":
			return fieldNum <= compNum, nil
		}
	case "==", "!=":
		if compIsNumber && fieldIsNumber {
			equal := fieldNum == compNum
			if op == "==" {
				return equal, nil
			} else {
				return !equal, nil
			}
		} else {
			// 处理字符串比较
			var fieldStrVal string
			if fieldIsNumber {
				fieldStrVal = fmt.Sprintf("%f", fieldNum)
				fieldStrVal = strings.TrimRight(fieldStrVal, "0")
				fieldStrVal = strings.TrimRight(fieldStrVal, ".")
			} else {
				fieldStrVal = fmt.Sprintf("%v", fieldValue)
			}

			compStrVal := compStr
			if compIsNumber {
				compStrVal = fmt.Sprintf("%f", compNum)
				compStrVal = strings.TrimRight(compStrVal, "0")
				compStrVal = strings.TrimRight(compStrVal, ".")
			}

			equal := fieldStrVal == compStrVal
			if op == "==" {
				return equal, nil
			} else {
				return !equal, nil
			}
		}
	default:
		return false, fmt.Errorf("unsupported operator: %s", op)
	}

	return false, nil
}
