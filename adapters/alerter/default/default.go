package _default

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
	alerter2 "WgInspector/usecase/alerter"
	log2 "log"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/18
 */

func init() {
	alerter2.RegisterDriver("default", AlerterDefault{})
}

type AlerterDefault struct {
}

func (e AlerterDefault) Init(config config.AlertConfig) (alerter.Alerter, error) {
	return AlerterDefault{}, nil
}

func (e AlerterDefault) Send(alerter.Content) error {
	log2.Printf("Alert Err - Empty Alert: this alerter has not init, please check config \n")
	return nil
}

var _ alerter.Alerter = (*AlerterDefault)(nil)
