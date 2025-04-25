package config

import (
	"encoding/json"
	"fmt"
)

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
	marshal, err := json.Marshal(n)
	if err != nil {
		return fmt.Sprintf("{ID:%d,Name:%s}", n.ID, n.Name)
	}
	return string(marshal)
}
