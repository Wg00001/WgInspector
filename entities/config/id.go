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

type IdKeyArray []IdKey

func (a IdKeyArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	v, err := json.Marshal(a)
	return v, err
}

func (a *IdKeyArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法将数据库字段转换为字节数组")
	}
	if len(bytes) == 0 {
		*a = make([]IdKey, 0)
		return nil
	}
	err := json.Unmarshal(bytes, a)
	return err
}

func (lf LogFilter) Value() (driver.Value, error) {
	return json.Marshal(lf)
}

// Scan 从数据库JSONB字段反序列化到LogFilter结构体
func (lf *LogFilter) Scan(value interface{}) error {
	if value == nil {
		*lf = LogFilter{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法将数据库字段转换为字节数组，实际类型：%T", value)
	}

	if len(bytes) == 0 {
		*lf = LogFilter{}
		return nil
	}

	return json.Unmarshal(bytes, lf)
}
