package alerter

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
	"fmt"
	"sync"
)

/**
 * @description: 读取配置后，alert在此处注册，insp执行后根据对应的id和目标来检查
 * @author Wg
 * @date 2025/2/15
 */

var pool = sync.Map{}

func Register(id config.Identity, alert alerter.Alerter) error {
	if _, ok := pool.Load(id); ok {
		return fmt.Errorf("alerter registry fail: alert is already exist, alert id repeat")
	}
	pool.Store(id, alert)
	return nil
}

func GetAlert(id config.Identity) alerter.Alerter {
	val, ok := pool.Load(id)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	t, ok := val.(alerter.Alerter)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	return t
}
