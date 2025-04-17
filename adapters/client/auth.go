package client

import (
	"WgInspector/entities/client"
	client2 "WgInspector/usecase/client"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

const (
	defaultUsername = "admin"
	defaultPassword = "123456"
)

type SQLiteAuth struct {
	db *sql.DB
}

func init() {
	auth, err := NewSQLiteAuth()
	if err != nil {
		panic(fmt.Sprintf("初始化 SQLite 认证失败: %v", err))
	}
	client2.UseAuthor(auth)
}

var _ client.Author = (*SQLiteAuth)(nil)

func NewSQLiteAuth() (*SQLiteAuth, error) {
	// 分离文件路径和 DSN
	filePath := "./app/auth.db" // 实际文件路径
	dsn := "file:" + filePath   // modernc 要求的 DSN 格式

	// 确保数据库目录存在（基于实际文件路径）
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开/创建数据库（使用 DSN）
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 创建用户表（IF NOT EXISTS 确保幂等性）
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			level INTEGER NOT NULL DEFAULT 0  -- 0表示false，1表示true
		)
    `)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建用户表失败: %w", err)
	}

	auth := &SQLiteAuth{db: db}

	//没默认用户时会新建默认用户
	auth.NewUser(client.User{
		UserName: defaultUsername,
		Password: defaultPassword,
		Level:    client.AuthLevelAdmin,
	})
	return auth, nil
}

func (a *SQLiteAuth) Close() error {
	return a.db.Close()
}

func (a *SQLiteAuth) Auth(username, password string) (client.User, error) {
	var hashedPassword string
	var admin int

	err := a.db.QueryRow("SELECT password, level FROM users WHERE username = ?", username).
		Scan(&hashedPassword, &admin)
	if err != nil {
		if err == sql.ErrNoRows {
			return client.User{}, fmt.Errorf("用户不存在")
		}
		return client.User{}, fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return client.User{}, fmt.Errorf("密码错误")
	}
	return client.User{
		UserName: username,
		Level:    admin,
	}, nil
}

func (a *SQLiteAuth) NewUser(user client.User) error {
	// 检查用户是否已存在
	var exists bool
	err := a.db.QueryRow("SELECT 1 FROM users WHERE username = ?", user.UserName).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if exists {
		return fmt.Errorf("用户已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 插入新用户
	_, err = a.db.Exec("INSERT INTO users (username, password, level) VALUES (?, ?, ?)",
		user.UserName, string(hashedPassword), user.Level)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

func (a *SQLiteAuth) DeleteUser(username, password string) error {
	// 先验证用户
	_, err := a.Auth(username, password)
	if err != nil {
		return err
	}

	// 删除用户
	result, err := a.db.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

func (a *SQLiteAuth) UpdateUser(user client.User) error {
	// 检查用户是否存在
	var exists bool
	err := a.db.QueryRow("SELECT 1 FROM users WHERE username = ?", user.UserName).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查用户是否存在失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("用户不存在")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新用户密码
	_, err = a.db.Exec("UPDATE users SET password = ?,level = ? WHERE username = ?",
		string(hashedPassword), user.Level, user.UserName)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}
