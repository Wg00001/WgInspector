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
	err := client.RegisterClientDatabaseDriver(driverName, &SQLiteAuth{}, &SQLiteNoticeDB{})
	if err != nil {
		fmt.Printf("注册 SQLite 客户端数据库失败: %v", err)
		return
	}
}
