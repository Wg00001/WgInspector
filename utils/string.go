package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/19
 */

func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// 在大写字母前插入下划线（非首字符且前一个字符不是大写）
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || (unicode.IsUpper(rune(s[i-1])) && i < len(s)-1 && unicode.IsLower(rune(s[i+1])))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			// 非字母数字替换为下划线（避免连续）
			if len(result) > 0 && result[len(result)-1] != '_' {
				result = append(result, '_')
			}
		} else {
			result = append(result, r)
		}
	}
	// 合并连续下划线并去除首尾
	str := strings.ReplaceAll(string(result), "__", "_")
	return strings.Trim(str, "_")
}

func DeepCopy(dst interface{}, src interface{}) error {
	if dst == nil {
		return errors.New("dst cannot be nil")
	}
	if src == nil {
		return errors.New("src cannot be nil")
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("unable to marshal src: %w", err)
	}
	err = json.Unmarshal(bytes, dst)
	if err != nil {
		return fmt.Errorf("unable to unmarshal into dst: %w", err)
	}
	return nil
}
