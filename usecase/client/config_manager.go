package client

import (
	config2 "WgInspector/entities/config"
	"WgInspector/usecase/config"
	"github.com/wg00001/wgo-sdk/wg"
	"sort"
)

/**
 * @description: 客户端中用于配置更新的用例
 * @author Wg
 * @date 2025/3/17
 */

//本地代码内缓存一个可供直接读取的配置结构体
//两种实现方式
//1. 维护版本号
//2. 修改时调用回调函数

type Config struct {
}

func LoadCurrentConfig() {
	config.RLock()
	defer config.RUnlock()
	c := config.Index
	_ = toSortedSlice(c.DB)
}

func toSortedSlice[T config2.Id](origin map[config2.Identity]T) []T {
	arr := wg.MapToValueSlice(origin)
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].GetIdentity() > arr[j].GetIdentity()
	})
	return arr
}
