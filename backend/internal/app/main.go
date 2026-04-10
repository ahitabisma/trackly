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
	"trackly-backend/internal/investor"
	"trackly-backend/internal/shareholding"
	"trackly-backend/internal/user"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/jobs"
	"trackly-backend/pkg/logger"
	"trackly-backend/pkg/middleware"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env)

	connLifeTime := time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute
	db, err := database.NewDatabase(
		cfg.Database.Host,
		cfg.Database.Database,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		connLifeTime,
	)
	if err != nil {
		return err
	}

	// Redis client
	redisClient := database.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return err
	}

	// Job queue and worker
	jobQueue := jobs.NewRedisQueue(redisClient)
	jobWorker := jobs.NewWorker(jobQueue, 10, log)

	// Register job handlers
	jobWorker.RegisterHandler(jobs.JobTypeShareholdingImport, shareholding.NewShareholdingImportJobHandler(
		shareholding.NewShareHoldingService(company.NewCompanyRepository(db), investor.NewInvestorRepository(db), shareholding.NewShareholdingRepository(db), log),
		log,
	))

	// router
	mux := http.NewServeMux()

	// repos
	userRepo := user.NewUserRepository(db)
	companyRepo := company.NewCompanyRepository(db)
	investorRepo := investor.NewInvestorRepository(db)
	shareHoldingRepo := shareholding.NewShareholdingRepository(db)

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
	shareHoldingService := shareholding.NewShareHoldingService(companyRepo, investorRepo, shareHoldingRepo, log)

	// handlers
	appHandler := NewAppHandler(log, cfg)
	authHandler := auth.NewAuthHandler(authService, log)
	userHandler := user.NewUserHandler(userService, log)
	companyHandler := company.NewCompanyHandler(companyService, log)
	shareHoldingHandler := shareholding.NewShareholdingHandlerWithQueue(shareHoldingService, jobQueue, log)

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
	SetupAppRoutes(mux, appHandler)
	auth.SetupAuthRoutes(mux, authHandler)
	user.SetupUserRoutes(mux, userHandler, authMiddleware, adminMiddleware)
	company.SetupCompanyRoutes(mux, companyHandler, authMiddleware, adminMiddleware)
	shareholding.SetupShareholdingRoutes(mux, shareHoldingHandler, adminMiddleware)

	// Apply CORS middleware
	allowedOrigins := []string{
		cfg.App.FrontendURL,
		"http://localhost:3000",
		"http://localhost:3001",
	}
	corsMiddleware := middleware.CORSMiddleware(allowedOrigins)
	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", cfg.App.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Info("server running on " + addr)

	// Start job worker
	ctx, cancel := context.WithCancel(context.Background())
	jobWorker.Start(ctx)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Stop job worker
	cancel()
	jobWorker.Stop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.WithError(err).Error("error closing redis connection")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	return server.Shutdown(shutdownCtx)
}
