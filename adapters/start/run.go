package start

import (
	"WgInspector/adapters/cron"
	"context"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

func Run(ctx context.Context) {
	cron.Start()
	select {
	case <-ctx.Done():
		cron.Exit()
		break
	}
}
