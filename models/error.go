package models

import "errors"

var (
	ErrUserExists      = errors.New("用户名已存在")
	ErrInvalidUser     = errors.New("用户名或密码错误")
	ErrBlogNotFound    = errors.New("博客不存在")
	ErrEmptyComment    = errors.New("评论内容不能为空")
	ErrEmptyCredential = errors.New("用户名和密码不能为空")
)
