package postgres

import (
	"WgInspector/entities/config"
	"WgInspector/entities/logger"
	config2 "WgInspector/usecase/config"
	logger2 "WgInspector/usecase/logger"
	"WgInspector/utils"
	"fmt"
	"gorm.io/gorm"
)

/**
 * @description: PostgreSQL日志实现
 * @author Wg
 * @date 2025/1/19
 */

func init() {
	logger2.RegisterDriver("postgres", LogPostgre{})
}

// LogPostgre 使用GORM实现的日志器
type LogPostgre struct {
	config.LogConfig
	conn      *gorm.DB
	DSN       string `json:"dsn"`
	TableName string `json:"table_name"`
}

func (l LogPostgre) Init(cfg config.LogConfig) (logger.Logger, error) {
	var temp LogPostgre
	err := cfg.Option.Unmarshall(&temp)
	if err != nil {
		return nil, err
	}
	if temp.DSN == "" {
		temp.DSN = config2.GetInitConfig().BaseDSN
	}
	if temp.TableName == "" {
		temp.TableName = logger.LogContent{}.TableName()
	}
	gormDB, err := utils.ConnectGormDB(temp.Driver, temp.DSN)
	if err != nil {
		return nil, err
	}

	// 自动创建表
	err = gormDB.AutoMigrate(&logger.LogContent{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate table: %w", err)
	}

	temp.conn = l.conn.Table(temp.TableName)
	temp.LogConfig = cfg
	return temp, nil
}

func (l LogPostgre) GetID() config.Identity {
	return l.LogConfig.Identity
}

func (l LogPostgre) Log(res logger.LogContent) error {
	// 确认连接和表名
	if l.conn == nil {
		return fmt.Errorf("Database connection not initialized")
	}

	// 插入记录
	result := l.conn.Create(&res)
	if result.Error != nil {
		return fmt.Errorf("Failed to insert log data: %v", result.Error)
	}

	return nil
}

func (l LogPostgre) ReadLog(filter config.LogFilter) ([]logger.LogContent, error) {
	// 检查连接
	if l.conn == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	query := l.conn
	// 应用过滤条件
	if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
		query = query.Where("timestamp BETWEEN ? AND ?", filter.StartTime, filter.EndTime)
	} else if !filter.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTime)
	} else if !filter.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTime)
	}

	// TaskNames 过滤
	if len(filter.TaskNames) > 0 {
		taskNames := make([]string, 0, len(filter.TaskNames))
		for _, tn := range filter.TaskNames {
			taskNames = append(taskNames, tn.Name)
		}
		query = query.Where("task_name IN ?", taskNames)
	}

	// DBIDs 过滤
	if len(filter.DBIDs) > 0 {
		dbNames := make([]string, 0, len(filter.DBIDs))
		for _, db := range filter.DBIDs {
			dbNames = append(dbNames, db.Name)
		}
		query = query.Where("db_name IN ?", dbNames)
	}

	// TaskIDs 过滤
	if len(filter.TaskIDs) > 0 {
		taskIDs := make([]string, 0, len(filter.TaskIDs))
		for _, tid := range filter.TaskIDs {
			taskIDs = append(taskIDs, tid.Name)
		}
		query = query.Where("task_id IN ?", taskIDs)
	}

	// InspNames 过滤
	if len(filter.InspNames) > 0 {
		inspNames := make([]string, 0, len(filter.InspNames))
		for _, in := range filter.InspNames {
			inspNames = append(inspNames, in.Name)
		}
		query = query.Where("inspect_name IN ?", inspNames)
	}

	// 执行查询
	var logRecords []logger.LogContent
	if err := query.Find(&logRecords).Error; err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return logRecords, nil
}
