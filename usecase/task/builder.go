package task

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"fmt"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/10
 */

func NewInspTask(taskConfig config.TaskConfig) Task {
	return Task{Config: taskConfig}
}

// alert如果没有设置，那么应该继承父节点的alertID
func newTaskPlan(taskCfg config.TaskConfig) (res *taskPlan, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("init task fail: %s", err.Error())
		}
	}()
	//if taskCfg.ID == nil {
	//	return nil, fmt.Errorf("config is nil")
	//}
	res = &taskPlan{
		targetDBs: make([]*config.DBConfig, 0, len(taskCfg.TargetDB)),
		inspNodes: []*config.InspNode{},
	}
	for _, val := range taskCfg.TargetDB {
		dbcfg, err := config2.Get[config.DBConfig](config2.Key{
			ConfigType: config.TypeDB,
			Identity:   val,
		})
		if err != nil {
			return nil, err
		}
		res.targetDBs = append(res.targetDBs, &dbcfg)
	}

	//是否全选 (全部insp)
	if taskCfg.AllInspector {
		res.inspNodes = config2.GetAllInsp()
	}
	//添加todo列表的insp
	for _, val := range taskCfg.Todo {
		temp := config2.GetInsp(val)
		if temp == nil {
			continue
		}
		res.inspNodes = append(res.inspNodes, temp)
	}
	//去掉not to do的insp (使用hash连接)
	notToDo := make(map[config.Identity]bool, len(taskCfg.NotTodo))
	for _, val := range taskCfg.NotTodo {
		notToDo[val] = true
	}
	newArr := make([]*config.InspNode, 0, len(res.inspNodes))
	for _, val := range res.inspNodes {
		if !notToDo[val.Identity] {
			newArr = append(newArr, val)
		}
	}
	res.inspNodes = newArr
	return res, nil
}
