package logger

import (
	"WgInspector/entities/config"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/1/19
 */

//根据config来设置多个logger，不同实现的logger的配置格式不同，在usecase中转换成标准格式送到adapter中各自解析。
//用户配置task时可以指定loggerId，没有指定则使用0号。

type Logger interface {
	Log(LogContent) error
	GetID() config.Identity
	Init(cfg config.LogConfig) (Logger, error)
	ReadLog(config.LogFilter) ([]LogContent, error)
}

type LogContent struct {
	Timestamp time.Time
	TaskName  string
	DBName    string
	InspName  string
	TaskID    string //task批次编号
	Result    []byte `gorm:"type:jsonb"`
}

func (receiver LogContent) TableName() string {
	return "inspect_log"
}
