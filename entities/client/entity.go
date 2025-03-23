package client

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/17
 */

type Client interface {
	Init(url string) (Client, error)
	UpdateCallback(configType string, data any) error
}
