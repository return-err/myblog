package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/return-err/myblog/internal/config"
	"github.com/return-err/myblog/internal/handle"
	"github.com/return-err/myblog/internal/middleware"
	"github.com/return-err/myblog/internal/repository"
	"github.com/return-err/myblog/internal/router"
	"github.com/return-err/myblog/internal/service"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("startup validation failed", "err", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogFormat)

	db, err := repository.InitAPP(cfg)
	if err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	jwtService := service.NewJWTService(cfg.JWTSecret, cfg.AccessTTL)

	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, jwtService, cfg.RefreshTTL)
	authHandle := handle.NewAuthHandle(authService)
	mw := middleware.NewMiddleware(jwtService, authRepo)

	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo)
	postHandle := handle.NewPostHandle(postService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandle := handle.NewUserHandle(userService)

	healthHandle := handle.NewHealthHandle(postRepo)
	r := router.SetupRouter(authHandle, postHandle, mw, userHandle, healthHandle)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced server shutdown", "err", err)
	}
	slog.Info("Exited")
}

func setupLogger(format string) {
	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}

	slog.SetDefault(slog.New(handler))
}
