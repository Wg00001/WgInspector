package config

import "fmt"

/**
 * @description:
 * @author Wg
 * @date 2025/2/10
 */

type Id interface {
	GetIdentity() Identity
}

func (n Identity) GetIdentity() Identity {
	return n
}

func (n Identity) ToString() string {
	return fmt.Sprintf("%d-%s", n.ID, n.Name)
}
