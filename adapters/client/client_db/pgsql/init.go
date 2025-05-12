package pgsql

import (
	"WgInspector/usecase/client"
	"fmt"
)

const driverName = "postgres"

/**
 * @description: TODO
 * @author Wg
 * @date 2025/5/7
 */

func init() {
	err := client.RegisterClientDatabaseDriver(driverName, &PgSQLAuth{}, &PgSQLNoticeDB{})
	if err != nil {
		fmt.Printf("注册 PostgreSQL 客户端数据库驱动 '%s' 失败: %v\n", driverName, err)
		return
	}
	fmt.Printf("PostgreSQL 客户端数据库驱动 '%s' 已成功注册。\n", driverName)
}
