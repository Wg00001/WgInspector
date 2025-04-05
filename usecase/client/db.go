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

// UseAuthor 设置认证器
func UseAuthor(a client.Author) {
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

var (
	noticeDB client.NoticeDB
	noticeMu sync.RWMutex
)

func UseNoticeDB(n client.NoticeDB) {
	noticeMu.Lock()
	defer noticeMu.Unlock()
	noticeDB = n
}

func GetNotice(page, pageSize int) ([]client.NoticeContent, error) {
	noticeMu.RLock()
	defer noticeMu.RUnlock()
	return noticeDB.Get(page, pageSize)
}

func CreateNotice(ctn client.NoticeContent) error {
	noticeMu.Lock()
	defer noticeMu.Unlock()
	return noticeDB.Create(ctn)
}

func UpdateNotice(ctn client.NoticeContent) error {
	noticeMu.Lock()
	defer noticeMu.Unlock()
	return noticeDB.Update(ctn)
}
