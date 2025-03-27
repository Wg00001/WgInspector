package utils

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/27
 */

type Option map[string]string
type OptionFunc func(opt Option)

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

func (o Option) GetOrDefault(key, def string) string {
	res, ok := o[key]
	if !ok {
		return def
	}
	return res
}
