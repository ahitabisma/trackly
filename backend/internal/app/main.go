package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"trackly-backend/internal/analisis"
	"trackly-backend/internal/auth"
	"trackly-backend/internal/company"
	"trackly-backend/internal/investor"
	"trackly-backend/internal/screening"
	"trackly-backend/internal/shareholding"
	"trackly-backend/internal/trading"
	"trackly-backend/internal/user"
	"trackly-backend/pkg/appsscript"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/jobs"
	"trackly-backend/pkg/logger"
	"trackly-backend/pkg/middleware"

	"github.com/sirupsen/logrus"
)

func startNightlyScreening(svc *screening.Service, log *logrus.Logger) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.WithError(err).Warn("failed to load Asia/Jakarta tz, using UTC")
		loc = time.UTC
	}

	go func() {
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, loc)
			if !now.Before(next) {
				next = next.Add(24 * time.Hour)
			}
			wait := next.Sub(now)
			log.WithField("next_run", next.Format("2006-01-02 15:04:05 MST")).Info("nightly screening scheduled")

			select {
			case <-time.After(wait):
				log.Info("nightly screening triggered")
				if err := svc.RunNightlyScreening(context.Background()); err != nil {
					log.WithError(err).Error("nightly screening failed")
				}
			}
		}
	}()
}

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

	// analisis service + handler
	appsscriptClient := appsscript.NewClient(cfg.AppsScript.URL, time.Duration(cfg.AppsScript.Timeout)*time.Second)
	workerPool := analisis.NewWorkerPool(log)
	analisisService := analisis.NewAnalisisService(
		appsscriptClient,
		log,
		workerPool,
		cfg.AppsScript.PythonScriptPath,
		time.Duration(cfg.AppsScript.PollIntervalMs)*time.Millisecond,
		cfg.AppsScript.PollMaxRetries,
	)
	analisisHandler := analisis.NewAnalisisHandler(analisisService, log, cfg)

	// trading service + handler
	tradingService := trading.NewTradingService(db, analisisService, log, cfg.AppsScript.PythonScriptPath, cfg.NvidiaApiKey, cfg.GeminiApiKey)
	tradingHandler := trading.NewTradingHandler(tradingService, log)

	// routes
	SetupAppRoutes(mux, appHandler)
	auth.SetupAuthRoutes(mux, authHandler)
	user.SetupUserRoutes(mux, userHandler, authMiddleware, adminMiddleware)
	company.SetupCompanyRoutes(mux, companyHandler, authMiddleware, adminMiddleware)
	shareholding.SetupShareholdingRoutes(mux, shareHoldingHandler, adminMiddleware)
	analisis.SetupAnalisisRoutes(mux, analisisHandler, authMiddleware)
	trading.SetupTradingRoutes(mux, tradingHandler, authMiddleware)

	// screening service + handler
	scriptDir := filepath.Dir(cfg.AppsScript.PythonScriptPath)
	screeningRepo := screening.NewRepository(db)
	screeningService := screening.NewService(analisisService, screeningRepo, log, scriptDir, cfg.NvidiaApiKey, cfg.GeminiApiKey)
	screeningHandler := screening.NewHandler(screeningService)
	screening.SetupScreeningRoutes(mux, screeningHandler, authMiddleware)

	// start nightly screening scheduler
	startNightlyScreening(screeningService, log)

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
