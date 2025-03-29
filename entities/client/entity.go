package client

import "context"

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/17
 */

type User struct {
	UserName string
	Password string
}

type Client interface {
	Init(url string) (Client, error)
	Listen(ctx context.Context)
	UpdateCallback(configType string, data any) error
	Close() error
}
