package pgsql

/**
 * @description: TODO
 * @author Wg
 * @date 2025/5/7
 */

import (
	"WgInspector/entities/client"
	"WgInspector/utils"
	"database/sql"
	"fmt"
	"time"
	// _ "github.com/lib/pq" // Ensure PostgreSQL driver is imported
)

type PgSQLNoticeDB struct {
	db *sql.DB
}

var _ client.NoticeDB = (*PgSQLNoticeDB)(nil)

// Init initializes the PgSQLNoticeDB module.
func (p *PgSQLNoticeDB) Init(option utils.Option) error {
	dsn := option.GetOrDefault("dsn", "")
	if dsn == "" {
		return fmt.Errorf("pgsql notice: DSN not provided in options")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("打开 PostgreSQL 通知数据库失败: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("无法连接到 PostgreSQL 通知数据库: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS notice_contents (
		id SERIAL PRIMARY KEY,
		content TEXT NOT NULL,
		origin_data JSONB NOT NULL DEFAULT '{}',
		time TIMESTAMPTZ NOT NULL,
		confirm_stat TEXT NOT NULL DEFAULT 'Unread' 
			CHECK(confirm_stat IN ('Unread', 'Read', 'UnConfirm', 'Allow', 'NotAllow')),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT DEFAULT 'system',
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_by TEXT DEFAULT 'system'
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建通知表失败: %w", err)
	}

	p.db = db
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return nil
}

func (p *PgSQLNoticeDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PgSQLNoticeDB) Get(page, pageSize int) ([]client.NoticeContent, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 { // Max page size of 100 for safety
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Select only fields present in the provided client.NoticeContent struct
	query := `
	SELECT id, content, time, confirm_stat, updated_at, updated_by
	FROM notice_contents 
	ORDER BY time DESC 
	LIMIT $1 OFFSET $2`

	rows, err := p.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	defer rows.Close()

	var contents []client.NoticeContent
	for rows.Next() {
		var nc client.NoticeContent
		// Scan into fields available in client.NoticeContent
		// client.NoticeContent has: ID, Content, Time, ConfirmStat, UpdatedAt, UpdatedBy
		if err := rows.Scan(&nc.ID, &nc.Content, &nc.Time, &nc.ConfirmStat, &nc.UpdatedAt, &nc.UpdatedBy); err != nil {
			return nil, fmt.Errorf("扫描通知行失败: %w", err)
		}
		contents = append(contents, nc)
	}
	return contents, nil
}

func (p *PgSQLNoticeDB) GetByID(id int) (*client.NoticeContent, error) {
	// SQL查询语句（包含所有需要的字段）
	query := `
        SELECT 
            id,
            content,
            origin_data,
            time,
            confirm_stat,
            updated_at,
            updated_by
        FROM notice_contents 
        WHERE id = $1`

	// 执行查询
	row := p.db.QueryRow(query, id)

	var notice client.NoticeContent

	// 扫描结果到结构体（注意字段顺序必须与SELECT顺序一致）
	err := row.Scan(
		&notice.ID,
		&notice.Content,
		&notice.OriginData,
		&notice.Time,
		&notice.ConfirmStat,
		&notice.UpdatedAt,
		&notice.UpdatedBy,
	)

	// 错误处理
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notice with ID %d not found", id)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}
	return &notice, nil
}

// Create adds a new notice. 'creator' specifies who is creating this notice.
func (p *PgSQLNoticeDB) Create(notice client.NoticeContent) error {
	if notice.ConfirmStat == "" {
		notice.ConfirmStat = client.Unread // Default status
	}
	if !isValidNoticeStatus(notice.ConfirmStat) {
		return fmt.Errorf("无效的通知状态: %s", notice.ConfirmStat)
	}

	now := time.Now()
	// created_by and updated_by will use the table's DEFAULT 'system'
	insertSQL := `
	INSERT INTO notice_contents (content , time, confirm_stat, created_at, updated_at, origin_data) 
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var insertedID int // Not used further as client.NoticeContent from interface has ID but it's for input to Update
	err := p.db.QueryRow(insertSQL,
		notice.Content,
		notice.Time,
		notice.ConfirmStat,
		now, // created_at
		now, // updated_at
		notice.OriginData,
	).Scan(&insertedID)

	if err != nil {
		return fmt.Errorf("创建通知失败: %w", err)
	}
	// notice.ID = insertedID // The client.NoticeContent might not have ID if it's for input only.
	// If client.NoticeContent is also used for output and needs the ID, ensure it's settable.
	return nil
}

// Update modifies an existing notice. 'updater' specifies who is performing the update.
func (p *PgSQLNoticeDB) Update(notice client.NoticeContent) error {
	if !isValidNoticeStatus(notice.ConfirmStat) {
		return fmt.Errorf("无效的通知状态: %s", notice.ConfirmStat)
	}

	now := time.Now()
	// updated_by will use the table's DEFAULT 'system'
	updateSQL := `
	UPDATE notice_contents 
	SET time = $1, confirm_stat = $2, updated_at = $3
	WHERE id = $4`

	result, err := p.db.Exec(updateSQL,
		notice.Time,
		notice.ConfirmStat,
		now, // updated_at
		notice.ID,
	)

	if err != nil {
		return fmt.Errorf("更新通知失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败 (更新通知时): %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("未找到ID为 %d 的通知进行更新", notice.ID)
	}
	return nil
}

// isValidNoticeStatus checks if the provided status is valid for a notice.
func isValidNoticeStatus(status string) bool {
	switch status {
	case client.Unread, client.Read, client.UnConfirm, client.Allow, client.NotAllow:
		return true
	default:
		return false
	}
}
