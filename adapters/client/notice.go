package client

import (
	"WgInspector/entities/client"
	client2 "WgInspector/usecase/client"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/5
 */

type SQLiteNoticeDB struct {
	db *sql.DB
}

func init() {
	auth, err := NewSQLiteNoticeDB()
	if err != nil {
		panic(fmt.Sprintf("初始化 SQLite KbaseNotice失败: %v", err))
	}
	client2.UseNoticeDB(auth)
}
func NewSQLiteNoticeDB() (*SQLiteNoticeDB, error) {
	const (
		filePath = "./app/notice.db"
		dsn      = "file:" + filePath
	)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 更新后的表结构
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS notice_contents (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		content      TEXT NOT NULL,
		time         DATETIME NOT NULL,
		confirm_stat TEXT NOT NULL DEFAULT 'Unread'
			CHECK(confirm_stat IN ('Unread', 'Read', 'UnConfirm', 'Allow', 'NotAllow')),
	    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    	updated_by      TEXT NOT NULL DEFAULT 'System'
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	db.SetMaxOpenConns(1)
	return &SQLiteNoticeDB{db: db}, nil
}

func (s *SQLiteNoticeDB) Get(page, pageSize int) ([]client.NoticeContent, error) {
	// 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
	SELECT id, content, time, confirm_stat 
	FROM notice_contents 
	ORDER BY time DESC 
	LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	var contents []client.NoticeContent
	for rows.Next() {
		var (
			id          int
			content     string
			timeStr     string
			confirmStat string
		)

		if err := rows.Scan(&id, &content, &timeStr, &confirmStat); err != nil {
			return nil, fmt.Errorf("解析行失败: %w", err)
		}

		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return nil, fmt.Errorf("时间解析错误: %w", err)
		}

		contents = append(contents, client.NoticeContent{
			ID:          id,
			Content:     content,
			Time:        t,
			ConfirmStat: confirmStat,
		})
	}

	return contents, nil
}

func (s *SQLiteNoticeDB) Create(content client.NoticeContent) error {
	// 设置默认状态
	if content.ConfirmStat == "" {
		content.ConfirmStat = client.Unread
	}

	// 验证状态有效性
	if !isValidStatus(content.ConfirmStat) {
		return fmt.Errorf("无效的状态值: %s", content.ConfirmStat)
	}

	insertSQL := `
	INSERT INTO notice_contents 
	(content, time, confirm_stat)
	VALUES (?, ?, ?)`

	result, err := s.db.Exec(insertSQL,
		content.Content,
		content.Time.Format(time.RFC3339),
		content.ConfirmStat,
	)

	if err != nil {
		return fmt.Errorf("创建失败: %w", err)
	}

	// 获取插入ID
	id, _ := result.LastInsertId()
	content.ID = int(id)
	return nil
}

func (s *SQLiteNoticeDB) Update(content client.NoticeContent) error {
	// 验证状态有效性
	if !isValidStatus(content.ConfirmStat) {
		return fmt.Errorf("无效的状态值: %s", content.ConfirmStat)
	}

	updateSQL := `
	UPDATE notice_contents 
	SET content = ?, time = ?, confirm_stat = ?
	WHERE id = ?`

	result, err := s.db.Exec(updateSQL,
		content.Content,
		content.Time.Format(time.RFC3339),
		content.ConfirmStat,
		content.ID,
	)

	if err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	// 检查受影响行数
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("未找到ID为%d的记录", content.ID)
	}

	return nil
}

// 辅助函数验证状态有效性
func isValidStatus(status string) bool {
	switch status {
	case client.Unread, client.Read, client.UnConfirm, client.Allow, client.NotAllow:
		return true
	default:
		return false
	}
}

var _ client.NoticeDB = (*SQLiteNoticeDB)(nil)
