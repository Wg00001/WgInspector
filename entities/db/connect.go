package db

import (
	"WgInspector/entities/config"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/6
 */

type SqlDB struct {
	*sql.DB
	Config config.DBConfig
	Err    error
}

func (c *SqlDB) Connect() error {
	c.DB, c.Err = sql.Open(c.Config.Driver, c.Config.DSN)
	if c.Err != nil {
		return c.Err
	}
	return nil
}

func (c *SqlDB) Error() error {
	return c.Err
}

func (c *SqlDB) Query(query string, args ...any) (*sql.Rows, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.DB.Query(query, args...)
}
