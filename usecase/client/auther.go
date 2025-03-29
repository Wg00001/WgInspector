package client

import (
	"WgInspector/entities/client"
	"fmt"
	"sync"
)

var (
	auther   client.Author
	autherMu sync.RWMutex
)

// UseAuth 设置认证器
func UseAuth(a client.Author) {
	autherMu.Lock()
	defer autherMu.Unlock()
	auther = a
}

// Auth 验证用户身份
func Auth(username, password string) (client.User, error) {
	autherMu.RLock()
	defer autherMu.RUnlock()

	if auther == nil {
		return client.User{}, fmt.Errorf("未设置认证器")
	}
	return auther.Auth(username, password)
}

// NewUser 创建新用户
func NewUser(user client.User) error {
	autherMu.RLock()
	defer autherMu.RUnlock()

	if auther == nil {
		return fmt.Errorf("未设置认证器")
	}
	return auther.NewUser(user)
}

// DeleteUser 删除用户
func DeleteUser(username, password string) error {
	autherMu.RLock()
	defer autherMu.RUnlock()

	if auther == nil {
		return fmt.Errorf("未设置认证器")
	}
	return auther.DeleteUser(username, password)
}

// UpdateUser 更新用户信息
func UpdateUser(user client.User) error {
	autherMu.RLock()
	defer autherMu.RUnlock()

	if auther == nil {
		return fmt.Errorf("未设置认证器")
	}
	return auther.UpdateUser(user)
}
