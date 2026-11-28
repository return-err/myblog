package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/return-err/myblog/internal/handle"
	"github.com/return-err/myblog/internal/middleware"
)

func SetupRouter(authHandle *handle.AuthHandle, postHandle *handle.PostHandle, mw *middleware.Middleware, userHandle *handle.UserHandle, healthHandle *handle.HealthHandle, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	if err := r.SetTrustedProxies([]string{"172.18.0.0/16"}); err != nil {
		slog.Error("failed to set trusted proxies", "error", err)
	}

	loginLimiter := middleware.NewRedisLimiter(rdb, "ratelimit:login:", 5, time.Minute)

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

	r.GET("/healthz", healthHandle.Healthz)
	r.GET("/readyz", healthHandle.Readyz)

	// test ip router
	// r.GET("/debug/ip", func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{
	// 		"remote_addr": ctx.Request.RemoteAddr,
	// 		"xff":         ctx.GetHeader("X-Forwarded-FOR"),
	// 		"client_ip":   ctx.ClientIP(),
	// 		"proto":       ctx.GetHeader("X-Forwarded-Proto"),
	// 	})
	// })

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
