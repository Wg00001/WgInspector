package client

import (
	"WgInspector/entities/client"
	"WgInspector/utils"
	"fmt"
	"sync"
)

var (
	drivers    = make(map[string]driver)
	noticeDB   client.NoticeDB
	author     client.Author
	registed   bool = false
	clientDBMu sync.RWMutex
)

type driver struct {
	client.Author
	client.NoticeDB
}

func RegisterClientDatabaseDriver(name string, authorDriver client.Author, noticeDriver client.NoticeDB) error {
	clientDBMu.Lock()
	defer clientDBMu.Unlock()
	if drivers == nil {
		return fmt.Errorf("ClientDatabase driver has already registed")
	}
	drivers[name] = driver{
		Author:   authorDriver,
		NoticeDB: noticeDriver,
	}
	return nil
}

func registerUseClientDatabase(option utils.Option) error {
	clientDBMu.Lock()
	defer clientDBMu.Unlock()
	if registed {
		return fmt.Errorf("notice driver has already registed")
	}
	d := option.GetOrDefault("client_db_driver", "postgres")
	if n, ok := drivers[d]; !ok {
		return fmt.Errorf("notice driver [%s] not exist", d)
	} else {
		noticeDB = n.NoticeDB
		author = n.Author
		if err := author.Init(option); err != nil {
			return err
		}
		if err := noticeDB.Init(option); err != nil {
			return err
		}
		registed = true
	}
	return nil
}

// Auth 验证用户身份
func Auth(username, password string) (client.User, error) {
	if author == nil {
		return client.User{}, fmt.Errorf("未设置认证器")
	}
	return author.Auth(username, password)
}

// NewUser 创建新用户
func NewUser(user client.User) error {
	if author == nil {
		return fmt.Errorf("未设置认证器")
	}
	return author.NewUser(user)
}

// DeleteUser 删除用户
func DeleteUser(username, password string) error {
	if author == nil {
		return fmt.Errorf("未设置认证器")
	}
	return author.DeleteUser(username, password)
}

// UpdateUser 更新用户信息
func UpdateUser(user client.User) error {
	if author == nil {
		return fmt.Errorf("未设置认证器")
	}
	return author.UpdateUser(user)
}

func GetNotice(page, pageSize int) ([]client.NoticeContent, error) {
	if !registed {
		return nil, fmt.Errorf("notice didn't registed")
	}
	return noticeDB.Get(page, pageSize)
}

func GetNoticeByID(id int) (*client.NoticeContent, error) {
	if !registed {
		return nil, fmt.Errorf("notice didn't registed")
	}
	return noticeDB.GetByID(id)
}

func CreateNotice(ctn client.NoticeContent) error {
	if !registed {
		return fmt.Errorf("notice didn't registed")
	}
	return noticeDB.Create(ctn)
}

func UpdateNotice(ctn client.NoticeContent) error {
	if !registed {
		return fmt.Errorf("notice didn't registed")
	}
	return noticeDB.Update(ctn)
}
