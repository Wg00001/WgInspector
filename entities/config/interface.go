package config

import "gorm.io/gorm"

/**
 * @author Wg
 * @date 2025/2/4
 */

type Reader interface { //配置读取器
	NewReader(db *gorm.DB) (Reader, error)
	ReadConfig() error          //读取配置并保存到配置中心
	SaveConfig(data Id) error   //将配置中心当前Meta配置持久化到存储中
	DeleteConfig(data Id) error //将配置中心当前Meta配置持久化到存储中
	Watch()                     //监听配置变更，将新变更应用到meta中，并将变更信息发送给client
}
