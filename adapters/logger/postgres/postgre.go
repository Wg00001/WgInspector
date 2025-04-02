package postgres

import (
	"WgInspector/entities/config"
	"WgInspector/entities/logger"
	"WgInspector/usecase/db"
	logger2 "WgInspector/usecase/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/1/19
 */

func init() {
	logger2.RegisterDriver("postgres", LogPostgre{})
}

type LogPostgre struct {
	Config       *config.LogConfig
	LogDBName    config.Identity
	LogTableName config.Identity
}

var _ logger.Logger = (*LogPostgre)(nil)

func (l LogPostgre) Init(cfg *config.LogConfig) (logger.Logger, error) {
	dbName, ok := cfg.Option["dbname"]
	if !ok {
		return LogPostgre{}, fmt.Errorf("Log target db is not exist, dbName:%s\n", dbName)
	}
	tableName, ok := cfg.Option["tablename"]
	if !ok {
		tableName = "inspect_log"
	}
	return LogPostgre{Config: cfg, LogDBName: config.Identity(dbName), LogTableName: config.Identity(tableName)}, nil
}

func (l LogPostgre) GetID() config.Identity {
	return l.Config.Identity
}

func (l LogPostgre) Log(res logger.Content) {
	db.Get(l.LogDBName)
	// 获取数据库连接
	logDB := db.Get(l.LogDBName)
	if logDB.Err != nil {
		log.Printf("Failed to get database connection: %v", logDB.Err)
		return
	}

	if l.LogTableName == "" || l.LogTableName == "inspect_log" {
		l.LogTableName = "inspect_log"
		// 检查表是否存在
		exists, err := checkTableExists(logDB.DB, l.LogTableName)
		if err != nil {
			log.Printf("Failed to check table existence: %v", err)
			return
		}

		// 如果表不存在，则创建表
		if !exists {
			err = createTable(logDB.DB, l.LogTableName)
			if err != nil {
				log.Printf("Failed to create table: %v", err)
				return
			}
		}
	}
	resultContent, err := json.Marshal(res.Result)
	if err != nil {
		log.Printf("log err: json marshal fail - %v", res.Result)
		resultContent = []byte(err.Error())
	}
	// 构建插入 SQL 语句
	insertQuery := fmt.Sprintf(`
        INSERT INTO %s (timestamp, task_name, task_id, inspect_name, db_name, result)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, l.LogTableName)

	// 执行插入操作
	_, err = logDB.DB.Exec(insertQuery, res.Timestamp, res.TaskName, res.TaskID, res.InspName, res.DBName, resultContent)
	if err != nil {
		log.Printf("Failed to insert log data: %v", err)
	}
}

// checkTableExists 检查指定表是否存在
func checkTableExists(db *sql.DB, tableName config.Identity) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE  table_schema = 'public'
        AND    table_name   = $1
    )`
	err := db.QueryRow(query, tableName.Str()).Scan(&exists)
	return exists, err
}

// createTable 创建指定表
func createTable(db *sql.DB, tableName config.Identity) error {
	createQuery := fmt.Sprintf(`
        CREATE TABLE %s (
            id SERIAL PRIMARY KEY,
            timestamp TIMESTAMP,
            task_name TEXT,
            task_id TEXT,
            inspect_name TEXT,
            db_name TEXT,
            result JSONB
        )
    `, tableName.Str())
	_, err := db.Exec(createQuery)
	return err
}

func (l LogPostgre) ReadLog(filter config.LogFilter) ([]logger.Content, error) {
	// 获取数据库连接
	logDB := db.Get(l.LogDBName)
	if logDB == nil {
		return nil, fmt.Errorf("database connection not found: %s", l.LogDBName)
	}

	// 构建动态 WHERE 子句和参数
	var whereClauses []string
	var args []interface{}
	argIdx := 1 // PostgreSQL 参数从 $1 开始

	// 处理时间范围
	if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp BETWEEN $%d AND $%d", argIdx, argIdx+1))
		args = append(args, filter.StartTime, filter.EndTime)
		argIdx += 2
	} else if !filter.StartTime.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, filter.StartTime)
		argIdx++
	} else if !filter.EndTime.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp <= $%d", argIdx))
		args = append(args, filter.EndTime)
		argIdx++
	}

	// 处理 TaskNames 过滤
	if len(filter.TaskNames) > 0 {
		placeholders := make([]string, len(filter.TaskNames))
		for i := range filter.TaskNames {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("task_name IN (%s)", strings.Join(placeholders, ",")))
		args = append(args, interfaceSlice(filter.TaskNames)...)
		argIdx += len(filter.TaskNames)
	}

	// 处理 DBNames 过滤
	if len(filter.DBNames) > 0 {
		placeholders := make([]string, len(filter.DBNames))
		for i := range filter.DBNames {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("db_name IN (%s)", strings.Join(placeholders, ",")))
		args = append(args, interfaceSlice(filter.DBNames)...)
		argIdx += len(filter.DBNames)
	}

	// 处理 TaskIDs 过滤
	if len(filter.TaskIDs) > 0 {
		placeholders := make([]string, len(filter.TaskIDs))
		for i := range filter.TaskIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("task_id IN (%s)", strings.Join(placeholders, ",")))
		args = append(args, interfaceSlice(filter.TaskIDs)...)
		argIdx += len(filter.TaskIDs)
	}

	// 构建完整 SQL
	query := `
        SELECT 
            id,
            COALESCE(timestamp, '1970-01-01'::timestamp),
            COALESCE(task_name, ''),
            COALESCE(task_id, ''),
            COALESCE(inspect_name, ''),
            COALESCE(db_name, ''),
            COALESCE(result::text, '{}')
        FROM public.inspect_log
    `
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 执行查询
	rows, err := logDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	var id int
	// 解析结果
	var contents []logger.Content
	for rows.Next() {
		var content logger.Content
		if err := rows.Scan(
			&id,
			&content.Timestamp,
			&content.TaskName,
			&content.TaskID,
			&content.InspName,
			&content.DBName,
			&content.ResultStr,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return contents, nil
}

// 辅助函数：将任意切片转为 []interface{}
func interfaceSlice[T any](s []T) []interface{} {
	rs := make([]interface{}, len(s))
	for i, v := range s {
		rs[i] = v
	}
	return rs
}
