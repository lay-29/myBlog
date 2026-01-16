package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"MyBlog/internal/models"
	"MyBlog/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func New(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/register", h.register)
	api.POST("/login", h.login)
	api.GET("/blogs", h.blogs)
	api.POST("/blogs/:id/comments", h.addComment)
}

func (h *Handler) register(c *gin.Context) {
	var payload models.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析请求"})
		return
	}
	if err := h.store.RegisterUser(payload); err != nil {
		h.handleStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func (h *Handler) login(c *gin.Context) {
	var payload models.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析请求"})
		return
	}
	if err := h.store.LoginUser(payload); err != nil {
		h.handleStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "登录成功", "token": "demo-token"})
}

func (h *Handler) blogs(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.Blogs())
}

func (h *Handler) addComment(c *gin.Context) {
	var payload models.Comment
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析请求"})
		return
	}
	blog, err := h.store.AddComment(c.Param("id"), payload)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客 ID"})
			return
		}
		h.handleStoreError(c, err)
		return
	}
	c.JSON(http.StatusOK, blog)
}

func (h *Handler) handleStoreError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, store.ErrEmptyCredential):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, store.ErrUserExists):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, store.ErrInvalidUser):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, store.ErrEmptyComment):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, store.ErrBlogNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
	}
}
