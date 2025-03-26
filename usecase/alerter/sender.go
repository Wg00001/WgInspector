package alerter

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/20
 */
var (
	inspAlertFuncTable map[inspAlertKey]func(alerter.Content) error
)

type inspAlertKey struct {
	InspName config.Identity
	SQL      string
}

func SendInspAlert(inode config.InspNode, content alerter.Content) {

}
