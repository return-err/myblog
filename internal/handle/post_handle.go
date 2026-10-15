package handle

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type PostService interface {
	CreatePost(userID int64, title, content string) error
	GetPostTitle(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error)
	PostDetail(postID int64) (model.PostDetail, error)
	DeletePost(postID, userID int64, userRole string) error
	EditPost(postID, userID int64, title, content, userRole string) error
	GetAllPosts(page, pageSize int) ([]model.ArticleSummary, int64, error)
	SearchPost(keyword string) ([]model.ArticleSummary, error)
}

type PostHandle struct {
	postService PostService
}

func NewPostHandle(s PostService) *PostHandle {
	return &PostHandle{postService: s}
}

func (h *PostHandle) CreatePost(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, apperror.ErrUnauthorized)
		return
	}
	userID := userIDVal.(int64)

	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	err := h.postService.CreatePost(userID, req.Title, req.Content)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully"})
}

func (h *PostHandle) PostDetail(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	post, err := h.postService.PostDetail(postID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post": post,
	})
}

func (h *PostHandle) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	userID := c.GetInt64("userID")
	userRole := c.GetString("role")

	err = h.postService.DeletePost(postID, userID, userRole)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

func (h *PostHandle) EditPost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	var req model.UpdatePostRequest
	if err := c.ShouldBind(&req); err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	userID := c.GetInt64("userID")
	userRole := c.GetString("role")

	err = h.postService.EditPost(postID, userID, req.Title, req.Content, userRole)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post updated successfully"})
}

func (h *PostHandle) PostsList(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page")) // strconv.Atoi stringをintに変換
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		pageSize = 10
	}

	posts, totalCount, err := h.postService.GetAllPosts(page, pageSize)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.PageResult{
		Posts:      posts,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (h *PostHandle) SearchPost(c *gin.Context) {
	keyword := c.Query("keyword")
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		c.JSON(http.StatusOK, []model.ArticleSummary{})
		return
	}

	posts, err := h.postService.SearchPost(keyword)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *PostHandle) GetUserPosts(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		pageSize = 10
	}

	posts, totalCount, err := h.postService.GetPostTitle(userID, page, pageSize)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.PageResult{
		Posts:      posts,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}
