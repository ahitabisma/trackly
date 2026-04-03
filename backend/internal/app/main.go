package app

import (
	"fmt"
	"net/http"

	"trackly-backend/internal/handler"
	"trackly-backend/internal/repository"
	"trackly-backend/internal/routes"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/logger"
)

func Run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize logger
	log := logger.New(cfg.App.Env)

	// Initialize database
	db, err := database.NewPostgres(
		cfg.Supabase.Host,
		cfg.Supabase.Port,
		cfg.Supabase.Database,
		cfg.Supabase.User,
		cfg.Supabase.Password,
	)
	if err != nil {
		return err
	}
	defer db.Close()

	// ── Initialize dependencies ───────────────────────────────────
	companyRepo := repository.NewCompanyRepository(db)
	companyService := service.NewCompanyService(companyRepo, log)
	companyHandler := handler.NewCompanyHandler(companyService)

	// ── Setup HTTP server ─────────────────────────────────────────
	mux := http.NewServeMux()

	// Setup all routes
	routes.Setup(mux, companyHandler)

	// Start the server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("server running on " + addr)

	return http.ListenAndServe(addr, mux)
}
