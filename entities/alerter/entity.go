package alerter

import (
	"WgInspector/entities/config"
	"WgInspector/entities/db"
	"time"
)

/**
 * @description: 监控报警功能，对insp某个具体数值进行监控，到达某个值时触发对应报警
 * @author Wg
 * @date 2025/1/19
 */

type Alerter interface {
	Send(Content) error
	Init(config.AlertConfig) (Alerter, error)
}

type Content struct {
	TimeStamp    time.Time       //报警的时刻
	TaskName     config.Identity //发生报警的任务名
	TaskID       string
	DBName       config.Identity
	InspName     config.Identity
	Result       db.Result //发生报警时所产生的结果
	AlertWhen    string    //AlertWhen会作为配置项读取
	AlertBecause string
}

func (c Content) AddAlertInfo(when, because string) Content {
	c.AlertWhen = when
	c.AlertBecause = because
	return c
}

func (c Content) AddWhen(when string) Content {
	c.AlertWhen = when
	return c
}

// todo: 打印报警原因
func (c Content) AddBecause(because string) Content {
	c.AlertBecause = because
	return c
}
