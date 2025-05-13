package pgsql

/**
 * @description: TODO
 * @author Wg
 * @date 2025/5/7
 */

import (
	"WgInspector/entities/client"
	"WgInspector/utils"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	// Import a PostgreSQL driver, e.g., "github.com/lib/pq"
	// Make sure to add it to your go.mod file
	// _ "github.com/lib/pq"
)

const (
	defaultUsername = "admin"
	defaultPassword = "123456" // Consider making this configurable or more secure
)

type PgSQLAuth struct {
	db *sql.DB
}

var _ client.Author = (*PgSQLAuth)(nil)

// Init initializes the PgSQLAuth module, including database connection and table creation.
func (p *PgSQLAuth) Init(option utils.Option) error {
	dsn := option.GetOrDefault("dsn", "") // Get DSN from options
	if dsn == "" {
		return fmt.Errorf("pgsql auth: DSN not provided in options")
	}

	db, err := sql.Open("postgres", dsn) // Ensure "postgres" matches your driver
	if err != nil {
		return fmt.Errorf("打开 PostgreSQL 数据库失败: %w", err)
	}

	// Ping the database to ensure connectivity
	if err = db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("无法连接到 PostgreSQL 数据库: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		level INTEGER NOT NULL DEFAULT 0, -- 0 for regular user, 1 for admin, adjust as needed
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT DEFAULT 'system',
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_by TEXT DEFAULT 'system',
		deleted_at TIMESTAMPTZ,
		deleted_by TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	p.db = db // Assign the db to the struct instance

	// Check and create default admin user
	var adminUserExists bool
	// Use p.db here as it's now assigned
	err = p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)", defaultUsername).Scan(&adminUserExists)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("检查默认管理员用户是否存在时出错: %v\n", err)
		// Continue if table creation was successful, but log the error.
	}

	if !adminUserExists {
		defaultUser := client.User{
			UserName: defaultUsername,
			Password: defaultPassword,
			Level:    client.AuthLevelAdmin,
		}
		// Call NewUser on the initialized instance (p)
		if errCreate := p.NewUser(defaultUser); errCreate != nil {
			fmt.Printf("创建默认管理员用户失败: %v\n", errCreate)
			// Decide if this error should halt initialization or just be logged
		}
	}

	return nil
}

func (p *PgSQLAuth) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PgSQLAuth) Auth(username, password string) (client.User, error) {
	var hashedPassword string
	var level int
	fmt.Println(username, password)
	err := p.db.QueryRow("SELECT password, level FROM users WHERE username = $1 AND deleted_at IS NULL", username).
		Scan(&hashedPassword, &level)
	if err != nil {
		if err == sql.ErrNoRows {
			return client.User{}, fmt.Errorf("用户不存在或已被删除")
		}
		return client.User{}, fmt.Errorf("查询用户失败: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return client.User{}, fmt.Errorf("密码错误")
	}

	return client.User{
		UserName: username,
		Level:    level,
		// ID is not part of the interface's User struct
	}, nil
}

// NewUser creates a new user. The 'creator' argument specifies who is creating this user.
func (p *PgSQLAuth) NewUser(user client.User) error {
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)", user.UserName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户 '%s' 已存在", user.UserName)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	now := time.Now()
	// created_by and updated_by will use the table's DEFAULT 'system'
	_, err = p.db.Exec(
		"INSERT INTO users (username, password, level, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		user.UserName, string(hashedPassword), user.Level, now, now,
	)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// DeleteUser performs a soft delete on the user.
// The 'deleter' argument specifies who is performing the deletion.
func (p *PgSQLAuth) DeleteUser(username, password string) error {
	// Authenticate user before deletion
	_, err := p.Auth(username, password)
	if err != nil {
		return fmt.Errorf("删除用户前的身份验证失败: %w", err)
	}

	now := time.Now()
	// deleted_by will use a placeholder or be NULL if no table default is set for it and not provided here.
	// For consistency, let's explicitly set it if the column exists for that purpose.
	// Assuming 'system' as placeholder for deleted_by based on previous logic for created_by/updated_by defaults.
	result, err := p.db.Exec(
		"UPDATE users SET deleted_at = $1, deleted_by = $2, updated_at = $1, updated_by = $2 WHERE username = $3 AND deleted_at IS NULL",
		now, "system", username, // Using username to identify the user
	)
	if err != nil {
		return fmt.Errorf("软删除用户失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在或已被删除")
	}
	return nil
}

// UpdateUser updates a user's password and/or level.
// The 'updater' argument specifies who is performing the update.
// Note: This implementation updates password and level. Username updates are typically more complex.
func (p *PgSQLAuth) UpdateUser(user client.User) error {
	var currentPassword string // Needed if password is not being updated
	err := p.db.QueryRow("SELECT password FROM users WHERE username = $1 AND deleted_at IS NULL", user.UserName).Scan(&currentPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户 '%s' 不存在或已被删除", user.UserName)
		}
		return fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	var hashedPasswordToStore string
	if user.Password != "" {
		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("新密码加密失败: %w", err)
		}
		hashedPasswordToStore = string(newHashedPassword)
	} else {
		hashedPasswordToStore = currentPassword // Keep old password if new one is empty
	}

	now := time.Now()
	// updated_by will use the table's DEFAULT 'system'
	stmt := "UPDATE users SET level = $1, updated_at = $2"
	args := []interface{}{user.Level, now}

	if user.Password != "" { // Only include password in update if it was provided
		stmt += ", password = $3"
		args = append(args, hashedPasswordToStore)
		stmt += " WHERE username = $4 AND deleted_at IS NULL"
		args = append(args, user.UserName)
	} else {
		stmt += " WHERE username = $3 AND deleted_at IS NULL"
		args = append(args, user.UserName)
	}

	result, err := p.db.Exec(stmt, args...)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败 (更新用户时): %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("未找到用户进行更新，或用户已被删除")
	}
	return nil
}
