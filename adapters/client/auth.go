package client

import (
	"WgInspector/entities/client"
	client2 "WgInspector/usecase/client"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
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
	client2.UseAuth(auth)
}

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
            password TEXT NOT NULL
        )
    `)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建用户表失败: %w", err)
	}

	auth := &SQLiteAuth{db: db}

	// 检查默认用户（带错误重试机制）
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err == nil {
			if count == 0 {
				if err := auth.NewUser(client.User{
					UserName: defaultUsername,
					Password: defaultPassword,
				}); err != nil {
					db.Close()
					return nil, fmt.Errorf("创建默认用户失败: %w", err)
				}
				log.Printf("已创建默认用户，用户名: %s, 密码: %s", defaultUsername, defaultPassword)
			}
			break
		}

		// 增强型错误检测（兼容 modernc 的错误信息）
		if i < maxRetries-1 && err.Error() == "database is locked" {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		db.Close()
		return nil, fmt.Errorf("检查用户数量失败: %w", err)
	}

	return auth, nil
}

func (a *SQLiteAuth) Close() error {
	return a.db.Close()
}

func (a *SQLiteAuth) Auth(username, password string) (client.User, error) {
	var user client.User
	var hashedPassword string

	err := a.db.QueryRow("SELECT username, password FROM users WHERE username = ?", username).
		Scan(&user.UserName, &hashedPassword)
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

	return user, nil
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
	_, err = a.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)",
		user.UserName, string(hashedPassword))
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
	_, err = a.db.Exec("UPDATE users SET password = ? WHERE username = ?",
		string(hashedPassword), user.UserName)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}
