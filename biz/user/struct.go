package user

import (
	"moredoc/pkg/types"
	"time"
)

// 用户登录响应
type LoginReply struct {
	User  *types.User `json:"user,omitempty"`
	Token string      `json:"token,omitempty"`
}

// 用户权限信息查询
type UserPermissionsReply struct {
	Permission []*Permission `json:"permission,omitempty"`
}

type Permission struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Method      string     `json:"method,omitempty"`
	Path        string     `json:"path,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Id          int64      `json:"id,omitempty"`
}

// 验证码响应
type CaptchaReply struct {
	Enable  bool   `json:"enable,omitempty"`
	Id      string `json:"id,omitempty"`
	Captcha string `json:"captcha,omitempty"`
	Type    string `json:"type,omitempty"`
}

// 用户注册登录请求
type RegisterAndLoginRequest struct {
	Username  string `json:"username,omitempty" validate:"min=3,max=32"`
	Password  string `json:"password,omitempty" validate:"min=6"`
	Captcha   string `json:"captcha,omitempty"`
	CaptchaId string `json:"captcha_id,omitempty"`
	Email     string `json:"email,omitempty"`
	Code      string `json:"code,omitempty"`
}
