package client

import (
	"WgInspector/utils"
	"context"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/17
 */

const (
	AuthLevelUser = iota
	AuthLevelAdmin
)

// 身份验证
type Author interface {
	Init(utils.Option) error
	Auth(username, password string) (User, error)
	NewUser(User) error
	DeleteUser(username, password string) error
	UpdateUser(User) error
}

type User struct {
	UserName string
	Password string
	Level    int
}

// 客户端
type Client interface {
	Init(url string) (Client, error)
	Listen(ctx context.Context)
	UpdateCallback(ctx context.Context, configType string, data any) error
	Close() error
	Notice(content NoticeContent) error //用于给用户发送消息，可以包括确认消息和报警
}

const (
	Unread    = "Unread"
	Read      = "Read"
	UnConfirm = "UnConfirm"
	Allow     = "Allow"
	NotAllow  = "NotAllow"
)

type NoticeContent struct {
	ID          int
	Content     string
	Time        time.Time
	ConfirmStat string
	UpdatedAt   time.Time
	UpdatedBy   string
}

type NoticeDB interface {
	Get(page, pageSize int) ([]NoticeContent, error)
	Create(NoticeContent) error
	Update(NoticeContent) error
}
