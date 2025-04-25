package start

import (
	"WgInspector/usecase/client"
	"WgInspector/usecase/task/cron"
	"context"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

func Run(ctx context.Context) {
	cron.Start()
	client.Listen(ctx)

	select {
	case <-ctx.Done():
		client.Close()
		cron.Stop()
	}
}
