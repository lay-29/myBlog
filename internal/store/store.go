package store

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"MyBlog/internal/models"
)

var (
	ErrUserExists      = errors.New("用户名已存在")
	ErrInvalidUser     = errors.New("用户名或密码错误")
	ErrBlogNotFound    = errors.New("博客不存在")
	ErrEmptyComment    = errors.New("评论内容不能为空")
	ErrEmptyCredential = errors.New("用户名和密码不能为空")
)

type Store struct {
	mu            sync.Mutex
	users         map[string]string
	blogs         []models.Blog
	nextCommentID int
}

func NewStore() *Store {
	return &Store{
		users: map[string]string{},
		blogs: []models.Blog{
			{
				ID:        1,
				Title:     "欢迎来到我的博客",
				Author:    "站长",
				Content:   "这是我的第一篇博客，记录了我搭建个人站点的想法和计划。",
				CreatedAt: "2024-07-01",
				Comments: []models.Comment{
					{
						ID:        1,
						Author:    "访客",
						Content:   "期待更多内容！",
						CreatedAt: "2024-07-02",
					},
				},
			},
			{
				ID:        2,
				Title:     "关于技术栈的选择",
				Author:    "站长",
				Content:   "前端使用 Vue，后端使用 Go，配合轻量的 API 设计，快速完成迭代。",
				CreatedAt: "2024-07-03",
				Comments:  []models.Comment{},
			},
		},
		nextCommentID: 2,
	}
}

func (s *Store) RegisterUser(user models.User) error {
	if user.Username == "" || user.Password == "" {
		return ErrEmptyCredential
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.Username]; exists {
		return ErrUserExists
	}
	s.users[user.Username] = user.Password
	return nil
}

func (s *Store) LoginUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	password, exists := s.users[user.Username]
	if !exists || password != user.Password {
		return ErrInvalidUser
	}
	return nil
}

func (s *Store) Blogs() []models.Blog {
	s.mu.Lock()
	defer s.mu.Unlock()
	blogs := make([]models.Blog, len(s.blogs))
	copy(blogs, s.blogs)
	return blogs
}

func (s *Store) AddComment(blogID string, comment models.Comment) (models.Blog, error) {
	id, err := strconv.Atoi(blogID)
	if err != nil {
		return models.Blog{}, err
	}
	if comment.Content == "" {
		return models.Blog{}, ErrEmptyComment
	}
	if comment.Author == "" {
		comment.Author = "匿名"
	}
	comment.CreatedAt = time.Now().Format("2006-01-02 15:04")

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.blogs {
		if s.blogs[i].ID == id {
			s.nextCommentID++
			comment.ID = s.nextCommentID
			s.blogs[i].Comments = append(s.blogs[i].Comments, comment)
			return s.blogs[i], nil
		}
	}
	return models.Blog{}, ErrBlogNotFound
}
