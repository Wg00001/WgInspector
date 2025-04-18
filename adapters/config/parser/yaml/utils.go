package yaml

import (
	"WgInspector/entities/config"
	insp2 "WgInspector/usecase/insp"
	"fmt"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/18
 */

const (
	keyAlertId   = "_alertId"
	keyAlertWhen = "_alertWhen"
	keySQL       = "_sql"
)

func ParseMap(n insp2.NodeBuilder, arg map[string]interface{}) (m insp2.NodeBuilder, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("inspect node build fail, please check inspect config \nerr: %v\n", r)
		}
	}()
	//alertId, ok := arg[keyAlertId]
	//if !ok {
	//	return n, nil
	//} else {
	//	n.AlertID = config.NewIdentity(alertId)
	//	delete(arg, keyAlertId)
	//}
	alertWhen, ok := arg[keyAlertWhen]
	if !ok {
		return n, nil
	} else {
		delete(arg, keyAlertWhen)
	}
	n.AlertWhen = alertWhen.(string)
	//todo:alert when save in struct
	//n = n.BuildAlertFunc(alertWhen.(string))

	sql, ok := arg[keySQL]
	if !ok {
		return n, nil
	} else {
		n.SQL = sql.(string)
		delete(arg, keySQL)
	}
	return n, nil
}

// Helper functions with panic for type conversion errors
func parseTime(s interface{}) time.Time {
	str, ok := s.(string)
	if !ok || s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		t, err = time.Parse(time.TimeOnly, str)
	}
	if err != nil {
		panic(err)
	}
	return t
}

func parseNames(s interface{}) []config.Identity {
	items, ok := s.([]interface{})
	if !ok {
		return nil
	}
	names := make([]config.Identity, 0, len(items))
	for _, item := range items {
		names = append(names, config.Identity(item.(string)))
	}
	return names
}

func parseStringSlice(data interface{}) []string {
	if data == nil {
		return nil
	}
	items, ok := data.([]interface{})
	if !ok {
		return nil
	}
	strs := make([]string, 0, len(items))
	for _, item := range items {
		strs = append(strs, fmt.Sprintf("%v", item))
	}
	return strs
}

func parseIntSlice(data interface{}) []int {
	if data == nil {
		return nil
	}
	items, ok := data.([]interface{})
	if !ok {
		return nil
	}
	ints := make([]int, 0, len(items))
	for _, item := range items {
		ints = append(ints, item.(int))
	}
	return ints
}

func parseWeekdaySlice(data interface{}) []time.Weekday {
	ints := parseIntSlice(data)
	res := make([]time.Weekday, len(ints))
	for i, v := range ints {
		res[i] = time.Weekday(v)
	}
	return res
}
