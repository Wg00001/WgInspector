package insp

import (
	"WgInspector/entities/config"
)

/**
 * @description: 用于构建insp的builder, 包括读取
 * @author Wg
 * @date 2025/1/19
 */

type NodeBuilder struct {
	config.InspNode
	error
}

func (n NodeBuilder) Build() (config.InspNode, error) {
	return n.InspNode, n.error
}

func (n NodeBuilder) WithName(name config.Identity) NodeBuilder {
	n.Name = name
	return n
}

func (n NodeBuilder) WithSQL(sql string) NodeBuilder {
	n.SQL = sql
	return n
}

//	func (n NodeBuilder) WithEmptyAlert() NodeBuilder {
//		n.AlertFunc = func(content alerter.Content) error {
//			return nil
//		}
//		return n
//
//func (n NodeBuilder) BuildAlertFunc(alertWhen string) NodeBuilder {
//	n.AlertFunc, n.error = buildAlertFunc(alertWhen, config.Identity(n.AlertID))
//	return n
//}
