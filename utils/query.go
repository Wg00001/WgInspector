package utils

import (
	"WgInspector/entities/logger"
	"database/sql"
	"fmt"
	"strings"
)

// PrintQuery 打印查询结果
func PrintQuery(task logger.Content, query *sql.Rows) {
	// 获取列名
	columns, err := query.Columns()
	if err != nil {
		fmt.Printf("Failed to get column names: %v\n", err)
		return
	}

	// 计算每列的最大宽度
	columnWidths := make([]int, len(columns))
	for i, col := range columns {
		if len(col) > columnWidths[i] {
			columnWidths[i] = len(col)
		}
	}

	// 创建用于存储每行数据的切片
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 遍历查询结果，计算每列的最大宽度
	var rows [][]string
	for query.Next() {
		// 扫描当前行的数据
		err = query.Scan(scanArgs...)
		if err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
			return
		}

		// 处理每行数据
		var row []string
		for i, col := range values {
			var value string
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			row = append(row, value)
			if len(value) > columnWidths[i] {
				columnWidths[i] = len(value)
			}
		}
		rows = append(rows, row)
	}

	// 检查遍历过程中是否有错误
	if err = query.Err(); err != nil {
		fmt.Printf("Error during iteration: %v\n", err)
		return
	}

	// 打印表名和分隔线
	fmt.Printf("Result ")
	fmt.Println(strings.Repeat("-", sum(columnWidths)+len(columnWidths)*3-6))

	// 打印列名
	for i, col := range columns {
		fmt.Printf("| %-*s ", columnWidths[i], col)
	}
	fmt.Println("|")
	fmt.Println(strings.Repeat("-", sum(columnWidths)+len(columnWidths)*3+1))

	// 打印每行数据
	for _, row := range rows {
		for i, value := range row {
			fmt.Printf("| %-*s ", columnWidths[i], value)
		}
		fmt.Println("|")
	}
	fmt.Println(strings.Repeat("-", sum(columnWidths)+len(columnWidths)*3+1))
}

// sum 计算切片元素的总和
func sum(nums []int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}
