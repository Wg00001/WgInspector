package logger

import (
	"WgInspector/entities/config"
	"WgInspector/entities/db"
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
	Log(Content)
	GetID() config.Identity
	Init(cfg *config.LogConfig) (Logger, error)
	ReadLog(config.LogFilter) ([]Content, error)
}

type Content struct {
	Timestamp time.Time
	TaskName  config.Identity
	DBName    config.Identity
	InspName  config.Identity
	TaskID    string //task批次编号
	Result    db.Result
	ResultStr string
}
