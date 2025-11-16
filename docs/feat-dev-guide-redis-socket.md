# 🚀 Complete Feature Development Guide

## Table of Contents


---

## Architecture Overview

### Project Structure

```
go-backend/
├── cmd/api/main.go                 # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   ├── database/                   # Database connections
│   ├── middleware/                 # HTTP middleware
│   ├── models/                     # Database models (shared)
│   ├── modules/                    # Feature modules
│   │   ├── auth/                   # Example: Auth module
│   │   │   ├── dto/               # Request/Response DTOs
│   │   │   ├── handler.go         # HTTP handlers
│   │   │   ├── repository.go      # Database operations
│   │   │   ├── service.go         # Business logic
│   │   │   └── routes.go          # Route definitions
│   │   ├── posts/                 # Your new module here
│   │   └── users/
│   ├── services/                  # Shared services
│   │   ├── cache/                 # Redis operations
│   │   ├── email/
│   │   └── upload/
│   ├── utils/                     # Utility functions
│   │   ├── helpers/               # WebSocket/Redis helpers
│   │   ├── jwt/
│   │   ├── logger/
│   │   ├── password/
│   │   └── response/              # Standard responses
│   └── websocket/                 # WebSocket implementation
└── migrations/                     # Database migrations
```

### Core Principles

1. **Modularity**: Each feature is a self-contained module
2. **Separation of Concerns**: Handler → Service → Repository
3. **Developer-Friendly**: Simple, consistent patterns
4. **Production-Ready**: Logging, error handling, validation

---

## Quick Start Checklist

Before adding a new feature, ensure:

- [ ] Server is running: `make dev`
- [ ] Redis is running: `make redis-start`
- [ ] Database is connected
- [ ] You understand the feature requirements
- [ ] You know if it needs:
    - [ ] Database models
    - [ ] Real-time updates (WebSocket)
    - [ ] Caching (Redis)
    - [ ] File uploads
    - [ ] Background jobs

---

## Feature Development Process

### Process Flow

```
1. Define Requirements
   ↓
2. Create Database Model (if needed)
   ↓
3. Create Migration
   ↓
4. Create Module Structure
   ↓
5. Implement Repository (Database Layer)
   ↓
6. Implement Service (Business Logic)
   ↓
7. Implement Handler (HTTP Layer)
   ↓
8. Define Routes
   ↓
9. Register Module in main.go
   ↓
10. Add Redis/WebSocket (if needed)
   ↓
11. Test & Document
```

---

## Step-by-Step Examples

## Example 1: Simple CRUD Feature (Posts)

### Step 1: Define Requirements

**Feature**: Blog Posts System

- Users can create, read, update, delete posts
- Posts have title, content, author, status
- Need to cache popular posts
- Need real-time notifications when posts are published

### Step 2: Create Database Model

Create `internal/models/post.go`:

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Post struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Slug      string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	AuthorID  string         `gorm:"type:uuid;not null;index" json:"authorId"`
	Author    User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Status    string         `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, published, archived
	ViewCount int            `gorm:"default:0" json:"viewCount"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Post) TableName() string {
	return "posts"
}
```

### Step 3: Create Migration

```bash
make migrate-create name=create_posts_table
```

Create `migrations/000002_create_posts_table.up.sql`:

```sql
-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create posts table
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'draft',
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_deleted_at ON posts(deleted_at);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
```

Create `migrations/000002_create_posts_table.down.sql`:

```sql
DROP TABLE IF EXISTS posts;
```

Run migration:

```bash
# Set your database URL
export DB_URL="postgresql://go_backend_admin:goPass@localhost:5432/go_backend?sslmode=disable"
make migrate-up
```

### Step 4: Create Module Structure

```bash
mkdir -p internal/modules/posts/dto
touch internal/modules/posts/dto/request.go
touch internal/modules/posts/dto/response.go
touch internal/modules/posts/repository.go
touch internal/modules/posts/service.go
touch internal/modules/posts/handler.go
touch internal/modules/posts/routes.go
```

### Step 5: Create DTOs

Create `internal/modules/posts/dto/request.go`:

```go
package dto

import "errors"

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=255"`
	Content string `json:"content" binding:"required,min=10"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published"`
}

func (r *CreatePostRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if r.Status == "" {
		r.Status = "draft"
	}
	return nil
}

type UpdatePostRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=3,max=255"`
	Content *string `json:"content" binding:"omitempty,min=10"`
	Status  *string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type ListPostsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=draft published archived"`
	AuthorID string `form:"authorId" binding:"omitempty,uuid"`
	Search   string `form:"search" binding:"omitempty"`
}

func (r *ListPostsRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 10
	}
}
```

Create `internal/modules/posts/dto/response.go`:

```go
package dto

import (
	"time"
	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
)

type PostResponse struct {
	ID        string                  `json:"id"`
	Title     string                  `json:"title"`
	Content   string                  `json:"content"`
	Slug      string                  `json:"slug"`
	AuthorID  string                  `json:"authorId"`
	Author    *authdto.UserResponse   `json:"author,omitempty"`
	Status    string                  `json:"status"`
	ViewCount int                     `json:"viewCount"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

type PostListResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	AuthorID  string    `json:"authorId"`
	Status    string    `json:"status"`
	ViewCount int       `json:"viewCount"`
	CreatedAt time.Time `json:"createdAt"`
}

func ToPostResponse(post *models.Post) *PostResponse {
	resp := &PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		Slug:      post.Slug,
		AuthorID:  post.AuthorID,
		Status:    post.Status,
		ViewCount: post.ViewCount,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
	
	if post.Author.ID != "" {
		resp.Author = authdto.ToUserResponse(&post.Author)
	}
	
	return resp
}

func ToPostListResponse(post *models.Post) *PostListResponse {
	return &PostListResponse{
		ID:        post.ID,
		Title:     post.Title,
		Slug:      post.Slug,
		AuthorID:  post.AuthorID,
		Status:    post.Status,
		ViewCount: post.ViewCount,
		CreatedAt: post.CreatedAt,
	}
}
```

### Step 6: Implement Repository

Create `internal/modules/posts/repository.go`:

```go
package posts

import (
	"context"
	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, post *models.Post) error
	FindByID(ctx context.Context, id string) (*models.Post, error)
	FindBySlug(ctx context.Context, slug string) (*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Post, int64, error)
	IncrementViewCount(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("id = ?", id).
		First(&post).Error
	return &post, err
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("slug = ?", slug).
		First(&post).Error
	return &post, err
}

func (r *repository) Update(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Post{}, "id = ?", id).Error
}

func (r *repository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Post{})
	
	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if authorID, ok := filters["authorId"].(string); ok && authorID != "" {
		query = query.Where("author_id = ?", authorID)
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// Count total
	query.Count(&total)
	
	// Paginate
	offset := (page - 1) * limit
	err := query.
		Preload("Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	
	return posts, total, err
}

func (r *repository) IncrementViewCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}
```

### Step 7: Implement Service

Create `internal/modules/posts/service.go`:

```go
package posts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/posts/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	CreatePost(ctx context.Context, userID string, req dto.CreatePostRequest) (*dto.PostResponse, error)
	GetPost(ctx context.Context, id string) (*dto.PostResponse, error)
	GetPostBySlug(ctx context.Context, slug string) (*dto.PostResponse, error)
	UpdatePost(ctx context.Context, userID, postID string, req dto.UpdatePostRequest) (*dto.PostResponse, error)
	DeletePost(ctx context.Context, userID, postID string) error
	ListPosts(ctx context.Context, req dto.ListPostsRequest) ([]*dto.PostListResponse, int64, error)
	PublishPost(ctx context.Context, userID, postID string) (*dto.PostResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(ctx context.Context, userID string, req dto.CreatePostRequest) (*dto.PostResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}
	
	// Generate slug
	postSlug := slug.Make(req.Title)
	
	// Check if slug exists
	_, err := s.repo.FindBySlug(ctx, postSlug)
	if err == nil {
		// Slug exists, make it unique
		postSlug = fmt.Sprintf("%s-%d", postSlug, time.Now().Unix())
	}
	
	post := &models.Post{
		Title:    req.Title,
		Content:  req.Content,
		Slug:     postSlug,
		AuthorID: userID,
		Status:   req.Status,
	}
	
	if err := s.repo.Create(ctx, post); err != nil {
		logger.Error("failed to create post", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create post", err)
	}
	
	// Fetch with author info
	post, _ = s.repo.FindByID(ctx, post.ID)
	
	logger.Info("post created", "postID", post.ID, "authorID", userID)
	
	return dto.ToPostResponse(post), nil
}

func (s *service) GetPost(ctx context.Context, id string) (*dto.PostResponse, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("post:%s", id)
	var post models.Post
	
	err := cache.GetJSON(ctx, cacheKey, &post)
	if err == nil {
		// Cache hit
		logger.Debug("post cache hit", "postID", id)
		return dto.ToPostResponse(&post), nil
	}
	
	// Cache miss, get from database
	post2, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, response.NotFoundError("Post")
	}
	
	// Increment view count asynchronously
	go s.repo.IncrementViewCount(context.Background(), id)
	
	// Cache the post (5 minutes)
	cache.SetJSON(ctx, cacheKey, post2, 5*time.Minute)
	
	return dto.ToPostResponse(post2), nil
}

func (s *service) GetPostBySlug(ctx context.Context, slug string) (*dto.PostResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("post:slug:%s", slug)
	var post models.Post
	
	err := cache.GetJSON(ctx, cacheKey, &post)
	if err == nil {
		return dto.ToPostResponse(&post), nil
	}
	
	// Get from database
	post2, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, response.NotFoundError("Post")
	}
	
	// Increment view count
	go s.repo.IncrementViewCount(context.Background(), post2.ID)
	
	// Cache it
	cache.SetJSON(ctx, cacheKey, post2, 5*time.Minute)
	
	return dto.ToPostResponse(post2), nil
}

func (s *service) UpdatePost(ctx context.Context, userID, postID string, req dto.UpdatePostRequest) (*dto.PostResponse, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, response.NotFoundError("Post")
	}
	
	// Check ownership
	if post.AuthorID != userID {
		return nil, response.ForbiddenError("You can only update your own posts")
	}
	
	// Update fields
	if req.Title != nil {
		post.Title = *req.Title
		post.Slug = slug.Make(*req.Title)
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Status != nil {
		post.Status = *req.Status
	}
	
	if err := s.repo.Update(ctx, post); err != nil {
		return nil, response.InternalServerError("Failed to update post", err)
	}
	
	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("post:%s", postID))
	cache.Delete(ctx, fmt.Sprintf("post:slug:%s", post.Slug))
	
	post, _ = s.repo.FindByID(ctx, post.ID)
	
	logger.Info("post updated", "postID", postID, "userID", userID)
	
	return dto.ToPostResponse(post), nil
}

func (s *service) DeletePost(ctx context.Context, userID, postID string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return response.NotFoundError("Post")
	}
	
	if post.AuthorID != userID {
		return response.ForbiddenError("You can only delete your own posts")
	}
	
	if err := s.repo.Delete(ctx, postID); err != nil {
		return response.InternalServerError("Failed to delete post", err)
	}
	
	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("post:%s", postID))
	cache.Delete(ctx, fmt.Sprintf("post:slug:%s", post.Slug))
	
	logger.Info("post deleted", "postID", postID, "userID", userID)
	
	return nil
}

func (s *service) ListPosts(ctx context.Context, req dto.ListPostsRequest) ([]*dto.PostListResponse, int64, error) {
	req.SetDefaults()
	
	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.AuthorID != "" {
		filters["authorId"] = req.AuthorID
	}
	if req.Search != "" {
		filters["search"] = req.Search
	}
	
	posts, total, err := s.repo.List(ctx, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch posts", err)
	}
	
	result := make([]*dto.PostListResponse, len(posts))
	for i, post := range posts {
		result[i] = dto.ToPostListResponse(post)
	}
	
	return result, total, nil
}

func (s *service) PublishPost(ctx context.Context, userID, postID string) (*dto.PostResponse, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, response.NotFoundError("Post")
	}
	
	if post.AuthorID != userID {
		return nil, response.ForbiddenError("You can only publish your own posts")
	}
	
	if post.Status == "published" {
		return nil, response.BadRequest("Post is already published")
	}
	
	post.Status = "published"
	if err := s.repo.Update(ctx, post); err != nil {
		return nil, response.InternalServerError("Failed to publish post", err)
	}
	
	// Send notification to followers (example)
	go func() {
		// Send WebSocket notification
		notification := map[string]interface{}{
			"type":    "new_post",
			"postId":  post.ID,
			"title":   post.Title,
			"author":  userID,
			"message": fmt.Sprintf("New post published: %s", post.Title),
		}
		
		// Broadcast to all users (or send to specific followers)
		helpers.BroadcastNotification(notification)
		
		logger.Info("post published notification sent", "postID", post.ID)
	}()
	
	// Invalidate cache
	cache.Delete(ctx, fmt.Sprintf("post:%s", postID))
	
	post, _ = s.repo.FindByID(ctx, post.ID)
	
	logger.Info("post published", "postID", postID, "userID", userID)
	
	return dto.ToPostResponse(post), nil
}
```

### Step 8: Implement Handler

Create `internal/modules/posts/handler.go`:

```go
package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/posts/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreatePost godoc
// @Summary Create a new post
// @Tags posts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Post data"
// @Success 201 {object} response.Response{data=dto.PostResponse}
// @Router /posts [post]
func (h *Handler) CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	
	post, err := h.service.CreatePost(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, post, "Post created successfully")
}

// GetPost godoc
// @Summary Get post by ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Router /posts/{id} [get]
func (h *Handler) GetPost(c *gin.Context) {
	postID := c.Param("id")
	
	post, err := h.service.GetPost(c.Request.Context(), postID)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, post, "Post retrieved successfully")
}

// GetPostBySlug godoc
// @Summary Get post by slug
// @Tags posts
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Router /posts/slug/{slug} [get]
func (h *Handler) GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	
	post, err := h.service.GetPostBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, post, "Post retrieved successfully")
}

// UpdatePost godoc
// @Summary Update post
// @Tags posts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body dto.UpdatePostRequest true "Update data"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Router /posts/{id} [put]
func (h *Handler) UpdatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("id")
	
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	
	post, err := h.service.UpdatePost(c.Request.Context(), userID.(string), postID, req)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, post, "Post updated successfully")
}

// DeletePost godoc
// @Summary Delete post
// @Tags posts
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response
// @Router /posts/{id} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("id")
	
	if err := h.service.DeletePost(c.Request.Context(), userID.(string), postID); err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, nil, "Post deleted successfully")
}

// ListPosts godoc
// @Summary List posts
// @Tags posts
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param status query string false "Filter by status"
// @Param authorId query string false "Filter by author"
// @Param search query string false "Search in title/content"
// @Success 200 {object} response.Response{data=[]dto.PostListResponse}
// @Router /posts [get]
func (h *Handler) ListPosts(c *gin.Context) {
	var req dto.ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}
	
	posts, total, err := h.service.ListPosts(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	
	// Add pagination metadata
	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	
	response.Paginated(c, posts, pagination, "Posts retrieved successfully")
}

// PublishPost godoc
// @Summary Publish a draft post
// @Tags posts
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 200 {object} response.Response{data=dto.PostResponse}
// @Router /posts/{id}/publish [post]
func (h *Handler) PublishPost(c *gin.Context) {
	userID, _ := c.Get("userID")
	postID := c.Param("id")
	
	post, err := h.service.PublishPost(c.Request.Context(), userID.(string), postID)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, post, "Post published successfully")
}
```

### Step 9: Define Routes

Create `internal/modules/posts/routes.go`:

```go
package posts

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	posts := router.Group("/posts")
	{
		// Public routes
		posts.GET("", handler.ListPosts)
		posts.GET("/:id", handler.GetPost)
		posts.GET("/slug/:slug", handler.GetPostBySlug)
		
		// Protected routes
		posts.Use(authMiddleware)
		posts.POST("", handler.CreatePost)
		posts.PUT("/:id", handler.UpdatePost)
		posts.DELETE("/:id", handler.DeletePost)
		posts.POST("/:id/publish", handler.PublishPost)
	}
}
```

### Step 10: Register Module in main.go

Update `cmd/api/main.go`:

```go
// Add import
import (
	"github.com/umar5678/go-backend/internal/modules/posts"
)

// In main() function, after auth module registration:
func main() {
	// ... existing code ...
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

		// Auth module
		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, cfg)
		authHandler := auth.NewHandler(authService)
		authMiddleware := middleware.Auth(cfg)
		auth.RegisterRoutes(v1, authHandler, authMiddleware)
		
		// Posts module (NEW)
		postsRepo := posts.NewRepository(db)
		postsService := posts.NewService(postsRepo)
		postsHandler := posts.NewHandler(postsService)
		posts.RegisterRoutes(v1, postsHandler, authMiddleware)
	}
	
	// ... rest of code ...
}
```

### Step 11: Update Swagger Documentation

```bash
make swagger
```

### Step 12: Test the Feature

```bash
# Start server
make dev

# Create a post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "This is the content of my first post",
    "status": "draft"
  }'

# Get post by ID
curl http://localhost:8080/api/v1/posts/POST_ID

# List posts
curl "http://localhost:8080/api/v1/posts?page=1&limit=10"

# Publish post
curl -X POST http://localhost:8080/api/v1/posts/POST_ID/publish \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update post
curl -X PUT http://localhost:8080/api/v1/posts/POST_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "content": "Updated content"
  }'

# Delete post
curl -X DELETE http://localhost:8080/api/v1/posts/POST_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Example 2: Real-Time Chat Feature with WebSocket

### Requirements

- Users can send direct messages
- Real-time message delivery
- Typing indicators
- Read receipts
- Message history
- Online/offline status

### Step 1: Create Message Model

Create `internal/models/message.go`:

```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type Message struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SenderID   string         `gorm:"type:uuid;not null;index" json:"senderId"`
	ReceiverID string         `gorm:"type:uuid;not null;index" json:"receiverId"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Type       string         `gorm:"type:varchar(50);default:'text'" json:"type"` // text, image, file
	Status     string         `gorm:"type:varchar(50);default:'sent'" json:"status"` // sent, delivered, read
	ReadAt     *time.Time     `json:"readAt,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Sender   User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}
```

### Step 2: Create Migration

```bash
make migrate-create name=create_messages_table
```

`migrations/000003_create_messages_table.up.sql`:

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'text',
    status VARCHAR(50) DEFAULT 'sent',
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);

-- Composite index for conversation queries
CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, created_at DESC);
```

`migrations/000003_create_messages_table.down.sql`:

```sql
DROP TABLE IF EXISTS messages;
```

Run migration:

```bash
make migrate-up
```

### Step 3: Create Module Structure

```bash
mkdir -p internal/modules/messages/dto
touch internal/modules/messages/dto/request.go
touch internal/modules/messages/dto/response.go
touch internal/modules/messages/repository.go
touch internal/modules/messages/service.go
touch internal/modules/messages/handler.go
touch internal/modules/messages/routes.go
```

### Step 4: Create DTOs

`internal/modules/messages/dto/request.go`:

```go
package dto

import "errors"

type SendMessageRequest struct {
	ReceiverID string `json:"receiverId" binding:"required,uuid"`
	Content    string `json:"content" binding:"required,min=1,max=5000"`
	Type       string `json:"type" binding:"omitempty,oneof=text image file"`
}

func (r *SendMessageRequest) Validate() error {
	if r.ReceiverID == "" {
		return errors.New("receiverId is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if r.Type == "" {
		r.Type = "text"
	}
	return nil
}

type GetMessagesRequest struct {
	UserID string `form:"userId" binding:"required,uuid"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (r *GetMessagesRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 50
	}
}

type MarkAsReadRequest struct {
	MessageIDs []string `json:"messageIds" binding:"required,min=1"`
}
```

`internal/modules/messages/dto/response.go`:

```go
package dto

import (
	"time"
	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
)

type MessageResponse struct {
	ID         string                 `json:"id"`
	SenderID   string                 `json:"senderId"`
	ReceiverID string                 `json:"receiverId"`
	Content    string                 `json:"content"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	ReadAt     *time.Time             `json:"readAt,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	Sender     *authdto.UserResponse  `json:"sender,omitempty"`
	Receiver   *authdto.UserResponse  `json:"receiver,omitempty"`
}

type ConversationResponse struct {
	UserID       string                `json:"userId"`
	User         *authdto.UserResponse `json:"user"`
	LastMessage  *MessageResponse      `json:"lastMessage"`
	UnreadCount  int64                 `json:"unreadCount"`
	IsOnline     bool                  `json:"isOnline"`
}

func ToMessageResponse(msg *models.Message) *MessageResponse {
	resp := &MessageResponse{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Content:    msg.Content,
		Type:       msg.Type,
		Status:     msg.Status,
		ReadAt:     msg.ReadAt,
		CreatedAt:  msg.CreatedAt,
	}
	
	if msg.Sender.ID != "" {
		resp.Sender = authdto.ToUserResponse(&msg.Sender)
	}
	if msg.Receiver.ID != "" {
		resp.Receiver = authdto.ToUserResponse(&msg.Receiver)
	}
	
	return resp
}
```

### Step 5: Implement Repository

`internal/modules/messages/repository.go`:

```go
package messages

import (
	"context"
	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, message *models.Message) error
	FindByID(ctx context.Context, id string) (*models.Message, error)
	GetConversation(ctx context.Context, userID1, userID2 string, page, limit int) ([]*models.Message, int64, error)
	MarkAsRead(ctx context.Context, messageIDs []string, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
	GetConversations(ctx context.Context, userID string) ([]*models.Message, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Receiver").
		Where("id = ?", id).
		First(&message).Error
	return &message, err
}

func (r *repository) GetConversation(ctx context.Context, userID1, userID2 string, page, limit int) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID1, userID2, userID2, userID1)
	
	query.Count(&total)
	
	offset := (page - 1) * limit
	err := query.
		Preload("Sender").
		Preload("Receiver").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	
	return messages, total, err
}

func (r *repository) MarkAsRead(ctx context.Context, messageIDs []string, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id IN ? AND receiver_id = ? AND status != ?", messageIDs, userID, "read").
		Updates(map[string]interface{}{
			"status":  "read",
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *repository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_id = ? AND status != ?", userID, "read").
		Count(&count).Error
	return count, err
}

func (r *repository) GetConversations(ctx context.Context, userID string) ([]*models.Message, error) {
	var messages []*models.Message
	
	// Get the latest message for each conversation
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (
				CASE 
					WHEN sender_id = ? THEN receiver_id 
					ELSE sender_id 
				END
			) *
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
			ORDER BY 
				CASE 
					WHEN sender_id = ? THEN receiver_id 
					ELSE sender_id 
				END,
				created_at DESC
		`, userID, userID, userID, userID).
		Preload("Sender").
		Preload("Receiver").
		Find(&messages).Error
	
	return messages, err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error
}
```

### Step 6: Implement Service with WebSocket

`internal/modules/messages/service.go`:

```go
package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/messages/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	SendMessage(ctx context.Context, senderID string, req dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetConversation(ctx context.Context, userID, otherUserID string, page, limit int) ([]*dto.MessageResponse, int64, error)
	MarkAsRead(ctx context.Context, userID string, req dto.MarkAsReadRequest) error
	GetConversations(ctx context.Context, userID string) ([]*dto.ConversationResponse, error)
	DeleteMessage(ctx context.Context, userID, messageID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(ctx context.Context, senderID string, req dto.SendMessageRequest) (*dto.MessageResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}
	
	// Check if receiver is online
	isOnline, _ := helpers.IsUserOnline(req.ReceiverID)
	
	message := &models.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		Type:       req.Type,
		Status:     "sent",
	}
	
	if err := s.repo.Create(ctx, message); err != nil {
		logger.Error("failed to create message", "error", err)
		return nil, response.InternalServerError("Failed to send message", err)
	}
	
	// Fetch with relations
	message, _ = s.repo.FindByID(ctx, message.ID)
	
	// Send via WebSocket to receiver (if online)
	if isOnline {
		helpers.SendChatMessage(req.ReceiverID, dto.ToMessageResponse(message))
		
		// Update status to delivered
		message.Status = "delivered"
		s.repo.Update(ctx, message)
	}
	
	// Send confirmation to sender
	helpers.SendChatMessageSent(senderID, dto.ToMessageResponse(message))
	
	logger.Info("message sent",
		"messageID", message.ID,
		"senderID", senderID,
		"receiverID", req.ReceiverID,
		"receiverOnline", isOnline,
	)
	
	return dto.ToMessageResponse(message), nil
}

func (s *service) GetConversation(ctx context.Context, userID, otherUserID string, page, limit int) ([]*dto.MessageResponse, int64, error) {
	// Try cache first for recent conversations
	cacheKey := fmt.Sprintf("conversation:%s:%s:%d:%d", userID, otherUserID, page, limit)
	
	var cachedMessages []*models.Message
	err := cache.GetJSON(ctx, cacheKey, &cachedMessages)
	if err == nil && len(cachedMessages) > 0 {
		result := make([]*dto.MessageResponse, len(cachedMessages))
		for i, msg := range cachedMessages {
			result[i] = dto.ToMessageResponse(msg)
		}
		return result, int64(len(cachedMessages)), nil
	}
	
	messages, total, err := s.repo.GetConversation(ctx, userID, otherUserID, page, limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch messages", err)
	}
	
	// Cache for 1 minute
	if page == 1 {
		cache.SetJSON(ctx, cacheKey, messages, 1*time.Minute)
	}
	
	result := make([]*dto.MessageResponse, len(messages))
	for i, msg := range messages {
		result[i] = dto.ToMessageResponse(msg)
	}
	
	return result, total, nil
}

func (s *service) MarkAsRead(ctx context.Context, userID string, req dto.MarkAsReadRequest) error {
	if len(req.MessageIDs) == 0 {
		return response.BadRequest("No message IDs provided")
	}
	
	if err := s.repo.MarkAsRead(ctx, req.MessageIDs, userID); err != nil {
		return response.InternalServerError("Failed to mark messages as read", err)
	}
	
	// Get first message to find sender
	message, err := s.repo.FindByID(ctx, req.MessageIDs[0])
	if err == nil {
		// Send read receipt to sender via WebSocket
		helpers.SendReadReceipt(message.SenderID, userID, req.MessageIDs)
	}
	
	// Invalidate cache
	cachePattern := fmt.Sprintf("conversation:%s:*", userID)
	cache.Delete(ctx, cachePattern)
	
	logger.Info("messages marked as read",
		"userID", userID,
		"count", len(req.MessageIDs),
	)
	
	return nil
}

func (s *service) GetConversations(ctx context.Context, userID string) ([]*dto.ConversationResponse, error) {
	messages, err := s.repo.GetConversations(ctx, userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch conversations", err)
	}
	
	result := make([]*dto.ConversationResponse, len(messages))
	
	for i, msg := range messages {
		// Determine the other user
		var otherUser *models.User
		if msg.SenderID == userID {
			otherUser = &msg.Receiver
		} else {
			otherUser = &msg.Sender
		}
		
		// Get unread count for this conversation
		unreadCount := int64(0)
		// You can implement this by querying unread messages from this user
		
		// Check if user is online
		isOnline, _ := helpers.IsUserOnline(otherUser.ID)
		
		result[i] = &dto.ConversationResponse{
			UserID:      otherUser.ID,
			User:        authdto.ToUserResponse(otherUser),
			LastMessage: dto.ToMessageResponse(msg),
			UnreadCount: unreadCount,
			IsOnline:    isOnline,
		}
	}
	
	return result, nil
}

func (s *service) DeleteMessage(ctx context.Context, userID, messageID string) error {
	message, err := s.repo.FindByID(ctx, messageID)
	if err != nil {
		return response.NotFoundError("Message")
	}
	
	// Only sender can delete
	if message.SenderID != userID {
		return response.ForbiddenError("You can only delete your own messages")
	}
	
	if err := s.repo.Delete(ctx, messageID); err != nil {
		return response.InternalServerError("Failed to delete message", err)
	}
	
	// Notify via WebSocket
	helpers.SendMessageDeleted(message.SenderID, message.ReceiverID, messageID)
	
	logger.Info("message deleted", "messageID", messageID, "userID", userID)
	
	return nil
}
```

### Step 7: Implement Handler

`internal/modules/messages/handler.go`:

```go
package messages

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/messages/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SendMessage godoc
// @Summary Send a message
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SendMessageRequest true "Message data"
// @Success 201 {object} response.Response{data=dto.MessageResponse}
// @Router /messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	
	message, err := h.service.SendMessage(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, message, "Message sent successfully")
}

// GetConversation godoc
// @Summary Get conversation with a user
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param userId query string true "Other user ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]dto.MessageResponse}
// @Router /messages [get]
func (h *Handler) GetConversation(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var req dto.GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}
	req.SetDefaults()
	
	messages, total, err := h.service.GetConversation(
		c.Request.Context(),
		userID.(string),
		req.UserID,
		req.Page,
		req.Limit,
	)
	if err != nil {
		c.Error(err)
		return
	}
	
	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, messages, pagination, "Messages retrieved successfully")
}

// MarkAsRead godoc
// @Summary Mark messages as read
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MarkAsReadRequest true "Message IDs"
// @Success 200 {object} response.Response
// @Router /messages/read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var req dto.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	
	if err := h.service.MarkAsRead(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, nil, "Messages marked as read")
}

// GetConversations godoc
// @Summary Get all conversations
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.ConversationResponse}
// @Router /messages/conversations [get]
func (h *Handler) GetConversations(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	conversations, err := h.service.GetConversations(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, conversations, "Conversations retrieved successfully")
}

// DeleteMessage godoc
// @Summary Delete a message
// @Tags messages
// @Security BearerAuth
// @Param id path string true "Message ID"
// @Success 200 {object} response.Response
// @Router /messages/{id} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	userID, _ := c.Get("userID")
	messageID := c.Param("id")
	
	if err := h.service.DeleteMessage(c.Request.Context(), userID.(string), messageID); err != nil {
		c.Error(err)
		return
	}
	
	response.Success(c, nil, "Message deleted successfully")
}
```

### Step 8: Define Routes

`internal/modules/messages/routes.go`:

```go
package messages

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	messages := router.Group("/messages")
	messages.Use(authMiddleware)
	{
		messages.POST("", handler.SendMessage)
		messages.GET("", handler.GetConversation)
		messages.POST("/read", handler.MarkAsRead)
		messages.GET("/conversations", handler.GetConversations)
		messages.DELETE("/:id", handler.DeleteMessage)
	}
}
```

### Step 9: Register in main.go

```go
// Add import
import (
	"github.com/umar5678/go-backend/internal/modules/messages"
)

// In main() function:
// Messages module
messagesRepo := messages.NewRepository(db)
messagesService := messages.NewService(messagesRepo)
messagesHandler := messages.NewHandler(messagesService)
messages.RegisterRoutes(v1, messagesHandler, authMiddleware)
```

### Step 10: Update Swagger & Test

```bash
# Regenerate swagger
make swagger

# Test sending message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiverId": "RECEIVER_USER_ID",
    "content": "Hello from the API!",
    "type": "text"
  }'

# Get conversation
curl "http://localhost:8080/api/v1/messages?userId=OTHER_USER_ID&page=1&limit=50" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Mark as read
curl -X POST http://localhost:8080/api/v1/messages/read \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "messageIds": ["MESSAGE_ID_1", "MESSAGE_ID_2"]
  }'
```

---

## Common Patterns & Best Practices

### Pattern 1: Module Structure

```
module/
├── dto/
│   ├── request.go      # Input validation
│   └── response.go     # Output formatting
├── handler.go          # HTTP layer
├── repository.go       # Data layer
├── service.go          # Business logic
└── routes.go           # Route definitions
```

### Pattern 2: Error Handling

```go
// In Service
if err != nil {
	logger.Error("operation failed", "error", err, "context", "value")
	return nil, response.InternalServerError("User-friendly message", err)
}

// In Handler
if err := h.service.DoSomething(ctx); err != nil {
	c.Error(err)  // Middleware handles the response
	return
}
```

### Pattern 3: Caching Strategy

```go
// 1. Check cache
cacheKey := fmt.Sprintf("resource:%s", id)
var data YourStruct
err := cache.GetJSON(ctx, cacheKey, &data)
if err == nil {
	return &data, nil  // Cache hit
}

// 2. Get from database
result, err := s.repo.FindByID(ctx, id)
if err != nil {
	return nil, err
}

// 3. Store in cache
cache.SetJSON(ctx, cacheKey, result, 5*time.Minute)

return result, nil
```

### Pattern 4: WebSocket Notifications

```go
// After database operation
if err := s.repo.Create(ctx, entity); err != nil {
	return err
}

// Send real-time notification
go func() {
	helpers.SendNotification(userID, map[string]interface{}{
		"type": "entity_created",
		"data": entity,
	})
}()
```

### Pattern 5: Pagination

```go
// In Handler
var req dto.ListRequest
if err := c.ShouldBindQuery(&req); err != nil {
	c.Error(response.BadRequest("Invalid parameters"))
	return
}
req.SetDefaults()

items, total, err := h.service.List(ctx, req)
pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
response.Paginated(c, items, pagination, "Success")
```

---

## Checklist for Adding New Features

### Planning Phase

- [ ] Define clear requirements
- [ ] Identify data models needed
- [ ] Determine if Redis caching is needed
- [ ] Determine if WebSocket is needed
- [ ] Plan API endpoints
- [ ] Design database schema

### Implementation Phase

- [ ] Create model in `internal/models/`
- [ ] Create and run migration
- [ ] Create module directory structure
- [ ] Implement DTOs (request/response)
- [ ] Implement repository (database operations)
- [ ] Implement service (business logic)
- [ ] Implement handler (HTTP layer)
- [ ] Define routes
- [ ] Register module in main.go
- [ ] Add caching if needed
- [ ] Add WebSocket helpers if needed
- [ ] Add logging
- [ ] Handle errors properly

### Testing Phase

- [ ] Test with Postman/curl
- [ ] Test WebSocket connections
- [ ] Test Redis caching
- [ ] Test error scenarios
- [ ] Test validation
- [ ] Update Swagger docs

### Documentation Phase

- [ ] Add API documentation comments
- [ ] Update README if needed
- [ ] Document WebSocket messages
- [ ] Add usage examples

---

## Advanced Patterns

### Pattern 6: Background Jobs with Redis

For tasks that don't need immediate completion:

```go
// In service
func (s *service) ProcessLargeTask(ctx context.Context, data interface{}) error {
	// Store task in Redis queue
	taskID := uuid.New().String()
	taskData := map[string]interface{}{
		"id":     taskID,
		"type":   "large_task",
		"data":   data,
		"status": "pending",
	}
	
	cache.SetJSON(ctx, fmt.Sprintf("task:%s", taskID), taskData, 1*time.Hour)
	cache.Set(ctx, "task:queue:large_task", taskID, 1*time.Hour)
	
	// Process asynchronously
	go s.processTaskAsync(taskID, data)
	
	return nil
}

func (s *service) processTaskAsync(taskID string, data interface{}) {
	ctx := context.Background()
	
	// Update status
	cache.Set(ctx, fmt.Sprintf("task:%s:status", taskID), "processing", 1*time.Hour)
	
	// Do the work
	result, err := s.doHeavyWork(data)
	
	// Update status
	if err != nil {
		cache.Set(ctx, fmt.Sprintf("task:%s:status", taskID), "failed", 1*time.Hour)
		cache.Set(ctx, fmt.Sprintf("task:%s:error", taskID), err.Error(), 1*time.Hour)
	} else {
		cache.Set(ctx, fmt.Sprintf("task:%s:status", taskID), "completed", 1*time.Hour)
		cache.SetJSON(ctx, fmt.Sprintf("task:%s:result", taskID), result, 1*time.Hour)
	}
	
	// Notify user via WebSocket
	helpers.SendNotification(userID, map[string]interface{}{
		"type":   "task_completed",
		"taskId": taskID,
		"status": "completed",
	})
}
```

### Pattern 7: Rate Limiting Per User

```go
// In service
func (s *service) CheckRateLimit(ctx context.Context, userID string, action string) error {
	key := fmt.Sprintf("ratelimit:%s:%s", userID, action)
	
	count, err := cache.IncrementWithExpiry(ctx, key, 1*time.Minute)
	if err != nil {
		return err
	}
	
	if count > 10 { // 10 requests per minute
		return response.TooManyRequests("Rate limit exceeded")
	}
	
	return nil
}

// In handler
func (h *Handler) CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	// Check rate limit
	if err := h.service.CheckRateLimit(c.Request.Context(), userID.(string), "create_post"); err != nil {
		c.Error(err)
		return
	}
	
	// Continue with normal logic...
}
```

### Pattern 8: Soft Delete with Cache Invalidation

```go
// In repository
func (r *repository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.YourModel{}, "id = ?", id).Error
}

// In service
func (s *service) Delete(ctx context.Context, id string) error {
	// Get entity first
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return response.NotFoundError("Entity")
	}
	
	// Soft delete
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return response.InternalServerError("Failed to delete", err)
	}
	
	// Invalidate all related caches
	cache.Delete(ctx, fmt.Sprintf("entity:%s", id))
	cache.Delete(ctx, fmt.Sprintf("user:entities:%s", entity.UserID))
	
	// Notify via WebSocket
	helpers.SendNotification(entity.UserID, map[string]interface{}{
		"type":     "entity_deleted",
		"entityId": id,
	})
	
	return nil
}
```

### Pattern 9: Optimistic Locking

```go
// Add version field to model
type Post struct {
	// ... other fields
	Version int `gorm:"not null;default:1" json:"version"`
}

// In repository
func (r *repository) UpdateWithVersion(ctx context.Context, post *models.Post, expectedVersion int) error {
	result := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ? AND version = ?", post.ID, expectedVersion).
		Updates(map[string]interface{}{
			"title":   post.Title,
			"content": post.Content,
			"version": gorm.Expr("version + 1"),
		})
	
	if result.RowsAffected == 0 {
		return errors.New("concurrent modification detected")
	}
	
	return result.Error
}

// In service
func (s *service) Update(ctx context.Context, id string, req dto.UpdateRequest) error {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return response.NotFoundError("Post")
	}
	
	// Update fields
	post.Title = req.Title
	post.Content = req.Content
	
	// Try to update with version check
	err = s.repo.UpdateWithVersion(ctx, post, post.Version)
	if err != nil {
		if err.Error() == "concurrent modification detected" {
			return response.ConflictError("Post was modified by another user. Please refresh and try again.")
		}
		return response.InternalServerError("Failed to update", err)
	}
	
	return nil
}
```

### Pattern 10: Bulk Operations with Transactions

```go
// In service
func (s *service) BulkCreate(ctx context.Context, items []dto.CreateRequest) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	for _, item := range items {
		entity := &models.Entity{
			Title:   item.Title,
			Content: item.Content,
		}
		
		if err := tx.Create(entity).Error; err != nil {
			tx.Rollback()
			return response.InternalServerError("Failed to create items", err)
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}
	
	// Invalidate cache after successful bulk insert
	cache.Delete(ctx, "entities:list:*")
	
	return nil
}
```

---

## WebSocket Integration Patterns

### Pattern 11: Real-Time Collaboration

```go
// For features like collaborative editing

// In service
func (s *service) UpdateDocument(ctx context.Context, docID, userID string, changes map[string]interface{}) error {
	// Update document
	if err := s.repo.UpdateDocument(ctx, docID, changes); err != nil {
		return err
	}
	
	// Get all collaborators
	collaborators, _ := s.repo.GetCollaborators(ctx, docID)
	
	// Notify all collaborators except the editor
	for _, collaboratorID := range collaborators {
		if collaboratorID != userID {
			helpers.SendNotification(collaboratorID, map[string]interface{}{
				"type":       "document_updated",
				"documentId": docID,
				"changes":    changes,
				"updatedBy":  userID,
			})
		}
	}
	
	return nil
}
```

### Pattern 12: Presence Tracking

```go
// Track who's viewing/editing what

// In service
func (s *service) JoinRoom(ctx context.Context, userID, roomID string) error {
	// Store in Redis
	key := fmt.Sprintf("room:%s:users", roomID)
	cache.SessionClient.SAdd(ctx, key, userID)
	cache.SessionClient.Expire(ctx, key, 5*time.Minute)
	
	// Get all users in room
	users, _ := cache.SessionClient.SMembers(ctx, key).Result()
	
	// Notify all users
	for _, uid := range users {
		if uid != userID {
			helpers.SendNotification(uid, map[string]interface{}{
				"type":   "user_joined_room",
				"roomId": roomID,
				"userId": userID,
			})
		}
	}
	
	return nil
}

func (s *service) LeaveRoom(ctx context.Context, userID, roomID string) error {
	key := fmt.Sprintf("room:%s:users", roomID)
	cache.SessionClient.SRem(ctx, key, userID)
	
	// Notify remaining users
	users, _ := cache.SessionClient.SMembers(ctx, key).Result()
	for _, uid := range users {
		helpers.SendNotification(uid, map[string]interface{}{
			"type":   "user_left_room",
			"roomId": roomID,
			"userId": userID,
		})
	}
	
	return nil
}
```

### Pattern 13: Notification Preferences

```go
// Allow users to control what notifications they receive

// Model
type NotificationPreference struct {
	UserID           string `gorm:"type:uuid;primaryKey"`
	EmailEnabled     bool   `gorm:"default:true"`
	PushEnabled      bool   `gorm:"default:true"`
	WebSocketEnabled bool   `gorm:"default:true"`
	Categories       string `gorm:"type:jsonb"` // {"posts":true,"messages":true}
}

// In notification service
func (s *service) SendNotification(ctx context.Context, userID, category string, notification interface{}) error {
	// Get user preferences
	prefs, _ := s.repo.GetPreferences(ctx, userID)
	
	// Check if category is enabled
	if !s.isCategoryEnabled(prefs, category) {
		return nil
	}
	
	// Send via appropriate channels
	if prefs.WebSocketEnabled {
		helpers.SendNotification(userID, notification)
	}
	
	if prefs.EmailEnabled {
		// Send email
		s.emailService.Send(ctx, userID, notification)
	}
	
	if prefs.PushEnabled {
		// Send push notification
		s.pushService.Send(ctx, userID, notification)
	}
	
	return nil
}
```

---

## Testing Guide

### Unit Testing

Create `internal/modules/posts/service_test.go`:

```go
package posts

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*models.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Post), args.Error(1)
}

// Test
func TestCreatePost(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	
	ctx := context.Background()
	req := dto.CreatePostRequest{
		Title:   "Test Post",
		Content: "Test content",
		Status:  "draft",
	}
	
	expectedPost := &models.Post{
		ID:      "123",
		Title:   req.Title,
		Content: req.Content,
	}
	
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)
	mockRepo.On("FindByID", ctx, "123").Return(expectedPost, nil)
	
	result, err := service.CreatePost(ctx, "user-id", req)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Title, result.Title)
	mockRepo.AssertExpectations(t)
}
```

### Integration Testing

Create `test/integration/posts_test.go`:

```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPostsAPI(t *testing.T) {
	// Setup test server
	router := setupTestRouter()
	
	// Test create post
	t.Run("Create Post", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":   "Integration Test Post",
			"content": "This is a test post",
			"status":  "draft",
		}
		
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testToken)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])
	})
}
```

---

## Performance Optimization

### 1. Database Indexing

```sql
-- Add indexes for common queries
CREATE INDEX idx_posts_author_status ON posts(author_id, status);
CREATE INDEX idx_posts_created_at_desc ON posts(created_at DESC);
CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, created_at DESC);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM posts WHERE author_id = 'xxx' AND status = 'published';
```

### 2. Query Optimization

```go
// BAD: N+1 query problem
func (r *repository) List(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	r.db.Find(&posts)
	
	// This will cause N queries (one per post)
	for _, post := range posts {
		r.db.Model(&post).Association("Author").Find(&post.Author)
	}
	return posts, nil
}

// GOOD: Use Preload
func (r *repository) List(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.
		Preload("Author").
		Preload("Comments").
		Find(&posts).Error
	return posts, err
}
```

### 3. Redis Caching Strategy

```go
// Cache hot data with short TTL
func (s *service) GetPopularPosts(ctx context.Context) ([]*dto.PostResponse, error) {
	cacheKey := "posts:popular"
	
	var cached []*models.Post
	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		// Convert and return cached data
		return s.convertToResponse(cached), nil
	}
	
	// Get from database
	posts, err := s.repo.GetPopular(ctx, 10)
	if err != nil {
		return nil, err
	}
	
	// Cache for 5 minutes
	cache.SetJSON(ctx, cacheKey, posts, 5*time.Minute)
	
	return s.convertToResponse(posts), nil
}
```

### 4. Connection Pooling

Already configured in your setup:

```go
// In database connection
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

---

## Troubleshooting Guide

### Issue 1: WebSocket Not Connecting

**Symptoms**: WebSocket connection fails with 401 or connection refused

**Solutions**:

```bash
# 1. Check if server is running
curl http://localhost:8080/health

# 2. Verify token is valid
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. Test WebSocket with correct token
websocat ws://localhost:8080/ws?token=VALID_TOKEN

# 4. Check Redis is running
make redis-cli
# In redis-cli: PING (should return PONG)

# 5. Check logs
tail -f logs/app.log
```

### Issue 2: Messages Not Being Received

**Symptoms**: Messages sent but not received via WebSocket

**Solutions**:

```go
// 1. Check if user is connected
isOnline, _ := helpers.IsUserOnline(userID)
fmt.Println("User online:", isOnline)

// 2. Check Redis Pub/Sub
// In redis-cli:
// SUBSCRIBE websocket:broadcast

// 3. Enable debug logging
// In .env: LOG_LEVEL=debug

// 4. Verify hub is running
// Check logs for "websocket hub started"
```

### Issue 3: Cache Not Working

**Symptoms**: Cache hits not happening, always fetching from DB

**Solutions**:

```go
// 1. Verify Redis connection
ctx := context.Background()
err := cache.Set(ctx, "test", "value", 1*time.Minute)
if err != nil {
	log.Fatal("Redis not working:", err)
}

result, err := cache.Get(ctx, "test")
fmt.Println("Cache test:", result, err)

// 2. Check cache keys
// In redis-cli:
// KEYS *

// 3. Check TTL
// In redis-cli:
// TTL your:cache:key

// 4. Monitor Redis operations
make redis-monitor
```

### Issue 4: Database Migration Fails

**Symptoms**: Migration error when running `make migrate-up`

**Solutions**:

```bash
# 1. Check database connection
psql -h localhost -U go_backend_admin -d go_backend

# 2. Check current migration version
migrate -path migrations -database "$DB_URL" version

# 3. Force to specific version if stuck
make migrate-force version=1

# 4. Rollback and retry
make migrate-down
make migrate-up

# 5. Check migration files for syntax errors
cat migrations/000002_create_posts_table.up.sql
```

### Issue 5: Rate Limiting Too Aggressive

**Symptoms**: Getting 429 errors frequently

**Solutions**:

```bash
# 1. Update .env
RATE_LIMIT_REQUESTS_PER_SECOND=100
RATE_LIMIT_BURST=200

# 2. Restart server
make dev

# 3. Check rate limit in Redis
# In redis-cli:
# KEYS ratelimit:*
# GET ratelimit:your-ip

# 4. Clear rate limit for testing
# In redis-cli:
# DEL ratelimit:your-ip
```

---

## Environment Variables Reference

### Complete .env File

```env
# Application
APP_NAME=go-backend
APP_ENV=development
APP_VERSION=1.0.0
APP_DEBUG=true

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30
SERVER_SHUTDOWN_TIMEOUT=5

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=go_backend_admin
DB_PASSWORD=goPass
DB_NAME=go_backend
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_MAX_LIFETIME=5
DB_LOG_LEVEL=1

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_POOL_SIZE=10
REDIS_MAIN_DB=0
REDIS_CACHE_DB=3
REDIS_SESSION_DB=4
REDIS_PUBSUB_DB=1

# JWT
JWT_SECRET=your-secret-key-change-this-in-production
JWT_ACCESS_EXPIRY=15
JWT_REFRESH_EXPIRY=168
JWT_ISSUER=your-project

# Logger
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=./logs/app.log

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization
CORS_ALLOW_CREDENTIALS=true

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_SECOND=100
RATE_LIMIT_BURST=200
```

---

## Quick Command Reference

```bash
# Development
make dev                    # Start with hot reload
make run                    # Start without hot reload
make build                  # Build binary

# Database
make migrate-create name=X  # Create migration
make migrate-up             # Run migrations
make migrate-down           # Rollback migrations
make migrate-force version=N # Force version

# Redis
make redis-start            # Start Redis
make redis-stop             # Stop Redis
make redis-cli              # Connect to Redis CLI
make redis-monitor          # Monitor Redis commands

# Documentation
make swagger                # Generate Swagger docs

# Testing
make test                   # Run tests
make test-integration       # Run integration tests

# WebSocket
make ws-health              # Check WebSocket health

# Cleanup
make clean                  # Clean build artifacts
make clean-all              # Clean everything including Redis
```

---

## Summary Checklist

### For Every New Feature:

1. **Planning** ✅
    
    - Define requirements clearly
    - Design database schema
    - Plan API endpoints
    - Decide on caching strategy
    - Decide on WebSocket needs
2. **Database** ✅
    
    - Create model
    - Create migration
    - Run migration
    - Add indexes
3. **Module Implementation** ✅
    
    - Create module structure
    - Implement DTOs
    - Implement repository
    - Implement service
    - Implement handler
    - Define routes
4. **Integration** ✅
    
    - Register in main.go
    - Add caching where needed
    - Add WebSocket helpers where needed
    - Add proper logging
    - Handle all errors
5. **Testing** ✅
    
    - Test with Postman/curl
    - Test error scenarios
    - Test WebSocket if applicable
    - Update Swagger docs
6. **Documentation** ✅
    
    - Add code comments
    - Document API endpoints
    - Add usage examples

---

## Need Help?

### Common Commands to Debug

```bash
# Check if everything is running
make ps

# View logs
make logs

# Check health
curl http://localhost:8080/health | jq .

# Test Redis
make redis-cli
# Then: PING

# Check WebSocket connections
curl http://localhost:8080/health | jq .websocket

# Monitor Redis operations
make redis-monitor
```

### Resources

- **Swagger Docs**: http://localhost:8080/docs/index.html
- **Health Check**: http://localhost:8080/health
- **Ready Check**: http://localhost:8080/ready

---

**You now have everything you need to build production-grade features with Redis caching and WebSocket real-time capabilities! 🚀**