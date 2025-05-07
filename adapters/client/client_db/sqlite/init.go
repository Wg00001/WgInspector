package sqlite

import (
	"WgInspector/usecase/client"
	"fmt"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/5/7
 */

const driverName = "sqlite"

func init() {
	auth, err := NewSQLiteAuth()
	if err != nil {
		fmt.Printf("初始化 SQLite 认证表失败: %v", err)
		return
	}
	notice, err := NewSQLiteNoticeDB()
	if err != nil {
		fmt.Printf("初始化 SQLite 通知表失败: %v", err)
		return
	}
	err = client.RegisterClientDatabaseDriver(driverName, auth, notice)
	if err != nil {
		fmt.Printf("注册 SQLite 客户端数据库失败: %v", err)
		return
	}
}
