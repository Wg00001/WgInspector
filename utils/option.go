package utils

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/27
 */

type Option map[string]string
type OptionFunc func(opt Option)

func (o *Option) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), &o)
}

func (o Option) Value() (driver.Value, error) {
	if o == nil {
		return nil, fmt.Errorf("id key is nil\n")
	}
	return json.Marshal(o)
}

func (o Option) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string(o))
}

func WithOption(opt map[string]string) Option {
	if opt == nil {
		return make(Option)
	}
	return opt
}

func (o Option) With(optionFunc ...OptionFunc) {
	for _, v := range optionFunc {
		v(o)
	}
}

func (o Option) WithOption(next Option) Option {
	for k, v := range next {
		o[k] = v
	}
	return o
}

func (o Option) GetOrDefault(key, def string) string {
	res, ok := o[key]
	if !ok {
		return def
	}
	return res
}

// todo: test copy
func (o Option) Unmarshall(v any) error {
	return DeepCopy(v, o)
}
