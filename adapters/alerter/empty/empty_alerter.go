package empty

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
	alerter2 "WgInspector/usecase/alerter"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/3
 */

func init() {
	alerter2.RegisterDriver("empty", AlerterEmpty{})
}

type AlerterEmpty struct {
}

func (e AlerterEmpty) Init(config config.AlertConfig) (alerter.Alerter, error) {
	return AlerterEmpty{}, nil
}

func (e AlerterEmpty) Send(alerter.Content) error {
	return nil
}

var _ alerter.Alerter = (*AlerterEmpty)(nil)
