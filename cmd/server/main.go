package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/return-err/myblog/internal/handle"
	"github.com/return-err/myblog/internal/middleware"
	"github.com/return-err/myblog/internal/repository"
	"github.com/return-err/myblog/internal/router"
	"github.com/return-err/myblog/internal/service"
)

func main() {
	godotenv.Load()

	secret := os.Getenv("JWT_SECRET")

	db, err := repository.InitAPP()
	if err != nil {
		log.Fatalln("データベース初期化失敗:", err)
	}
	defer db.Close()

	jwtService := service.NewJWTService(secret)

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, jwtService)
	authHandle := handle.NewAuthHandle(authService)
	middleware := middleware.NewMiddleware(jwtService, authRepo)

	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo)
	postHandle := handle.NewPostHandle(postService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandle := handle.NewUserHandle(userService)

	r := router.SetupRouter(authHandle, postHandle, middleware, userHandle)

	r.Run(":8080")
}
