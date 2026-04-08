package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trackly-backend/internal/auth"
	"trackly-backend/internal/company"
	"trackly-backend/internal/user"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/logger"
	"trackly-backend/pkg/middleware"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env)

	db, err := database.NewDatabase(
		cfg.Database.Host,
		cfg.Database.Database,
		cfg.Database.User,
		cfg.Database.Password,
	)
	if err != nil {
		return err
	}

	// router
	mux := http.NewServeMux()

	// repos
	userRepo := user.NewUserRepository(db)
	companyRepo := company.NewCompanyRepository(db)

	// jwt
	jwtService := auth.NewJwtService(
		cfg.Jwt.Secret,
		cfg.App.Name,
		cfg.Jwt.Audience,
	)

	// services
	authService := auth.NewAuthService(userRepo, jwtService)
	userService := user.NewUserService(userRepo)
	companyService := company.NewCompanyService(companyRepo, log)

	// handlers
	authHandler := auth.NewAuthHandler(authService, log)
	userHandler := user.NewUserHandler(userService, log)
	companyHandler := company.NewCompanyHandler(companyService, log)

	// middleware
	authMiddleware := middleware.AuthMiddleware(jwtService, log)

	// Admin middleware chain
	adminMiddleware := func(next http.Handler) http.Handler {
		return middleware.ChainMiddleware(
			authMiddleware,
			middleware.RoleMiddleware("admin")(log),
		)(next)
	}

	// routes
	auth.SetupAuthRoutes(mux, authHandler)
	user.SetupUserRoutes(mux, userHandler, authMiddleware, adminMiddleware)
	company.SetupCompanyRoutes(mux, companyHandler, authMiddleware, adminMiddleware)

	addr := fmt.Sprintf(":%d", cfg.App.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Info("server running on " + addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
