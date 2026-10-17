package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/handle"
	"github.com/return-err/myblog/internal/middleware"
	"golang.org/x/time/rate"
)

func SetupRouter(authHandle *handle.AuthHandle, postHandle *handle.PostHandle, mw *middleware.Middleware, userHandle *handle.UserHandle) *gin.Engine {
	r := gin.Default()

	loginLimiter := middleware.NewIPLimiter(rate.Every(10*time.Second), 10)

	r.Static("/static", "./static")

	r.GET("/", func(ctx *gin.Context) {
		ctx.File("templates/index.html")
	})

	r.GET("/register", func(ctx *gin.Context) {
		ctx.File("templates/register.html")
	})

	r.GET("/login", func(ctx *gin.Context) {
		ctx.File("templates/login.html")
	})

	r.GET("/create", func(ctx *gin.Context) {
		ctx.File("templates/post.html")
	})

	r.GET("/users/:id", func(ctx *gin.Context) {
		ctx.File("templates/user_profile.html")
	})

	r.GET("/posts/:id", func(ctx *gin.Context) {
		ctx.File("templates/post_detail.html")
	})

	r.GET("/posts/:id/edit", func(ctx *gin.Context) {
		ctx.File("templates/edit.html")
	})

	r.GET("/admin/users", func(ctx *gin.Context) {
		ctx.File("templates/admin_users.html")
	})

	api := r.Group("/api")
	{
		api.GET("/posts", postHandle.PostsList)
		api.GET("/posts/:id", postHandle.PostDetail)
		api.GET("/posts/search", postHandle.SearchPost)
		api.GET("/users/:id/posts", postHandle.GetUserPosts)

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandle.Register)
			auth.POST("/login", middleware.LoginRateLimit(loginLimiter), authHandle.Login)
			auth.POST("/refresh", authHandle.Refresh)
		}

		opt := api.Group("/opt")
		opt.Use(mw.OptionalAuthMiddleware())
		{
			opt.GET("/users/:id", userHandle.GetUserProfile)
		}

		protected := api.Group("")
		protected.Use(mw.AuthMiddleware())
		{
			protected.POST("/posts", postHandle.CreatePost)
			protected.PUT("/posts/:id", postHandle.EditPost)
			protected.DELETE("/posts/:id", postHandle.DeletePost)
			protected.GET("/me", userHandle.MyPage)
			protected.PUT("/me/profile", userHandle.UpdateProfile)
			protected.POST("/auth/logout", authHandle.Logout)
		}

		admin := api.Group("/admin")
		admin.Use(mw.AuthMiddleware())
		admin.Use(mw.RequireRole("admin"))
		{
			admin.GET("/users", userHandle.GetAllUsers)
			admin.PUT("/users/:id/role", userHandle.UpdateRole)
		}
	}

	return r
}
