package config

import (
	"database/sql/driver"
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

func (n Identity) IsZero() bool {
	return n.ID == 0
}

func (n Identity) ToString() string {
	marshal, err := json.Marshal(n)
	if err != nil {
		return fmt.Sprintf("{ID:%d,Name:%s}", n.ID, n.Name)
	}
	return string(marshal)
}

func (n *Identity) IdKey() IdKey {
	return IdKey(*n)
}

func (n *IdKey) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), &n)
}

func (n *IdKey) Value() (driver.Value, error) {
	if n == nil {
		return nil, fmt.Errorf("id key is nil\n")
	}
	return json.Marshal(n)
}

func (n *IdKey) Identity() Identity {
	return Identity(*n)
}
